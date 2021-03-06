package cmd

import (
	"encoding/base64"
	"log"
	"os"
	"os/exec"
	"strings"

	"fmt"

	"net/url"

	metalgo "github.com/metal-stack/metal-go"
	"github.com/metal-stack/metal-go/api/models"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type FilterOpts struct {
	ID        string
	Partition string
	Size      string
	Name      string
	Project   string
	Image     string
	Hostname  string
	Mac       string
	Tags      []string
}

const (
	// Port open on our control-plane to connect via ssh to get machine console access.
	bmcConsolePort = 5222
	forceFlag      = "yes-i-really-mean-it"
)

var (
	filterOpts = &FilterOpts{}

	machineCmd = &cobra.Command{
		Use:   "machine",
		Short: "manage machines",
		Long:  "metal machines are bare metal servers.",
	}

	machineCreateCmd = &cobra.Command{
		Use:   "create",
		Short: "create a machine",
		Long:  `create a new machine with the given operating system, the size and a project.`,
		Example: `machine create can be done in two different ways:

- default with automatic allocation:

metalctl machine create \
	--hostname worker01 \
	--name worker \
	--image ubuntu-18.04 \ # query available with: metalctl image list
	--size t1-small-x86 \  # query available with: metalctl size list
	--partition test \     # query available with: metalctl partition list
	--project cluster01 \
	--sshpublickey "@~/.ssh/id_rsa.pub"

- for metal administration with reserved machines:

reserve a machine you want to allocate:

metalctl machine reserve 00000000-0000-0000-0000-0cc47ae54694 --description "blocked for maintenance"

allocate this machine:

metalctl machine create \
	--hostname worker01 \
	--name worker \
	--image ubuntu-18.04 \ # query available with: metalctl image list
	--project cluster01 \
	--sshpublickey "@~/.ssh/id_rsa.pub" \
	--id 00000000-0000-0000-0000-0cc47ae54694

after you do not want to use this machine exclusive, remove the reservation:

metalctl machine reserve 00000000-0000-0000-0000-0cc47ae54694 --remove

Once created the machine installation can not be modified anymore.

`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return machineCreate(driver)
		},
		PreRun: bindPFlags,
	}

	machineListCmd = &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "list all machines",
		Long:    "list all machines with almost all properties in tabular form.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return machineList(driver)
		},
		PreRun: bindPFlags,
	}

	machineDescribeCmd = &cobra.Command{
		Use:   "describe <machine ID>",
		Short: "describe a machine",
		Long:  "describe a machine in a very detailed form with all properties.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return machineDescribe(driver, args)
		},
		PreRun: bindPFlags,
	}

	machineConsolePasswordCmd = &cobra.Command{
		Use:   "consolepassword <machine ID>",
		Short: "fetch the consolepassword for a machine",
		RunE: func(cmd *cobra.Command, args []string) error {
			return machineConsolePassword(driver, args)
		},
		PreRun: bindPFlags,
	}

	machineDestroyCmd = &cobra.Command{
		Use:     "destroy <machine ID>",
		Aliases: []string{"delete", "rm"},
		Short:   "destroy a machine",
		Long: `destroy a machine and destroy all data stored on the local disks. Once destroyed it is back for usage by other projects.
A destroyed machine can not restored anymore`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return machineDestroy(driver, args)
		},
		PreRun: bindPFlags,
	}

	machinePowerCmd = &cobra.Command{
		Use:   "power",
		Short: "manage machine power",
	}

	machinePowerOnCmd = &cobra.Command{
		Use:   "on <machine ID>",
		Short: "power on a machine",
		Long:  "set the machine to power on state, if the machine already was on nothing happens.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return machinePowerOn(driver, args)
		},
		PreRun: bindPFlags,
	}

	machinePowerOffCmd = &cobra.Command{
		Use:   "off <machine ID>",
		Short: "power off a machine",
		Long: `set the machine to power off state, if the machine already was off nothing happens.
It will usually take some time to power off the machine, depending on the machine type.
Power on will therefore not work if the machine is in the powering off phase.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return machinePowerOff(driver, args)
		},
		PreRun: bindPFlags,
	}

	machinePowerResetCmd = &cobra.Command{
		Use:   "reset <machine ID>",
		Short: "power reset a machine",
		Long:  "reset the machine power. This will ensure a power cycle.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return machinePowerReset(driver, args)
		},
		PreRun: bindPFlags,
	}

	machineUpdateCmd = &cobra.Command{
		Use:     "update",
		Aliases: []string{"firmware-update"},
		Short:   "update a machine firmware",
	}

	machineUpdateBiosCmd = &cobra.Command{
		Use:   "bios <machine ID>",
		Short: "update a machine BIOS",
		Long:  "the machine BIOS will be updated to given revision. If revision flag is not specified an update plan will be printed instead.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return machineUpdateBios(driver, args)
		},
		PreRun: bindPFlags,
	}

	machineUpdateBmcCmd = &cobra.Command{
		Use:   "bmc <machine ID>",
		Short: "update a machine BMC",
		Long:  "the machine BMC will be updated to given revision. If revision flag is not specified an update plan will be printed instead.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return machineUpdateBmc(driver, args)
		},
		PreRun: bindPFlags,
	}

	machineBootBiosCmd = &cobra.Command{
		Use:   "bios <machine ID>",
		Short: "boot a machine into BIOS",
		Long:  "the machine will boot into bios.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return machineBootBios(driver, args)
		},
		PreRun: bindPFlags,
	}

	machineBootPxeCmd = &cobra.Command{
		Use:   "pxe <machine ID>",
		Short: "boot a machine from PXE",
		Long:  "the machine will boot from PXE.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return machineBootPxe(driver, args)
		},
		PreRun: bindPFlags,
	}

	machineBootDiskCmd = &cobra.Command{
		Use:   "disk <machine ID>",
		Short: "boot a machine from disk",
		Long:  "the machine will boot from disk.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return machineBootDisk(driver, args)
		},
		PreRun: bindPFlags,
	}

	machineIdentifyCmd = &cobra.Command{
		Use:   "identify",
		Short: "manage machine chassis identify LED power",
	}

	machineIdentifyOnCmd = &cobra.Command{
		Use:   "on <machine ID>",
		Short: "power on the machine chassis identify LED",
		Long:  `set the machine chassis identify LED to on state`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return machineIdentifyOn(driver, args)
		},
		PreRun: bindPFlags,
	}

	machineIdentifyOffCmd = &cobra.Command{
		Use:   "off <machine ID>",
		Short: "power off the machine chassis identify LED",
		Long:  `set the machine chassis identify LED to off state`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return machineIdentifyOff(driver, args)
		},
		PreRun: bindPFlags,
	}

	machineReserveCmd = &cobra.Command{
		Use:   "reserve <machine ID>",
		Short: "reserve a machine",
		Long: `reserve a machine for exclusive usage, this machine will no longer be picked by other allocations.
This is useful for maintenance of the machine or testing. After the reservation is not needed anymore, the reservation
should be removed with --remove.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return machineReserve(driver, args)
		},
		PreRun: bindPFlags,
	}

	machineLockCmd = &cobra.Command{
		Use:   "lock <machine ID>",
		Short: "lock a machine",
		Long:  `when a machine is locked, it can not be destroyed, to destroy a machine you must first remove the lock from that machine with --remove`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return machineLock(driver, args)
		},
		PreRun: bindPFlags,
	}

	machineReinstallCmd = &cobra.Command{
		Use:   "reinstall <machine ID>",
		Short: "reinstalls an already allocated machine",
		Long: `reinstalls an already allocated machine. If it is not yet allocated, nothing happens, otherwise only the machine's primary disk
is wiped and the new image will subsequently be installed on that device`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return machineReinstall(driver, args)
		},
		PreRun: bindPFlags,
	}

	machineConsoleCmd = &cobra.Command{
		Use: "console <machine ID>",
		Short: `console access to a machine, machine must be created with a ssh public key, authentication is done with your private key.
In case the machine did not register properly a direct ipmi console access is available via the --ipmi flag. This is only for administrative access.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return machineConsole(driver, args)
		},
		PreRun: bindPFlags,
	}
	machineIpmiCmd = &cobra.Command{
		Use:   "ipmi [<machine ID>]",
		Short: `display ipmi details of the machine, if no machine ID is given all ipmi addresses are returned.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return machineIpmi(driver, args)
		},
		PreRun: bindPFlags,
	}
	machineIssuesCmd = &cobra.Command{
		Use:   "issues",
		Short: `display machines which are in a potential bad state`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return machineIssues(driver)
		},
		PreRun: bindPFlags,
	}
	machineLogsCmd = &cobra.Command{
		Use:     "logs <machine ID>",
		Aliases: []string{"log"},
		Short:   `display machine provisioning logs`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return machineLogs(driver, args)
		},
		PreRun: bindPFlags,
	}
)

func init() {
	addMachineCreateFlags(machineCreateCmd, "machine")
	machineCmd.AddCommand(machineCreateCmd)

	machineListCmd.Flags().StringVarP(&filterOpts.ID, "id", "", "", "ID to filter [optional]")
	machineListCmd.Flags().StringVarP(&filterOpts.Partition, "partition", "", "", "partition to filter [optional]")
	machineListCmd.Flags().StringVarP(&filterOpts.Size, "size", "", "", "size to filter [optional]")
	machineListCmd.Flags().StringVarP(&filterOpts.Name, "name", "", "", "allocation name to filter [optional]")
	machineListCmd.Flags().StringVarP(&filterOpts.Project, "project", "", "", "allocation project to filter [optional]")
	machineListCmd.Flags().StringVarP(&filterOpts.Image, "image", "", "", "allocation image to filter [optional]")
	machineListCmd.Flags().StringVarP(&filterOpts.Hostname, "hostname", "", "", "allocation hostname to filter [optional]")
	machineListCmd.Flags().StringVarP(&filterOpts.Mac, "mac", "", "", "mac to filter [optional]")
	machineListCmd.Flags().StringSliceVar(&filterOpts.Tags, "tags", []string{}, "tags to filter, use it like: --tags \"tag1,tag2\" or --tags \"tag3\".")

	err := machineListCmd.RegisterFlagCompletionFunc("partition", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return partitionListCompletion(driver)
	})
	if err != nil {
		log.Fatal(err.Error())
	}
	err = machineListCmd.RegisterFlagCompletionFunc("size", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return sizeListCompletion(driver)
	})
	if err != nil {
		log.Fatal(err.Error())
	}
	err = machineListCmd.RegisterFlagCompletionFunc("project", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return projectListCompletion(driver)
	})
	if err != nil {
		log.Fatal(err.Error())
	}
	err = machineListCmd.RegisterFlagCompletionFunc("id", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return machineListCompletion(driver)
	})
	if err != nil {
		log.Fatal(err.Error())
	}
	err = machineListCmd.RegisterFlagCompletionFunc("image", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return imageListCompletion(driver)
	})
	if err != nil {
		log.Fatal(err.Error())
	}

	machineIpmiCmd.Flags().StringVarP(&filterOpts.ID, "id", "", "", "ID to filter [optional]")
	machineIpmiCmd.Flags().StringVarP(&filterOpts.Partition, "partition", "", "", "partition to filter [optional]")
	machineIpmiCmd.Flags().StringVarP(&filterOpts.Size, "size", "", "", "size to filter [optional]")
	machineIpmiCmd.Flags().StringVarP(&filterOpts.Name, "name", "", "", "allocation name to filter [optional]")
	machineIpmiCmd.Flags().StringVarP(&filterOpts.Project, "project", "", "", "allocation project to filter [optional]")
	machineIpmiCmd.Flags().StringVarP(&filterOpts.Image, "image", "", "", "allocation image to filter [optional]")
	machineIpmiCmd.Flags().StringVarP(&filterOpts.Hostname, "hostname", "", "", "allocation hostname to filter [optional]")
	machineIpmiCmd.Flags().StringVarP(&filterOpts.Mac, "mac", "", "", "mac to filter [optional]")
	machineIpmiCmd.Flags().StringSliceVar(&filterOpts.Tags, "tags", []string{}, "tags to filter, use it like: --tags \"tag1,tag2\" or --tags \"tag3\".")

	err = machineIpmiCmd.RegisterFlagCompletionFunc("partition", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return partitionListCompletion(driver)
	})
	if err != nil {
		log.Fatal(err.Error())
	}
	err = machineIpmiCmd.RegisterFlagCompletionFunc("size", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return sizeListCompletion(driver)
	})
	if err != nil {
		log.Fatal(err.Error())
	}
	err = machineIpmiCmd.RegisterFlagCompletionFunc("project", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return projectListCompletion(driver)
	})
	if err != nil {
		log.Fatal(err.Error())
	}
	err = machineIpmiCmd.RegisterFlagCompletionFunc("id", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return machineListCompletion(driver)
	})
	if err != nil {
		log.Fatal(err.Error())
	}

	machineConsolePasswordCmd.Flags().StringP("reason", "", "", "a short description why access to the consolepassword is required")

	machineCmd.AddCommand(machineListCmd)
	machineCmd.AddCommand(machineDestroyCmd)
	machineCmd.AddCommand(machineDescribeCmd)
	machineCmd.AddCommand(machineConsolePasswordCmd)

	machineUpdateBiosCmd.Flags().StringP("revision", "", "", "the BIOS revision")
	machineUpdateBiosCmd.Flags().StringP("description", "", "", "the reason why the BIOS should be updated")
	err = machineUpdateBiosCmd.RegisterFlagCompletionFunc("revision", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return firmwareRevisionCompletion(driver)
	})
	if err != nil {
		log.Fatal(err.Error())
	}
	machineUpdateCmd.AddCommand(machineUpdateBiosCmd)

	machineUpdateBmcCmd.Flags().StringP("revision", "", "", "the BMC revision")
	machineUpdateBmcCmd.Flags().StringP("description", "", "", "the reason why the BMC should be updated")
	err = machineUpdateBmcCmd.RegisterFlagCompletionFunc("revision", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return firmwareRevisionCompletion(driver)
	})
	if err != nil {
		log.Fatal(err.Error())
	}
	machineUpdateCmd.AddCommand(machineUpdateBmcCmd)

	machineCmd.AddCommand(machineUpdateCmd)

	machinePowerCmd.AddCommand(machinePowerOnCmd)
	machinePowerCmd.AddCommand(machinePowerOffCmd)
	machinePowerCmd.AddCommand(machinePowerResetCmd)
	machinePowerCmd.AddCommand(machineBootBiosCmd)
	machinePowerCmd.AddCommand(machineBootDiskCmd)
	machinePowerCmd.AddCommand(machineBootPxeCmd)
	machineCmd.AddCommand(machinePowerCmd)

	machineIdentifyOnCmd.Flags().StringP("description", "d", "", "description of the reason for chassis identify LED turn-on.")
	machineIdentifyCmd.AddCommand(machineIdentifyOnCmd)

	machineIdentifyOffCmd.Flags().StringP("description", "d", "Triggered by metalctl", "description of the reason for chassis identify LED turn-off.")
	machineIdentifyCmd.AddCommand(machineIdentifyOffCmd)
	machineCmd.AddCommand(machineIdentifyCmd)

	machineReserveCmd.Flags().StringP("description", "d", "", "description of the reason for the reservation.")
	machineReserveCmd.Flags().BoolP("remove", "r", false, "remove the reservation.")
	machineCmd.AddCommand(machineReserveCmd)

	machineLockCmd.Flags().StringP("description", "d", "", "description of the reason for the lock.")
	machineLockCmd.Flags().BoolP("remove", "r", false, "remove the lock.")
	machineCmd.AddCommand(machineLockCmd)

	machineReinstallCmd.Flags().StringP("image", "", "", "id of the image to get installed. [required]")
	machineReinstallCmd.Flags().StringP("description", "d", "", "description of the reinstallation. [optional]")
	err = machineReinstallCmd.MarkFlagRequired("image")
	if err != nil {
		log.Fatal(err.Error())
	}
	machineCmd.AddCommand(machineReinstallCmd)

	machineConsoleCmd.Flags().StringP("sshidentity", "p", "", "SSH key file, if not given the default ssh key will be used if present [optional].")
	machineConsoleCmd.Flags().BoolP("ipmi", "", false, "use ipmitool with direct network access (admin only).")
	machineConsoleCmd.Flags().StringP("ipmiuser", "", "", "overwrite ipmi user (admin only).")
	machineConsoleCmd.Flags().StringP("ipmipassword", "", "", "overwrite ipmi password (admin only).")
	machineCmd.AddCommand(machineConsoleCmd)
	machineCmd.AddCommand(machineIpmiCmd)
	machineCmd.AddCommand(machineIssuesCmd)
	machineCmd.AddCommand(machineLogsCmd)

	machineDestroyCmd.Flags().Bool("remove-from-database", false, "remove given machine from the database, is only required for maintenance reasons [optional] (admin only).")
}

func addMachineCreateFlags(cmd *cobra.Command, name string) {
	cmd.Flags().StringP("description", "d", "", "Description of the "+name+" to create. [optional]")
	cmd.Flags().StringP("partition", "S", "", "partition/datacenter where the "+name+" is created. [required, except for reserved machines]")
	cmd.Flags().StringP("hostname", "H", "", "Hostname of the "+name+". [required]")
	cmd.Flags().StringP("image", "i", "", "OS Image to install. [required]")
	cmd.Flags().StringP("filesystemlayout", "", "", "Filesystemlayout to use during machine installation. [optional]")
	cmd.Flags().StringP("name", "n", "", "Name of the "+name+". [optional]")
	cmd.Flags().StringP("id", "I", "", "ID of a specific "+name+" to allocate, if given, size and partition are ignored. Need to be set to reserved (--reserve) state before.")
	cmd.Flags().StringP("project", "P", "", "Project where the "+name+" should belong to. [required]")
	cmd.Flags().StringP("size", "s", "", "Size of the "+name+". [required, except for reserved machines]")
	cmd.Flags().StringP("sshpublickey", "p", "",
		`SSH public key for access via ssh and console. [optional]
Can be either the public key as string, or pointing to the public key file to use e.g.: "@~/.ssh/id_rsa.pub".
If ~/.ssh/[id_ed25519.pub | id_rsa.pub | id_dsa.pub] is present it will be picked as default, matching the first one in this order.`)
	cmd.Flags().StringSlice("tags", []string{}, "tags to add to the "+name+", use it like: --tags \"tag1,tag2\" or --tags \"tag3\".")
	cmd.Flags().StringP("userdata", "", "", `cloud-init.io compatible userdata. [optional]
Can be either the userdata as string, or pointing to the userdata file to use e.g.: "@/tmp/userdata.cfg".`)

	switch name {
	case "machine":
		cmd.Flags().StringSlice("ips", []string{},
			`Sets the machine's IP address. Usage: [--ips[=IPV4-ADDRESS[,IPV4-ADDRESS]...]]...
IPV4-ADDRESS specifies the IPv4 address to add.
It can only be used in conjunction with --networks.`)
		cmd.Flags().StringSlice("networks", []string{},
			`Adds a network. Usage: [--networks NETWORK[:MODE][,NETWORK[:MODE]]...]...
NETWORK specifies the name or id of an existing network.
MODE cane be omitted or one of:
	auto	IP address is automatically acquired from the given network
	noauto	IP address for the given network must be provided via --ips`)
		err := cmd.MarkFlagRequired("networks")
		if err != nil {
			log.Fatal(err.Error())
		}
	case "firewall":
		cmd.Flags().StringSlice("ips", []string{},
			`Sets the firewall's IP address. Usage: [--ips[=IPV4-ADDRESS[,IPV4-ADDRESS]...]]...
IPV4-ADDRESS specifies the IPv4 address to add.
It can only be used in conjunction with --networks.`)
		cmd.Flags().StringSlice("networks", []string{},
			`Adds network(s). Usage: --networks NETWORK[:MODE][,NETWORK[:MODE]]... [--networks NETWORK[:MODE][,
NETWORK[:MODE]]...]...
NETWORK specifies the id of an existing network.
MODE can be omitted or one of:
	auto	IP address is automatically acquired from the given network
	noauto	No automatic IP address acquisition`)
		err := cmd.MarkFlagRequired("networks")
		if err != nil {
			log.Fatal(err.Error())
		}
	default:
		log.Fatal(fmt.Errorf("illegal name: %s. Must be one of (machine, firewall)", name))
	}

	err := cmd.MarkFlagRequired("hostname")
	if err != nil {
		log.Fatal(err.Error())
	}
	err = cmd.MarkFlagRequired("image")
	if err != nil {
		log.Fatal(err.Error())
	}
	err = cmd.MarkFlagRequired("project")
	if err != nil {
		log.Fatal(err.Error())
	}

	// Completion for arguments
	err = cmd.RegisterFlagCompletionFunc("networks", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return networkListCompletion(driver)
	})
	if err != nil {
		log.Fatal(err.Error())
	}
	err = cmd.RegisterFlagCompletionFunc("ips", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return ipListCompletion(driver)
	})
	if err != nil {
		log.Fatal(err.Error())
	}
	err = cmd.RegisterFlagCompletionFunc("partition", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return partitionListCompletion(driver)
	})
	if err != nil {
		log.Fatal(err.Error())
	}
	err = cmd.RegisterFlagCompletionFunc("size", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return sizeListCompletion(driver)
	})
	if err != nil {
		log.Fatal(err.Error())
	}
	err = cmd.RegisterFlagCompletionFunc("project", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return projectListCompletion(driver)
	})
	if err != nil {
		log.Fatal(err.Error())
	}
	err = cmd.RegisterFlagCompletionFunc("id", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return machineListCompletion(driver)
	})
	if err != nil {
		log.Fatal(err.Error())
	}
	err = cmd.RegisterFlagCompletionFunc("image", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return imageListCompletion(driver)
	})
	if err != nil {
		log.Fatal(err.Error())
	}
	err = cmd.RegisterFlagCompletionFunc("filesystemlayout", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return filesystemLayoutListCompletion(driver)
	})
	if err != nil {
		log.Fatal(err.Error())
	}
}

func machineCreate(driver *metalgo.Driver) error {
	mcr, err := machineCreateRequest()
	if err != nil {
		return fmt.Errorf("machine create error:%w", err)
	}
	resp, err := driver.MachineCreate(mcr)
	if err != nil {
		return fmt.Errorf("machine create error:%w", err)
	}
	return printer.Print(resp.Machine)
}

func machineCreateRequest() (*metalgo.MachineCreateRequest, error) {
	sshPublicKeyArgument := viper.GetString("sshpublickey")

	if strings.HasPrefix(sshPublicKeyArgument, "@") {
		var err error
		sshPublicKeyArgument, err = readFromFile(sshPublicKeyArgument[1:])
		if err != nil {
			return nil, err
		}
	}

	if len(sshPublicKeyArgument) == 0 {
		sshKey, err := searchSSHKey()
		if err != nil {
			return nil, err
		}
		sshPublicKey := sshKey + ".pub"
		sshPublicKeyArgument, err = readFromFile(sshPublicKey)
		if err != nil {
			return nil, err
		}
	}

	var keys []string
	if sshPublicKeyArgument != "" {
		keys = append(keys, sshPublicKeyArgument)
	}

	userDataArgument := viper.GetString("userdata")
	if strings.HasPrefix(userDataArgument, "@") {
		var err error
		userDataArgument, err = readFromFile(userDataArgument[1:])
		if err != nil {
			return nil, err
		}
	}
	if userDataArgument != "" {
		userDataArgument = base64.StdEncoding.EncodeToString([]byte(userDataArgument))
	}

	possibleNetworks := viper.GetStringSlice("networks")
	networks, err := parseNetworks(possibleNetworks)
	if err != nil {
		return nil, err
	}

	mcr := &metalgo.MachineCreateRequest{
		Description:   viper.GetString("description"),
		Partition:     viper.GetString("partition"),
		Hostname:      viper.GetString("hostname"),
		Image:         viper.GetString("image"),
		Name:          viper.GetString("name"),
		UUID:          viper.GetString("id"),
		Project:       viper.GetString("project"),
		Size:          viper.GetString("size"),
		SSHPublicKeys: keys,
		Tags:          viper.GetStringSlice("tags"),
		UserData:      userDataArgument,
		Networks:      networks,
		IPs:           viper.GetStringSlice("ips"),
	}
	if viper.GetString("filesystemlayout") != "" {
		mcr.FilesystemLayout = viper.GetString("filesystemlayout")
	}

	return mcr, nil
}

func machineList(driver *metalgo.Driver) error {
	var resp *metalgo.MachineListResponse
	var err error
	if atLeastOneViperStringFlagGiven("id", "partition", "size", "name", "project", "image", "hostname", "mac") ||
		atLeastOneViperStringSliceFlagGiven("tags") {
		mfr := &metalgo.MachineFindRequest{}
		if filterOpts.ID != "" {
			mfr.ID = &filterOpts.ID
		}
		if filterOpts.Partition != "" {
			mfr.PartitionID = &filterOpts.Partition
		}
		if filterOpts.Size != "" {
			mfr.SizeID = &filterOpts.Size
		}
		if filterOpts.Name != "" {
			mfr.AllocationName = &filterOpts.Name
		}
		if filterOpts.Project != "" {
			mfr.AllocationProject = &filterOpts.Project
		}
		if filterOpts.Image != "" {
			mfr.AllocationImageID = &filterOpts.Image
		}
		if filterOpts.Hostname != "" {
			mfr.AllocationHostname = &filterOpts.Hostname
		}
		if filterOpts.Hostname != "" {
			mfr.AllocationHostname = &filterOpts.Hostname
		}
		if filterOpts.Mac != "" {
			mfr.NicsMacAddresses = []string{filterOpts.Mac}
		}
		if len(filterOpts.Tags) > 0 {
			mfr.Tags = filterOpts.Tags
		}
		resp, err = driver.MachineFind(mfr)
	} else {
		resp, err = driver.MachineList()
	}
	if err != nil {
		return err
	}
	return printer.Print(resp.Machines)
}

func machineDescribe(driver *metalgo.Driver, args []string) error {
	machineID, err := getMachineID(args)
	if err != nil {
		return err
	}
	resp, err := driver.MachineGet(machineID)
	if err != nil {
		return err
	}
	return detailer.Detail(resp.Machine)
}

func machineConsolePassword(driver *metalgo.Driver, args []string) error {
	machineID, err := getMachineID(args)
	if err != nil {
		return err
	}
	reason := viper.GetString("reason")
	resp, err := driver.MachineConsolePassword(machineID, reason)
	if err != nil {
		return err
	}
	fmt.Printf("%s\n", *resp.ConsolePassword)
	return nil
}

func machineDestroy(driver *metalgo.Driver, args []string) error {
	machineID, err := getMachineID(args)
	if err != nil {
		return err
	}

	if viper.GetBool("remove-from-database") {
		if !viper.GetBool(forceFlag) {
			return fmt.Errorf("remove-from-database is set but you forgot to add --%s", forceFlag)
		}
		resp, err := driver.MachineDeleteFromDatabase(machineID)
		if err != nil {
			return err
		}
		return printer.Print(resp.Machine)
	}

	resp, err := driver.MachineDelete(machineID)
	if err != nil {
		return err
	}
	return printer.Print(resp.Machine)
}

func machinePowerOn(driver *metalgo.Driver, args []string) error {
	machineID, err := getMachineID(args)
	if err != nil {
		return err
	}

	resp, err := driver.MachinePowerOn(machineID)
	if err != nil {
		return err
	}
	return printer.Print(resp.Machine)
}

func machinePowerOff(driver *metalgo.Driver, args []string) error {
	machineID, err := getMachineID(args)
	if err != nil {
		return err
	}

	resp, err := driver.MachinePowerOff(machineID)
	if err != nil {
		return err
	}
	return printer.Print(resp.Machine)
}

func machinePowerReset(driver *metalgo.Driver, args []string) error {
	machineID, err := getMachineID(args)
	if err != nil {
		return err
	}

	resp, err := driver.MachinePowerReset(machineID)
	if err != nil {
		return err
	}
	return printer.Print(resp.Machine)
}

func machineUpdateBios(driver *metalgo.Driver, args []string) error {
	m, vendor, board, err := firmwareData(args)
	if err != nil {
		return err
	}
	revision := viper.GetString("revision")
	currentVersion := ""
	if m.Bios != nil && m.Bios.Version != nil {
		currentVersion = *m.Bios.Version
	}

	return machineUpdateFirmware(driver, metalgo.Bios, *m.ID, vendor, board, revision, currentVersion)
}

func machineUpdateBmc(driver *metalgo.Driver, args []string) error {
	m, vendor, board, err := firmwareData(args)
	if err != nil {
		return err
	}
	revision := viper.GetString("revision")
	currentVersion := ""
	if m.Ipmi != nil && m.Ipmi.Bmcversion != nil {
		currentVersion = *m.Ipmi.Bmcversion
	}

	return machineUpdateFirmware(driver, metalgo.Bmc, *m.ID, vendor, board, revision, currentVersion)
}

func firmwareData(args []string) (*models.V1MachineIPMIResponse, string, string, error) {
	m, err := getMachine(args)
	if err != nil {
		return nil, "", "", err
	}
	machineID := *m.ID
	if m.Ipmi == nil {
		return nil, "", "", fmt.Errorf("no ipmi data available of machine %s", machineID)
	}

	fru := *m.Ipmi.Fru
	vendor := strings.ToLower(fru.ProductManufacturer)
	board := strings.ToUpper(fru.BoardPartNumber)

	return m, vendor, board, nil
}

func machineUpdateFirmware(driver *metalgo.Driver, kind metalgo.FirmwareKind, machineID, vendor, board, revision, currentVersion string) error {
	f, err := driver.ListFirmwares(kind, "", "")
	if err != nil {
		return err
	}

	var rr []string
	revisionAvailable, containsCurrentVersion := false, false
	vv, ok := f.Firmwares.Revisions[string(kind)]
	if ok {
		bb, ok := vv.VendorRevisions[vendor]
		if ok {
			rr, ok = bb.BoardRevisions[board]
			if ok {
				for _, rev := range rr {
					if rev == revision {
						revisionAvailable = true
					}
					if rev == currentVersion {
						containsCurrentVersion = true
					}
				}
			}
		}
	}

	printPlan := revision == "" || !revisionAvailable
	if printPlan {
		fmt.Println("Available:")
		for _, rev := range rr {
			if rev == currentVersion {
				fmt.Printf("%s (current)\n", rev)
			} else {
				fmt.Println(rev)
			}
		}
		if !containsCurrentVersion {
			fmt.Printf("---\nCurrent %s version: %s\n", strings.ToUpper(string(kind)), currentVersion)
		}
	}

	if revision == "" {
		return nil
	}
	if !revisionAvailable {
		return fmt.Errorf("specified revision %s not available", revision)
	}

	if !viper.GetBool(forceFlag) {
		return fmt.Errorf("flag %q not set", forceFlag)
	}

	description := viper.GetString("description")
	if description == "" {
		description = "unknown"
	}

	resp, err := driver.MachineUpdateFirmware(kind, machineID, revision, description)
	if err != nil {
		return err
	}
	return printer.Print(resp.Machine)
}

func machineBootBios(driver *metalgo.Driver, args []string) error {
	machineID, err := getMachineID(args)
	if err != nil {
		return err
	}

	resp, err := driver.MachineBootBios(machineID)
	if err != nil {
		return err
	}
	return printer.Print(resp.Machine)
}

func machineBootDisk(driver *metalgo.Driver, args []string) error {
	machineID, err := getMachineID(args)
	if err != nil {
		return err
	}

	resp, err := driver.MachineBootDisk(machineID)
	if err != nil {
		return err
	}
	return printer.Print(resp.Machine)
}

func machineBootPxe(driver *metalgo.Driver, args []string) error {
	machineID, err := getMachineID(args)
	if err != nil {
		return err
	}

	resp, err := driver.MachineBootPxe(machineID)
	if err != nil {
		return err
	}
	return printer.Print(resp.Machine)
}

func machineIdentifyOn(driver *metalgo.Driver, args []string) error {
	machineID, err := getMachineID(args)
	if err != nil {
		return err
	}

	description := viper.GetString("description")
	resp, err := driver.ChassisIdentifyLEDPowerOn(machineID, description)
	if err != nil {
		return err
	}
	return printer.Print(resp.Machine)
}

func machineIdentifyOff(driver *metalgo.Driver, args []string) error {
	machineID, err := getMachineID(args)
	if err != nil {
		return err
	}

	description := viper.GetString("description")
	resp, err := driver.ChassisIdentifyLEDPowerOff(machineID, description)
	if err != nil {
		return err
	}
	return printer.Print(resp.Machine)
}

func machineReserve(driver *metalgo.Driver, args []string) error {
	machineID, err := getMachineID(args)
	if err != nil {
		return err
	}
	description := viper.GetString("description")
	remove := viper.GetBool("remove")

	var resp *metalgo.MachineStateResponse
	if remove {
		resp, err = driver.MachineUnReserve(machineID)
		if err != nil {
			return err
		}
	} else {
		resp, err = driver.MachineReserve(machineID, description)
		if err != nil {
			return err
		}
	}
	return printer.Print(resp.Machine)
}

func machineLock(driver *metalgo.Driver, args []string) error {
	machineID, err := getMachineID(args)
	if err != nil {
		return err
	}
	description := viper.GetString("description")
	remove := viper.GetBool("remove")

	var resp *metalgo.MachineStateResponse
	if remove {
		resp, err = driver.MachineUnLock(machineID)
		if err != nil {
			return err
		}
	} else {
		resp, err = driver.MachineLock(machineID, description)
		if err != nil {
			return err
		}
	}
	return printer.Print(resp.Machine)
}

func machineReinstall(driver *metalgo.Driver, args []string) error {
	machineID, err := getMachineID(args)
	if err != nil {
		return err
	}
	imageID := viper.GetString("image")
	description := viper.GetString("description")

	var resp *metalgo.MachineGetResponse
	resp, err = driver.MachineReinstall(machineID, imageID, description)
	if err != nil {
		return err
	}
	return printer.Print(resp.Machine)
}

func machineLogs(driver *metalgo.Driver, args []string) error {
	// FIXME add ipmi sel as well
	machineID, err := getMachineID(args)
	if err != nil {
		return err
	}

	resp, err := driver.MachineGet(machineID)
	if err != nil {
		return err
	}
	machine := resp.Machine
	return printer.Print(machine.Events.Log)
}

func machineConsole(driver *metalgo.Driver, args []string) error {
	machineID, err := getMachineID(args)
	if err != nil {
		return err
	}
	useIpmi := viper.GetBool("ipmi")
	if useIpmi {
		path, err := exec.LookPath("ipmitool")
		if err != nil {
			return fmt.Errorf("unable to locate ipmitool in path")
		}

		resp, err := driver.MachineIPMIGet(machineID)
		if err != nil {
			return err
		}

		ipmi := resp.Machine.Ipmi
		intf := "lanplus"
		if *ipmi.Interface != "" {
			intf = *ipmi.Interface
		}
		// -I lanplus  -H 192.168.2.19 -U ADMIN -P ADMIN sol activate
		hostAndPort := strings.Split(*ipmi.Address, ":")
		if len(hostAndPort) < 2 {
			hostAndPort = append(hostAndPort, "623")
		}
		usr := *ipmi.User
		if *ipmi.User == "" {
			fmt.Printf("no ipmi user stored, please specify with --ipmiuser\n")
		}
		ipmiuser := viper.GetString("ipmiuser")
		if ipmiuser != "" {
			usr = ipmiuser
		}
		password := *ipmi.Password
		if *ipmi.Password == "" {
			fmt.Printf("no ipmi password stored, please specify with --ipmipassword\n")
		}

		ipmipassword := viper.GetString("ipmipassword")
		if ipmipassword != "" {
			password = ipmipassword
		}

		args := []string{"-I", intf, "-H", hostAndPort[0], "-p", hostAndPort[1], "-U", usr, "-P", "<hidden>", "sol", "activate"}
		fmt.Printf("connecting to console with:\n%s %s\nExit with ~.\n\n", path, strings.Join(args, " "))
		args[9] = password
		cmd := exec.Command(path, args...)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stdout
		return cmd.Run()
	}

	key := viper.GetString("sshidentity")
	if key == "" {
		key, err = searchSSHKey()
		if err != nil {
			return fmt.Errorf("machine console error:%w", err)
		}
	}
	parsedurl, err := url.Parse(driverURL)
	if err != nil {
		return err
	}
	err = SSHClient(machineID, key, parsedurl.Host, bmcConsolePort)
	if err != nil {
		return fmt.Errorf("machine console error:%w", err)
	}
	return nil
}

func machineIpmi(driver *metalgo.Driver, args []string) error {
	if len(args) == 1 {
		machineID, err := getMachineID(args)
		if err != nil {
			return err
		}
		resp, err := driver.MachineIPMIGet(machineID)
		if err != nil {
			return err
		}

		hidden := "<hidden>"
		resp.Machine.Ipmi.Password = &hidden
		return detailer.Detail(resp.Machine)
	}

	mfr := &metalgo.MachineFindRequest{}
	if filterOpts.ID != "" {
		mfr.ID = &filterOpts.ID
	}
	if filterOpts.Partition != "" {
		mfr.PartitionID = &filterOpts.Partition
	}
	if filterOpts.Size != "" {
		mfr.SizeID = &filterOpts.Size
	}
	if filterOpts.Name != "" {
		mfr.AllocationName = &filterOpts.Name
	}
	if filterOpts.Project != "" {
		mfr.AllocationProject = &filterOpts.Project
	}
	if filterOpts.Image != "" {
		mfr.AllocationImageID = &filterOpts.Image
	}
	if filterOpts.Hostname != "" {
		mfr.AllocationHostname = &filterOpts.Hostname
	}
	if filterOpts.Hostname != "" {
		mfr.AllocationHostname = &filterOpts.Hostname
	}
	if filterOpts.Mac != "" {
		mfr.NicsMacAddresses = []string{filterOpts.Mac}
	}
	if len(filterOpts.Tags) > 0 {
		mfr.Tags = filterOpts.Tags
	}
	resp, err := driver.MachineIPMIList(mfr)
	if err != nil {
		return err
	}
	return printer.Print(resp.Machines)
}

func machineIssues(driver *metalgo.Driver) error {
	resp, err := driver.MachineList()
	if err != nil {
		return err
	}
	res := make(MachineIssues)

	asnMap := make(map[int64][]models.V1MachineResponse)
	for _, m := range resp.Machines {
		var issues []string

		if m.Partition == nil {
			issues = append(issues, "machine with no partition")
		}

		if m.Liveliness != nil && *m.Liveliness != "Alive" {
			issues = append(issues, "machine not alive")
		}

		if m.Allocation == nil && len(m.Events.Log) > 0 && *m.Events.Log[0].Event == "Phoned Home" {
			issues = append(issues, "machine phones home but not allocated")
		}

		if m.Events.IncompleteProvisioningCycles != nil &&
			*m.Events.IncompleteProvisioningCycles != "" &&
			*m.Events.IncompleteProvisioningCycles != "0" {
			if m.Events != nil && len(m.Events.Log) > 0 && *m.Events.Log[0].Event == "Waiting" {
				// Machine which are waiting are not considered to have issues
			} else {
				issues = append(issues, "machine has incomplete cycles")
			}
		}

		if m.Allocation != nil {
			// collecting ASN overlaps
			for _, n := range m.Allocation.Networks {
				if n.Asn == nil {
					continue
				}

				machines, ok := asnMap[*n.Asn]
				if !ok {
					machines = []models.V1MachineResponse{}
				}

				alreadyContained := false
				for _, mm := range machines {
					if *mm.ID == *m.ID {
						alreadyContained = true
						break
					}
				}

				if alreadyContained {
					continue
				}

				machines = append(machines, *m)
				asnMap[*n.Asn] = machines
			}
		}

		if len(issues) > 0 {
			res[*m.ID] = MachineAndIssues{
				Machine: *m,
				Issues:  issues,
			}
		}
	}

	for asn, ms := range asnMap {
		if len(ms) > 1 {
			var sharedIDs []string
			for _, m := range ms {
				sharedIDs = append(sharedIDs, *m.ID)
			}

			for _, m := range ms {
				mWithIssues, ok := res[*m.ID]
				if !ok {
					mWithIssues = MachineAndIssues{
						Machine: m,
					}
				}
				mWithIssues.Issues = append(mWithIssues.Issues, fmt.Sprintf("ASN (%d) not unique, shared with %s", asn, sharedIDs))
				res[*m.ID] = mWithIssues
			}
		}
	}

	return printer.Print(res)
}

func getMachineID(args []string) (string, error) {
	if len(args) < 1 {
		return "", fmt.Errorf("no machine ID given")
	}

	machineID := args[0]
	_, err := driver.MachineGet(machineID)
	if err != nil {
		return "", err
	}
	return machineID, nil
}

func getMachine(args []string) (*models.V1MachineIPMIResponse, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("no machine ID given")
	}

	machineID := args[0]
	m, err := driver.MachineIPMIGet(machineID)
	if err != nil {
		return nil, err
	}
	return m.Machine, nil
}
