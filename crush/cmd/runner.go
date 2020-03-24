package cmd

import (
	"time"

	"fmt"
	"log"
	"os"

	"sync"

	metalgo "github.com/metal-stack/metal-go"
	"github.com/metal-stack/metal-go/api/models"
)

var (
	readyState = map[string]string{
		"t1-small-x86":  booting,
		"c1-medium-x86": booting,
		"default":       phonedHome,
	}
)

const (
	booting    = "Booting New Kernel"
	phonedHome = "Phoned Home"
	waiting    = "Waiting"
)

type machineResults struct {
	errors    map[*models.V1MachineResponse]error
	durations map[*models.V1MachineResponse]time.Duration
}

type runner struct {
	driver      *metalgo.Driver
	maxWaitTime time.Duration
	partitions  []string
	sizes       []string
	images      []string
}

func NewRunner(url, bearer, hmac string) *runner {
	driver, err := metalgo.NewDriver(url, bearer, hmac)
	if err != nil {
		fmt.Print(err)
		os.Exit(1)
	}
	return &runner{
		driver:      driver,
		maxWaitTime: 30 * time.Minute,
		partitions:  []string{"fra-equ01"},
		sizes: []string{
			"t1-small-x86",
			"c1-medium-x86",
			// "c1-large-x86",
			// "c1-xlarge-x86",
			// "s1-large-x86",
		},
		images: []string{
			"ubuntu-19.04",
			// "ubuntu-18.04",
			// "firewall-1",
		},
	}
}

func (r *runner) Run() {
	resultLen := len(r.partitions) * len(r.sizes) * len(r.images)
	results := make(chan machineResults)

	var wg sync.WaitGroup
	wg.Add(resultLen)

	for _, partition := range r.partitions {
		for _, size := range r.sizes {
			for _, image := range r.images {

				go func(partition, size, image string) {
					defer wg.Done()
					mr := machineResults{
						errors:    make(map[*models.V1MachineResponse]error),
						durations: make(map[*models.V1MachineResponse]time.Duration),
					}
					start := time.Now()

					machine, err := r.RunMachine(partition, size, image, "crush", "crush")
					if err != nil {
						mr.errors[machine] = err
						if machine != nil {
							_, err := r.driver.MachineDelete(*machine.ID)
							if err != nil {
								logMachine(machine, fmt.Sprintf("machineDelete completed with error:%v", err))
							}
						}
						logMachine(machine, fmt.Sprintf("cycle completed with error:%v", err))
					}
					duration := time.Since(start)
					mr.durations[machine] = duration
					logMachine(machine, fmt.Sprintf("cycle completed after:%s", duration))
					results <- mr
				}(partition, size, image)
			}
		}
	}

	wg.Wait()
	close(results)

	go func() {
		for result := range results {
			if len(result.errors) > 0 {
				log.Printf("\n\nError Summary:")
				log.Printf("the following machines did not succeed:")
			}
			for machine, err := range result.errors {
				logMachine(machine, err.Error())
			}

			log.Printf("\n\nTiming Summary:")
			for machine, duration := range result.durations {
				logMachine(machine, fmt.Sprintf("cycle duration:%s", duration))
			}
		}
	}()
}

func (r *runner) RunMachine(partition, size, image, name, project string) (*models.V1MachineResponse, error) {
	mcr := &metalgo.MachineCreateRequest{
		Description: name,
		Hostname:    name,
		Name:        name,
		Size:        size,
		Project:     project,
		Partition:   partition,
		Image:       image,
	}

	m, err := r.driver.MachineCreate(mcr)
	if err != nil {
		return nil, err
	}

	machine := m.Machine
	size = *m.Machine.Size.ID
	logMachine(machine, "Created")
	err = r.waitAndLog(machine, getReadyState(size))
	if err != nil {
		return machine, fmt.Errorf("machine did not come up:%v", err)
	}

	if getReadyState(size) == booting {
		// sleep a while to wait until booted into target OS
		logMachine(machine, "wait for booting finished")
		time.Sleep(30 * time.Second)
	}
	logMachine(machine, "Deleting")
	_, err = r.driver.MachineDelete(*machine.ID)
	if err != nil {
		return machine, fmt.Errorf("unable to delete machine:%v", err)
	}
	logMachine(machine, "Deleted")
	err = r.waitAndLog(machine, waiting)
	if err != nil {
		return machine, fmt.Errorf("machine cant be deleted:%v", err)
	}
	return machine, nil
}

func logMachine(machine *models.V1MachineResponse, state string) {
	log.Printf("partition:%s machine:%s size:%s image:%s %s\n",
		*machine.Partition.ID, *machine.ID, *machine.Size.ID, *machine.Allocation.Image.ID, state)
}

func getReadyState(size string) string {
	r, ok := readyState[size]
	if ok {
		return r
	}
	return phonedHome
}

func (r *runner) waitAndLog(machine *models.V1MachineResponse, targetState string) error {
	machineID := *machine.ID
	tick := time.NewTicker(500 * time.Millisecond)
	start := time.Now()
	lastState := ""
	reached := false
	for !reached {
		<-tick.C
		if time.Since(start) > r.maxWaitTime {
			return fmt.Errorf("max waittime exceeded")
		}
		m, err := r.driver.MachineGet(machineID)
		if err != nil {
			return err
		}

		if len(m.Machine.Events.Log) > 0 {
			lastLog := m.Machine.Events.Log[0]
			state := lastLog.Event
			if *state != lastState {
				logMachine(machine, *state)
				lastState = *state
			}
			if *state == targetState {
				reached = true
			}
			if *state == "Crashed" {
				return fmt.Errorf("%s", lastLog.Message)
			}
		}
	}
	return nil
}
