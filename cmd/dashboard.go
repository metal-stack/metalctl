package cmd

import (
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"
	"time"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	metalgo "github.com/metal-stack/metal-go"
	"github.com/metal-stack/metal-lib/pkg/tag"
	"github.com/metal-stack/metal-lib/rest"
	"github.com/metal-stack/v"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/sync/semaphore"
)

var (
	dashboardCmd = &cobra.Command{
		Use:   "dashboard",
		Short: "shows a live dashboard optimized for operation",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDashboard()
		},
		PreRun: bindPFlags,
	}
)

func init() {
	tabs := dashboardTabs()

	dashboardCmd.Flags().String("partition", "", "show resources in partition [optional]")
	dashboardCmd.Flags().String("size", "", "show machines with given size [optional]")
	dashboardCmd.Flags().String("color-theme", "default", "the dashboard's color theme [default|dark] [optional]")
	dashboardCmd.Flags().String("initial-tab", strings.ToLower(tabs[0].Name()), "the tab to show when starting the dashboard [optional]")
	dashboardCmd.Flags().Duration("refresh-interval", 3*time.Second, "refresh interval [optional]")

	err := dashboardCmd.RegisterFlagCompletionFunc("partition", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return partitionListCompletion(driver)
	})
	if err != nil {
		log.Fatal(err.Error())
	}
	err = dashboardCmd.RegisterFlagCompletionFunc("size", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return sizeListCompletion(driver)
	})
	if err != nil {
		log.Fatal(err.Error())
	}
	err = dashboardCmd.RegisterFlagCompletionFunc("color-theme", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{
			"default\twith bright fonts, optimized for dark terminal backgrounds",
			"dark\twith dark fonts, optimized for bright terminal backgrounds",
		}, cobra.ShellCompDirectiveNoFileComp
	})
	if err != nil {
		log.Fatal(err.Error())
	}
	err = dashboardCmd.RegisterFlagCompletionFunc("initial-tab", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		var names []string
		for _, t := range tabs {
			names = append(names, fmt.Sprintf("%s\t%s", strings.ToLower(t.Name()), t.Description()))
		}
		return names, cobra.ShellCompDirectiveNoFileComp
	})
	if err != nil {
		log.Fatal(err.Error())
	}
}

func dashboardApplyTheme(theme string) error {
	switch theme {
	case "default":
		ui.Theme.BarChart.Labels = []ui.Style{ui.NewStyle(ui.ColorWhite)}
		ui.Theme.BarChart.Nums = []ui.Style{ui.NewStyle(ui.ColorWhite)}

		ui.Theme.Gauge.Bar = ui.ColorWhite

		ui.Theme.Tab.Active = ui.NewStyle(ui.ColorYellow)
	case "dark":
		ui.Theme.Default = ui.NewStyle(ui.ColorBlack)
		ui.Theme.Block.Border = ui.NewStyle(ui.ColorBlack)
		ui.Theme.Block.Title = ui.NewStyle(ui.ColorBlack)

		ui.Theme.BarChart.Labels = []ui.Style{ui.NewStyle(ui.ColorBlack)}
		ui.Theme.BarChart.Nums = []ui.Style{ui.NewStyle(ui.ColorBlack)}

		ui.Theme.Gauge.Label = ui.NewStyle(ui.ColorBlack)
		ui.Theme.Gauge.Label.Fg = ui.ColorBlack
		ui.Theme.Gauge.Bar = ui.ColorBlack

		ui.Theme.Paragraph.Text = ui.NewStyle(ui.ColorBlack)

		ui.Theme.Tab.Active = ui.NewStyle(ui.ColorYellow)
		ui.Theme.Tab.Inactive = ui.NewStyle(ui.ColorBlack)

		ui.Theme.Table.Text = ui.NewStyle(ui.ColorBlack)
	default:
		return fmt.Errorf("unknown theme: %s", theme)
	}
	return nil
}

func runDashboard() error {
	if err := ui.Init(); err != nil {
		return err
	}
	defer ui.Close()

	var (
		interval      = viper.GetDuration("refresh-interval")
		width, height = ui.TerminalDimensions()
	)

	d, err := NewDashboard()
	if err != nil {
		return err
	}

	d.Resize(0, 0, width, height)
	d.Render()

	uiEvents := ui.PollEvents()
	ticker := time.NewTicker(interval)

	panelNumbers := map[string]bool{}
	for i := range d.tabs {
		panelNumbers[strconv.Itoa(i+1)] = true
	}

	for {
		select {
		case e := <-uiEvents:
			switch e.ID {
			case "q", "<C-c>":
				return nil
			case "<Resize>":
				payload := e.Payload.(ui.Resize)
				var (
					height = payload.Height
					width  = payload.Width
				)
				d.Resize(0, 0, width, height)
				ui.Clear()
				d.Render()
			default:
				_, ok := panelNumbers[e.ID]
				if ok {
					i, _ := strconv.Atoi(e.ID)
					d.tabPane.ActiveTabIndex = i - 1
					ui.Clear()
					d.Render()
				}
			}
		case <-ticker.C:
			d.Render()
		}
	}
}

func dashboardTabs() dashboardTabPanes {
	return dashboardTabPanes{
		NewDashboardMachinePane(),
	}
}

type dashboard struct {
	statusHeader *widgets.Paragraph
	filterHeader *widgets.Paragraph

	filterPartition string
	filterSize      string

	tabPane *widgets.TabPane
	tabs    dashboardTabPanes

	sem *semaphore.Weighted
}

type dashboardTabPane interface {
	Name() string
	Description() string
	Render() error
	Resize(x1, y1, x2, y2 int)
}

type dashboardTabPanes []dashboardTabPane

func (d dashboardTabPanes) FindIndexByName(name string) (int, error) {
	for i, p := range d {
		if strings.EqualFold(p.Name(), name) {
			return i, nil
		}
	}
	return 0, fmt.Errorf("tab with name %q not found", name)
}

func NewDashboard() (*dashboard, error) {
	err := dashboardApplyTheme(viper.GetString("color-theme"))
	if err != nil {
		return nil, err
	}

	d := &dashboard{
		sem:             semaphore.NewWeighted(1),
		filterPartition: viper.GetString("partition"),
		filterSize:      viper.GetString("size"),
	}

	d.statusHeader = widgets.NewParagraph()
	d.statusHeader.Title = "metal-stack Dashboard"
	d.statusHeader.WrapText = false

	d.filterHeader = widgets.NewParagraph()
	d.filterHeader.Title = "Filters"
	d.filterHeader.WrapText = false

	d.tabs = dashboardTabs()
	var tabNames []string
	for i, p := range d.tabs {
		tabNames = append(tabNames, fmt.Sprintf("(%d) %s", i+1, p.Name()))
	}
	d.tabPane = widgets.NewTabPane(tabNames...)
	d.tabPane.Title = "Tabs"
	d.tabPane.Border = false

	if viper.IsSet("initial-tab") {
		initialPanelIndex, err := d.tabs.FindIndexByName(viper.GetString("initial-tab"))
		if err != nil {
			return nil, err
		}
		d.tabPane.ActiveTabIndex = initialPanelIndex
	}

	return d, nil
}

func (d *dashboard) Resize(x1, y1, x2, y2 int) {
	d.statusHeader.SetRect(x1, y1, x2-25, d.headerHeight())
	d.filterHeader.SetRect(x2-25, y1, x2, d.headerHeight())

	for _, p := range d.tabs {
		p.Resize(x1, d.headerHeight(), x2, y2-1)
	}

	d.tabPane.SetRect(x1, y2-1, x2, y2)
}

func (d *dashboard) headerHeight() int {
	return 5
}

func (d *dashboard) Render() {
	if !d.sem.TryAcquire(1) { // prevent concurrent updates
		return
	}
	defer d.sem.Release(1)

	d.filterHeader.Text = fmt.Sprintf("Partition=%s\nSize=%s", d.filterPartition, d.filterSize)

	ui.Render(d.filterHeader, d.tabPane)

	var (
		apiVersion       = "unknown"
		apiHealth        = "unknown"
		apiHealthMessage string

		lastErr error
	)

	renderHeader := func() {
		var coloredHealth string
		switch apiHealth {
		case rest.HealthStatusHealthy:
			coloredHealth = "[" + apiHealth + "](fg:green)"
		case rest.HealthStatusUnhealthy:
			if apiHealthMessage != "" {
				coloredHealth = "[" + apiHealth + fmt.Sprintf(" (%s)](fg:red)", apiHealthMessage)
			} else {
				coloredHealth = "[" + apiHealth + "](fg:red)"
			}
		default:
			coloredHealth = apiHealth
		}

		versionLine := fmt.Sprintf("metal-api %s (API Health: %s), metalctl %s (%s)", apiVersion, coloredHealth, v.Version, v.GitSHA1)
		fetchInfoLine := fmt.Sprintf("Last Update: %s", time.Now().Format("15:04:05"))
		if lastErr != nil {
			fetchInfoLine += fmt.Sprintf(", [Update Error: %s](fg:red)", lastErr)
		}
		glossaryLine := "Switch between tabs with number keys. Press q to quit."

		d.statusHeader.Text = fmt.Sprintf("%s\n%s\n%s", versionLine, fetchInfoLine, glossaryLine)
		ui.Render(d.statusHeader)
	}
	defer renderHeader()

	var infoResp *metalgo.VersionGetResponse
	infoResp, lastErr = driver.VersionGet()
	if lastErr != nil {
		return
	}
	apiVersion = *infoResp.Version.Version

	var healthResp *metalgo.HealthGetResponse
	healthResp, lastErr = driver.HealthGet()
	if lastErr != nil {
		return
	}
	apiHealth = *healthResp.Health.Status
	apiHealthMessage = *healthResp.Health.Message

	renderHeader()

	lastErr = d.tabs[d.tabPane.ActiveTabIndex].Render()
}

type dashboardMachinePane struct {
	sem *semaphore.Weighted

	machineState  *widgets.BarChart
	machineIssues *widgets.BarChart

	partitionCapacity *widgets.BarChart

	freeMachines       *widgets.Gauge
	freeInternetIPs    *widgets.Gauge
	freeTenantPrefixes *widgets.Gauge
}

func NewDashboardMachinePane() *dashboardMachinePane {
	d := &dashboardMachinePane{}

	d.sem = semaphore.NewWeighted(1)

	d.partitionCapacity = widgets.NewBarChart()
	d.partitionCapacity.Labels = []string{"Free", "Allocated", "Other", "Faulty"}
	d.partitionCapacity.Title = "Partition Capacity"
	d.partitionCapacity.PaddingLeft = 5
	d.partitionCapacity.BarWidth = 5
	d.partitionCapacity.BarGap = 10
	d.partitionCapacity.BarColors = []ui.Color{ui.ColorGreen, ui.ColorGreen, ui.ColorYellow, ui.ColorRed}

	d.freeMachines = widgets.NewGauge()
	d.freeMachines.Title = "Free Machines"

	d.freeInternetIPs = widgets.NewGauge()
	d.freeInternetIPs.Title = "Free Internet IPs"

	d.freeTenantPrefixes = widgets.NewGauge()
	d.freeTenantPrefixes.Title = "Free Tenant Prefixes"

	d.machineIssues = widgets.NewBarChart()
	d.machineIssues.Labels = []string{"No Issues", "Issues"}
	d.machineIssues.Title = "Machine Issues"
	d.machineIssues.PaddingLeft = 5
	d.machineIssues.BarWidth = 5
	d.machineIssues.BarGap = 10
	d.machineIssues.BarColors = []ui.Color{ui.ColorGreen, ui.ColorRed}

	d.machineState = widgets.NewBarChart()
	d.machineState.Labels = []string{"Crashed", "PXE Booting", "Preparing", "Registering", "Waiting", "Installing", "Booting New Kernel", "Phoned Home", "Other"}
	d.machineState.Title = "Machine Provisioning State"
	d.machineState.PaddingLeft = 5
	d.machineState.BarWidth = 5
	d.machineState.BarGap = 10
	d.machineState.BarColors = []ui.Color{ui.ColorRed, ui.Theme.Default.Fg, ui.Theme.Default.Fg, ui.Theme.Default.Fg, ui.Theme.Default.Fg, ui.Theme.Default.Fg, ui.Theme.Default.Fg, ui.Theme.Default.Fg, ui.Theme.Default.Fg}

	return d
}

func (d *dashboardMachinePane) Name() string {
	return "Machines"
}

func (d *dashboardMachinePane) Description() string {
	return "Machine health and issues"
}

func (d *dashboardMachinePane) Resize(x1, y1, x2, y2 int) {
	rowHeight := int(math.Ceil((float64(y2) - (float64(y1))) / 2))
	columnWidth := int(math.Ceil((float64(x2) - (float64(x1))) / 3))

	d.partitionCapacity.SetRect(x1, y1, columnWidth*2, rowHeight-9)

	d.freeMachines.SetRect(x1, rowHeight-9, columnWidth*2, rowHeight-6)
	d.freeInternetIPs.SetRect(x1, rowHeight-6, columnWidth*2, rowHeight-3)
	d.freeTenantPrefixes.SetRect(x1, rowHeight-3, columnWidth*2, rowHeight)

	d.machineIssues.SetRect(columnWidth*2, y1, x2, rowHeight)

	d.machineState.SetRect(x1, rowHeight, x2, y2)
}

func (d *dashboardMachinePane) Render() error {
	if !d.sem.TryAcquire(1) { // prevent concurrent updates
		return nil
	}
	defer d.sem.Release(1)

	var (
		stateCrashed     int
		statePXE         int
		statePreparing   int
		stateRegistering int
		stateWaiting     int
		stateInstalling  int
		stateBooting     int
		statePhonedHome  int
		stateOther       int

		issues   int
		noIssues int

		freeInternetIPs    int
		freeTenantPrefixes int

		capFree      int
		capAllocated int
		capOther     int
		capFaulty    int
	)

	partitionResp, err := driver.PartitionCapacity(metalgo.PartitionCapacityRequest{
		ID:   viperString("partition"),
		Size: viperString("size"),
	})
	if err != nil {
		return err
	}
	capacity := partitionResp.Capacity

	for _, c := range capacity {
		for _, s := range c.Servers {
			capFree += int(*s.Free)
			capAllocated += int(*s.Allocated)
			capOther += int(*s.Other)
			capFaulty += int(*s.Faulty)
		}
	}

	totalMachines := capFree + capAllocated + capOther + capFaulty
	d.freeMachines.Percent = int((float64(capFree) / float64(totalMachines)) * 100)
	if d.freeMachines.Percent < 10 {
		d.freeMachines.BarColor = ui.ColorRed
	} else if d.freeMachines.Percent < 30 {
		d.freeMachines.BarColor = ui.ColorYellow
	} else {
		d.freeMachines.BarColor = ui.ColorGreen
	}
	ui.Render(d.freeMachines)

	// for some reason the UI hangs when all values are zero...
	if capFree > 0 || capAllocated > 0 || capOther > 0 || capFaulty > 0 {
		d.partitionCapacity.Data = []float64{float64(capFree), float64(capAllocated), float64(capOther), float64(capFaulty)}
		ui.Render(d.partitionCapacity)
	}

	networkResp, err := driver.NetworkFind(&metalgo.NetworkFindRequest{
		Labels: map[string]string{
			tag.NetworkDefault: "",
		},
	})
	if err != nil {
		return err
	}
	networks := networkResp.Networks

	if len(networks) == 1 {
		internetNetwork := networks[0]

		if internetNetwork.Usage != nil {
			availableInternetIPs := int(*internetNetwork.Usage.AvailableIps)
			freeInternetIPs = availableInternetIPs - int(*internetNetwork.Usage.UsedIps)

			d.freeInternetIPs.Percent = int(float64(freeInternetIPs) / float64(availableInternetIPs) * 100)
			if d.freeInternetIPs.Percent < 10 {
				d.freeInternetIPs.BarColor = ui.ColorRed
			} else if d.freeInternetIPs.Percent < 30 {
				d.freeInternetIPs.BarColor = ui.ColorYellow
			} else {
				d.freeInternetIPs.BarColor = ui.ColorGreen
			}
			ui.Render(d.freeInternetIPs)
		}
	} else {
		return fmt.Errorf("no: %d", len(networks))
	}

	networkResp, err = driver.NetworkFind(&metalgo.NetworkFindRequest{
		PartitionID:  viperString("partition"),
		PrivateSuper: boolPtr(true),
	})
	if err != nil {
		return err
	}
	networks = networkResp.Networks

	for _, n := range networks {
		if n.Usage == nil {
			continue
		}

		availablePrefixes := int(*n.Usage.AvailablePrefixes)
		freeTenantPrefixes = availablePrefixes - int(*n.Usage.UsedPrefixes)

		d.freeTenantPrefixes.Percent = int(float64(freeTenantPrefixes) / float64(availablePrefixes) * 100)
		if d.freeTenantPrefixes.Percent < 10 {
			d.freeTenantPrefixes.BarColor = ui.ColorRed
		} else if d.freeTenantPrefixes.Percent < 30 {
			d.freeTenantPrefixes.BarColor = ui.ColorYellow
		} else {
			d.freeTenantPrefixes.BarColor = ui.ColorGreen
		}
		ui.Render(d.freeTenantPrefixes)
	}

	machineResp, err := driver.MachineIPMIList(&metalgo.MachineFindRequest{
		PartitionID: viperString("partition"),
		SizeID:      viperString("size"),
	})
	if err != nil {
		return err
	}

	machines := machineResp.Machines

	for _, m := range machines {
		if m.Events != nil {
			if len(m.Events.Log) == 0 {
				stateOther++
			} else {
				switch *m.Events.Log[0].Event {
				case "Crashed":
					stateCrashed++
				case "PXE Booting":
					statePXE++
				case "Preparing":
					statePreparing++
				case "Registering":
					stateRegistering++
				case "Waiting":
					stateWaiting++
				case "Installing":
					stateInstalling++
				case "Phoned Home":
					statePhonedHome++
				default:
					stateOther++
				}
			}
		}
	}

	if len(machines) <= 0 {
		return nil
	}

	// for some reason the UI hangs when all values are zero...
	if stateCrashed > 0 || statePXE > 0 || statePreparing > 0 || stateRegistering > 0 || stateWaiting > 0 || stateInstalling > 0 || stateBooting > 0 || statePhonedHome > 0 || stateOther > 0 {
		d.machineState.Data = []float64{float64(stateCrashed), float64(statePXE), float64(statePreparing), float64(stateRegistering), float64(stateWaiting), float64(stateInstalling), float64(stateBooting), float64(statePhonedHome), float64(stateOther)}
		ui.Render(d.machineState)
	}

	issues = len(getMachineIssues(machines))
	noIssues = len(machines) - issues
	d.machineIssues.Data = []float64{float64(noIssues), float64(issues)}
	ui.Render(d.machineIssues)

	return nil
}
