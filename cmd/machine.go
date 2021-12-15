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
	"github.com/metal-stack/metalctl/cmd/output"
	"github.com/metal-stack/metalctl/pkg/api"
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

var filterOpts = &FilterOpts{}

func newMachineCmd(c *config) *cobra.Command {

	machineCmd := &cobra.Command{
		Use:   "machine",
		Short: "manage machines",
		Long:  "metal machines are bare metal servers.",
	}

	machineCreateCmd := &cobra.Command{
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
			return c.machineCreate()
		},
		PreRun: bindPFlags,
	}

	machineListCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "list all machines",
		Long:    "list all machines with almost all properties in tabular form.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.machineList()
		},
		PreRun: bindPFlags,
	}

	machineDescribeCmd := &cobra.Command{
		Use:   "describe <machine ID>",
		Short: "describe a machine",
		Long:  "describe a machine in a very detailed form with all properties.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.machineDescribe(args)
		},
		PreRun:            bindPFlags,
		ValidArgsFunction: c.comp.MachineListCompletion,
	}

	machineConsolePasswordCmd := &cobra.Command{
		Use:   "consolepassword <machine ID>",
		Short: "fetch the consolepassword for a machine",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.machineConsolePassword(args)
		},
		PreRun:            bindPFlags,
		ValidArgsFunction: c.comp.MachineListCompletion,
	}

	machineDestroyCmd := &cobra.Command{
		Use:     "destroy <machine ID>",
		Aliases: []string{"delete", "rm"},
		Short:   "destroy a machine",
		Long: `destroy a machine and destroy all data stored on the local disks. Once destroyed it is back for usage by other projects.
A destroyed machine can not restored anymore`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.machineDestroy(args)
		},
		PreRun:            bindPFlags,
		ValidArgsFunction: c.comp.MachineListCompletion,
	}

	machinePowerCmd := &cobra.Command{
		Use:   "power",
		Short: "manage machine power",
	}

	machinePowerOnCmd := &cobra.Command{
		Use:   "on <machine ID>",
		Short: "power on a machine",
		Long:  "set the machine to power on state, if the machine already was on nothing happens.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.machinePowerOn(args)
		},
		PreRun:            bindPFlags,
		ValidArgsFunction: c.comp.MachineListCompletion,
	}

	machinePowerOffCmd := &cobra.Command{
		Use:   "off <machine ID>",
		Short: "power off a machine",
		Long: `set the machine to power off state, if the machine already was off nothing happens.
It will usually take some time to power off the machine, depending on the machine type.
Power on will therefore not work if the machine is in the powering off phase.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.machinePowerOff(args)
		},
		PreRun:            bindPFlags,
		ValidArgsFunction: c.comp.MachineListCompletion,
	}

	machinePowerResetCmd := &cobra.Command{
		Use:   "reset <machine ID>",
		Short: "power reset a machine",
		Long:  "(hard) reset the machine power.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.machinePowerReset(args)
		},
		PreRun:            bindPFlags,
		ValidArgsFunction: c.comp.MachineListCompletion,
	}

	machinePowerCycleCmd := &cobra.Command{
		Use:   "cycle <machine ID>",
		Short: "power cycle a machine",
		Long:  "(soft) cycle the machine power.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.machinePowerCycle(args)
		},
		PreRun:            bindPFlags,
		ValidArgsFunction: c.comp.MachineListCompletion,
	}

	machineUpdateCmd := &cobra.Command{
		Use:     "update",
		Aliases: []string{"firmware-update"},
		Short:   "update a machine firmware",
	}

	machineUpdateBiosCmd := &cobra.Command{
		Use:   "bios <machine ID>",
		Short: "update a machine BIOS",
		Long:  "the machine BIOS will be updated to given revision. If revision flag is not specified an update plan will be printed instead.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.machineUpdateBios(args)
		},
		PreRun:            bindPFlags,
		ValidArgsFunction: c.comp.MachineListCompletion,
	}

	machineUpdateBmcCmd := &cobra.Command{
		Use:   "bmc <machine ID>",
		Short: "update a machine BMC",
		Long:  "the machine BMC will be updated to given revision. If revision flag is not specified an update plan will be printed instead.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.machineUpdateBmc(args)
		},
		PreRun:            bindPFlags,
		ValidArgsFunction: c.comp.MachineListCompletion,
	}

	machineBootBiosCmd := &cobra.Command{
		Use:   "bios <machine ID>",
		Short: "boot a machine into BIOS",
		Long:  "the machine will boot into bios.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.machineBootBios(args)
		},
		PreRun:            bindPFlags,
		ValidArgsFunction: c.comp.MachineListCompletion,
	}

	machineBootPxeCmd := &cobra.Command{
		Use:   "pxe <machine ID>",
		Short: "boot a machine from PXE",
		Long:  "the machine will boot from PXE.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.machineBootPxe(args)
		},
		PreRun:            bindPFlags,
		ValidArgsFunction: c.comp.MachineListCompletion,
	}

	machineBootDiskCmd := &cobra.Command{
		Use:   "disk <machine ID>",
		Short: "boot a machine from disk",
		Long:  "the machine will boot from disk.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.machineBootDisk(args)
		},
		PreRun:            bindPFlags,
		ValidArgsFunction: c.comp.MachineListCompletion,
	}

	machineIdentifyCmd := &cobra.Command{
		Use:   "identify",
		Short: "manage machine chassis identify LED power",
	}

	machineIdentifyOnCmd := &cobra.Command{
		Use:   "on <machine ID>",
		Short: "power on the machine chassis identify LED",
		Long:  `set the machine chassis identify LED to on state`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.machineIdentifyOn(args)
		},
		PreRun:            bindPFlags,
		ValidArgsFunction: c.comp.MachineListCompletion,
	}

	machineIdentifyOffCmd := &cobra.Command{
		Use:   "off <machine ID>",
		Short: "power off the machine chassis identify LED",
		Long:  `set the machine chassis identify LED to off state`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.machineIdentifyOff(args)
		},
		PreRun:            bindPFlags,
		ValidArgsFunction: c.comp.MachineListCompletion,
	}

	machineReserveCmd := &cobra.Command{
		Use:   "reserve <machine ID>",
		Short: "reserve a machine",
		Long: `reserve a machine for exclusive usage, this machine will no longer be picked by other allocations.
This is useful for maintenance of the machine or testing. After the reservation is not needed anymore, the reservation
should be removed with --remove.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.machineReserve(args)
		},
		PreRun:            bindPFlags,
		ValidArgsFunction: c.comp.MachineListCompletion,
	}

	machineLockCmd := &cobra.Command{
		Use:   "lock <machine ID>",
		Short: "lock a machine",
		Long:  `when a machine is locked, it can not be destroyed, to destroy a machine you must first remove the lock from that machine with --remove`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.machineLock(args)
		},
		PreRun:            bindPFlags,
		ValidArgsFunction: c.comp.MachineListCompletion,
	}

	machineReinstallCmd := &cobra.Command{
		Use:   "reinstall <machine ID>",
		Short: "reinstalls an already allocated machine",
		Long: `reinstalls an already allocated machine. If it is not yet allocated, nothing happens, otherwise only the machine's primary disk
is wiped and the new image will subsequently be installed on that device`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.machineReinstall(args)
		},
		PreRun:            bindPFlags,
		ValidArgsFunction: c.comp.MachineListCompletion,
	}

	machineConsoleCmd := &cobra.Command{
		Use: "console <machine ID>",
		Short: `console access to a machine, machine must be created with a ssh public key, authentication is done with your private key.
In case the machine did not register properly a direct ipmi console access is available via the --ipmi flag. This is only for administrative access.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.machineConsole(args)
		},
		PreRun:            bindPFlags,
		ValidArgsFunction: c.comp.MachineListCompletion,
	}
	machineIpmiCmd := &cobra.Command{
		Use:   "ipmi [<machine ID>]",
		Short: `display ipmi details of the machine, if no machine ID is given all ipmi addresses are returned.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.machineIpmi(args)
		},
		PreRun: bindPFlags,
	}
	machineIssuesCmd := &cobra.Command{
		Use:   "issues",
		Short: `display machines which are in a potential bad state`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.machineIssues()
		},
		PreRun: bindPFlags,
	}
	machineLogsCmd := &cobra.Command{
		Use:     "logs <machine ID>",
		Aliases: []string{"log"},
		Short:   `display machine provisioning logs`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.machineLogs(args)
		},
		PreRun:            bindPFlags,
		ValidArgsFunction: c.comp.MachineListCompletion,
	}

	c.addMachineCreateFlags(machineCreateCmd, "machine")
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

	must(machineListCmd.RegisterFlagCompletionFunc("partition", c.comp.PartitionListCompletion))
	must(machineListCmd.RegisterFlagCompletionFunc("size", c.comp.SizeListCompletion))
	must(machineListCmd.RegisterFlagCompletionFunc("project", c.comp.ProjectListCompletion))
	must(machineListCmd.RegisterFlagCompletionFunc("id", c.comp.MachineListCompletion))
	must(machineListCmd.RegisterFlagCompletionFunc("image", c.comp.ImageListCompletion))

	machineIpmiCmd.Flags().StringVarP(&filterOpts.ID, "id", "", "", "ID to filter [optional]")
	machineIpmiCmd.Flags().StringVarP(&filterOpts.Partition, "partition", "", "", "partition to filter [optional]")
	machineIpmiCmd.Flags().StringVarP(&filterOpts.Size, "size", "", "", "size to filter [optional]")
	machineIpmiCmd.Flags().StringVarP(&filterOpts.Name, "name", "", "", "allocation name to filter [optional]")
	machineIpmiCmd.Flags().StringVarP(&filterOpts.Project, "project", "", "", "allocation project to filter [optional]")
	machineIpmiCmd.Flags().StringVarP(&filterOpts.Image, "image", "", "", "allocation image to filter [optional]")
	machineIpmiCmd.Flags().StringVarP(&filterOpts.Hostname, "hostname", "", "", "allocation hostname to filter [optional]")
	machineIpmiCmd.Flags().StringVarP(&filterOpts.Mac, "mac", "", "", "mac to filter [optional]")
	machineIpmiCmd.Flags().StringSliceVar(&filterOpts.Tags, "tags", []string{}, "tags to filter, use it like: --tags \"tag1,tag2\" or --tags \"tag3\".")

	must(machineIpmiCmd.RegisterFlagCompletionFunc("partition", c.comp.PartitionListCompletion))
	must(machineIpmiCmd.RegisterFlagCompletionFunc("size", c.comp.SizeListCompletion))
	must(machineIpmiCmd.RegisterFlagCompletionFunc("project", c.comp.ProjectListCompletion))
	must(machineIpmiCmd.RegisterFlagCompletionFunc("id", c.comp.MachineListCompletion))

	machineConsolePasswordCmd.Flags().StringP("reason", "", "", "a short description why access to the consolepassword is required")

	machineCmd.AddCommand(machineListCmd)
	machineCmd.AddCommand(machineDestroyCmd)
	machineCmd.AddCommand(machineDescribeCmd)
	machineCmd.AddCommand(machineConsolePasswordCmd)

	machineUpdateBiosCmd.Flags().StringP("revision", "", "", "the BIOS revision")
	machineUpdateBiosCmd.Flags().StringP("description", "", "", "the reason why the BIOS should be updated")
	must(machineUpdateBiosCmd.RegisterFlagCompletionFunc("revision", c.comp.FirmwareBiosRevisionCompletion))
	machineUpdateCmd.AddCommand(machineUpdateBiosCmd)

	machineUpdateBmcCmd.Flags().StringP("revision", "", "", "the BMC revision")
	machineUpdateBmcCmd.Flags().StringP("description", "", "", "the reason why the BMC should be updated")
	must(machineUpdateBmcCmd.RegisterFlagCompletionFunc("revision", c.comp.FirmwareBmcRevisionCompletion))
	machineUpdateCmd.AddCommand(machineUpdateBmcCmd)

	machineCmd.AddCommand(machineUpdateCmd)

	machinePowerCmd.AddCommand(machinePowerOnCmd)
	machinePowerCmd.AddCommand(machinePowerOffCmd)
	machinePowerCmd.AddCommand(machinePowerResetCmd)
	machinePowerCmd.AddCommand(machinePowerCycleCmd)
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
	must(machineReinstallCmd.MarkFlagRequired("image"))
	machineCmd.AddCommand(machineReinstallCmd)

	machineIssuesCmd.Flags().StringSliceP("only", "", []string{}, "issue types to include [optional]")
	machineIssuesCmd.Flags().StringSliceP("omit", "", []string{}, "issue types to omit [optional]")
	machineIssuesCmd.Flags().StringVarP(&filterOpts.ID, "id", "", "", "ID to filter [optional]")
	machineIssuesCmd.Flags().StringVarP(&filterOpts.Partition, "partition", "", "", "partition to filter [optional]")
	machineIssuesCmd.Flags().StringVarP(&filterOpts.Size, "size", "", "", "size to filter [optional]")
	machineIssuesCmd.Flags().StringVarP(&filterOpts.Name, "name", "", "", "allocation name to filter [optional]")
	machineIssuesCmd.Flags().StringVarP(&filterOpts.Project, "project", "", "", "allocation project to filter [optional]")
	machineIssuesCmd.Flags().StringVarP(&filterOpts.Image, "image", "", "", "allocation image to filter [optional]")
	machineIssuesCmd.Flags().StringVarP(&filterOpts.Hostname, "hostname", "", "", "allocation hostname to filter [optional]")
	machineIssuesCmd.Flags().StringVarP(&filterOpts.Mac, "mac", "", "", "mac to filter [optional]")
	machineIssuesCmd.Flags().StringSliceVar(&filterOpts.Tags, "tags", []string{}, "tags to filter, use it like: --tags \"tag1,tag2\" or --tags \"tag3\".")

	must(machineIssuesCmd.RegisterFlagCompletionFunc("partition", c.comp.PartitionListCompletion))
	must(machineIssuesCmd.RegisterFlagCompletionFunc("size", c.comp.SizeListCompletion))
	must(machineIssuesCmd.RegisterFlagCompletionFunc("project", c.comp.ProjectListCompletion))
	must(machineIssuesCmd.RegisterFlagCompletionFunc("id", c.comp.MachineListCompletion))
	must(machineIssuesCmd.RegisterFlagCompletionFunc("image", c.comp.ImageListCompletion))
	must(machineIssuesCmd.RegisterFlagCompletionFunc("omit", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		var shortNames []string
		for _, i := range api.AllIssues {
			shortNames = append(shortNames, i.ShortName+"\t"+i.Description)
		}
		return shortNames, cobra.ShellCompDirectiveNoFileComp
	}))
	must(machineIssuesCmd.RegisterFlagCompletionFunc("only", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		var shortNames []string
		for _, i := range api.AllIssues {
			shortNames = append(shortNames, i.ShortName+"\t"+i.Description)
		}
		return shortNames, cobra.ShellCompDirectiveNoFileComp
	}))

	machineConsoleCmd.Flags().StringP("sshidentity", "p", "", "SSH key file, if not given the default ssh key will be used if present [optional].")
	machineConsoleCmd.Flags().BoolP("ipmi", "", false, "use ipmitool with direct network access (admin only).")
	machineConsoleCmd.Flags().StringP("ipmiuser", "", "", "overwrite ipmi user (admin only).")
	machineConsoleCmd.Flags().StringP("ipmipassword", "", "", "overwrite ipmi password (admin only).")
	machineCmd.AddCommand(machineConsoleCmd)
	machineCmd.AddCommand(machineIpmiCmd)
	machineCmd.AddCommand(machineIssuesCmd)
	machineCmd.AddCommand(machineLogsCmd)

	machineDestroyCmd.Flags().Bool("remove-from-database", false, "remove given machine from the database, is only required for maintenance reasons [optional] (admin only).")

	return machineCmd
}

func (c *config) addMachineCreateFlags(cmd *cobra.Command, name string) {
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

	must(cmd.MarkFlagRequired("hostname"))
	must(cmd.MarkFlagRequired("image"))
	must(cmd.MarkFlagRequired("project"))

	// Completion for arguments
	must(cmd.RegisterFlagCompletionFunc("networks", c.comp.NetworkListCompletion))
	must(cmd.RegisterFlagCompletionFunc("ips", c.comp.IpListCompletion))
	must(cmd.RegisterFlagCompletionFunc("partition", c.comp.PartitionListCompletion))
	must(cmd.RegisterFlagCompletionFunc("size", c.comp.SizeListCompletion))
	must(cmd.RegisterFlagCompletionFunc("project", c.comp.ProjectListCompletion))
	must(cmd.RegisterFlagCompletionFunc("id", c.comp.MachineListCompletion))
	must(cmd.RegisterFlagCompletionFunc("image", c.comp.ImageListCompletion))
	must(cmd.RegisterFlagCompletionFunc("filesystemlayout", c.comp.FilesystemLayoutListCompletion))
}

func (c *config) machineCreate() error {
	mcr, err := c.machineCreateRequest()
	if err != nil {
		return fmt.Errorf("machine create error:%w", err)
	}

	resp, err := c.driver.MachineCreate(mcr)
	if err != nil {
		return fmt.Errorf("machine create error:%w", err)
	}
	return output.New().Print(resp.Machine)
}

func (c *config) machineCreateRequest() (*metalgo.MachineCreateRequest, error) {
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

func (c *config) machineList() error {
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
		resp, err = c.driver.MachineFind(mfr)
	} else {
		resp, err = c.driver.MachineList()
	}
	if err != nil {
		return err
	}
	return output.New().Print(resp.Machines)
}

func (c *config) machineDescribe(args []string) error {
	machineID, err := c.getMachineID(args)
	if err != nil {
		return err
	}
	resp, err := c.driver.MachineGet(machineID)
	if err != nil {
		return err
	}
	return output.NewDetailer().Detail(resp.Machine)
}

func (c *config) machineConsolePassword(args []string) error {
	machineID, err := c.getMachineID(args)
	if err != nil {
		return err
	}
	reason := viper.GetString("reason")
	resp, err := c.driver.MachineConsolePassword(machineID, reason)
	if err != nil {
		return err
	}
	fmt.Printf("%s\n", *resp.ConsolePassword)
	return nil
}

func (c *config) machineDestroy(args []string) error {
	machineID, err := c.getMachineID(args)
	if err != nil {
		return err
	}

	if viper.GetBool("remove-from-database") {
		if !viper.GetBool(forceFlag) {
			return fmt.Errorf("remove-from-database is set but you forgot to add --%s", forceFlag)
		}
		resp, err := c.driver.MachineDeleteFromDatabase(machineID)
		if err != nil {
			return err
		}
		return output.New().Print(resp.Machine)
	}

	resp, err := c.driver.MachineDelete(machineID)
	if err != nil {
		return err
	}
	return output.New().Print(resp.Machine)
}

func (c *config) machinePowerOn(args []string) error {
	machineID, err := c.getMachineID(args)
	if err != nil {
		return err
	}

	resp, err := c.driver.MachinePowerOn(machineID)
	if err != nil {
		return err
	}
	return output.New().Print(resp.Machine)
}

func (c *config) machinePowerOff(args []string) error {
	machineID, err := c.getMachineID(args)
	if err != nil {
		return err
	}

	resp, err := c.driver.MachinePowerOff(machineID)
	if err != nil {
		return err
	}
	return output.New().Print(resp.Machine)
}

func (c *config) machinePowerReset(args []string) error {
	machineID, err := c.getMachineID(args)
	if err != nil {
		return err
	}

	resp, err := c.driver.MachinePowerReset(machineID)
	if err != nil {
		return err
	}
	return output.New().Print(resp.Machine)
}

func (c *config) machinePowerCycle(args []string) error {
	machineID, err := c.getMachineID(args)
	if err != nil {
		return err
	}

	resp, err := c.driver.MachinePowerCycle(machineID)
	if err != nil {
		return err
	}
	return output.New().Print(resp.Machine)
}

func (c *config) machineUpdateBios(args []string) error {
	m, vendor, board, err := c.firmwareData(args)
	if err != nil {
		return err
	}
	revision := viper.GetString("revision")
	currentVersion := ""
	if m.Bios != nil && m.Bios.Version != nil {
		currentVersion = *m.Bios.Version
	}

	return c.machineUpdateFirmware(metalgo.Bios, *m.ID, vendor, board, revision, currentVersion)
}

func (c *config) machineUpdateBmc(args []string) error {
	m, vendor, board, err := c.firmwareData(args)
	if err != nil {
		return err
	}
	revision := viper.GetString("revision")
	currentVersion := ""
	if m.Ipmi != nil && m.Ipmi.Bmcversion != nil {
		currentVersion = *m.Ipmi.Bmcversion
	}

	return c.machineUpdateFirmware(metalgo.Bmc, *m.ID, vendor, board, revision, currentVersion)
}

func (c *config) firmwareData(args []string) (*models.V1MachineIPMIResponse, string, string, error) {
	m, err := c.getMachine(args)
	if err != nil {
		return nil, "", "", err
	}
	machineID := *m.ID
	if m.Ipmi == nil {
		return nil, "", "", fmt.Errorf("no ipmi data available of machine %s", machineID)
	}

	fru := *m.Ipmi.Fru
	vendor := strings.ToLower(fru.BoardMfg)
	board := strings.ToUpper(fru.BoardPartNumber)

	return m, vendor, board, nil
}

func (c *config) machineUpdateFirmware(kind metalgo.FirmwareKind, machineID, vendor, board, revision, currentVersion string) error {
	f, err := c.driver.ListFirmwares(kind, "", "")
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

	switch kind {
	case metalgo.Bios:
		fmt.Println("It is recommended to power off the machine before updating the BIOS. This command will power on your machine automatically after the update or trigger a reboot.\n\nThe update may take a couple of minutes (up to ~10 minutes). Please wait until the machine powers on / reboots automatically as otherwise the update is still progressing or an error occurred during the update.")
	case metalgo.Bmc:
		fmt.Println("The update may take a couple of minutes (up to ~10 minutes). You can look up the result through the server's BMC interface.")
	}

	if !viper.GetBool("yes-i-really-mean-it") {
		err = Prompt("Do you want to continue? (y/n)", "y")
		if err != nil {
			return err
		}
	}

	description := viper.GetString("description")
	if description == "" {
		description = "unknown"
	}

	resp, err := c.driver.MachineUpdateFirmware(kind, machineID, revision, description)
	if err != nil {
		return err
	}
	return output.New().Print(resp.Machine)
}

func (c *config) machineBootBios(args []string) error {
	machineID, err := c.getMachineID(args)
	if err != nil {
		return err
	}

	resp, err := c.driver.MachineBootBios(machineID)
	if err != nil {
		return err
	}
	return output.New().Print(resp.Machine)
}

func (c *config) machineBootDisk(args []string) error {
	machineID, err := c.getMachineID(args)
	if err != nil {
		return err
	}

	resp, err := c.driver.MachineBootDisk(machineID)
	if err != nil {
		return err
	}
	return output.New().Print(resp.Machine)
}

func (c *config) machineBootPxe(args []string) error {
	machineID, err := c.getMachineID(args)
	if err != nil {
		return err
	}

	resp, err := c.driver.MachineBootPxe(machineID)
	if err != nil {
		return err
	}
	return output.New().Print(resp.Machine)
}

func (c *config) machineIdentifyOn(args []string) error {
	machineID, err := c.getMachineID(args)
	if err != nil {
		return err
	}

	description := viper.GetString("description")
	resp, err := c.driver.ChassisIdentifyLEDPowerOn(machineID, description)
	if err != nil {
		return err
	}
	return output.New().Print(resp.Machine)
}

func (c *config) machineIdentifyOff(args []string) error {
	machineID, err := c.getMachineID(args)
	if err != nil {
		return err
	}

	description := viper.GetString("description")
	resp, err := c.driver.ChassisIdentifyLEDPowerOff(machineID, description)
	if err != nil {
		return err
	}
	return output.New().Print(resp.Machine)
}

func (c *config) machineReserve(args []string) error {
	machineID, err := c.getMachineID(args)
	if err != nil {
		return err
	}
	description := viper.GetString("description")
	remove := viper.GetBool("remove")

	var resp *metalgo.MachineStateResponse
	if remove {
		resp, err = c.driver.MachineUnReserve(machineID)
		if err != nil {
			return err
		}
	} else {
		resp, err = c.driver.MachineReserve(machineID, description)
		if err != nil {
			return err
		}
	}
	return output.New().Print(resp.Machine)
}

func (c *config) machineLock(args []string) error {
	machineID, err := c.getMachineID(args)
	if err != nil {
		return err
	}
	description := viper.GetString("description")
	remove := viper.GetBool("remove")

	var resp *metalgo.MachineStateResponse
	if remove {
		resp, err = c.driver.MachineUnLock(machineID)
		if err != nil {
			return err
		}
	} else {
		resp, err = c.driver.MachineLock(machineID, description)
		if err != nil {
			return err
		}
	}
	return output.New().Print(resp.Machine)
}

func (c *config) machineReinstall(args []string) error {
	machineID, err := c.getMachineID(args)
	if err != nil {
		return err
	}
	imageID := viper.GetString("image")
	description := viper.GetString("description")

	var resp *metalgo.MachineGetResponse
	resp, err = c.driver.MachineReinstall(machineID, imageID, description)
	if err != nil {
		return err
	}
	return output.New().Print(resp.Machine)
}

func (c *config) machineLogs(args []string) error {
	// FIXME add ipmi sel as well
	machineID, err := c.getMachineID(args)
	if err != nil {
		return err
	}

	resp, err := c.driver.MachineGet(machineID)
	if err != nil {
		return err
	}
	machine := resp.Machine
	return output.New().Print(machine.Events.Log)
}

func (c *config) machineConsole(args []string) error {
	machineID, err := c.getMachineID(args)
	if err != nil {
		return err
	}
	useIpmi := viper.GetBool("ipmi")
	if useIpmi {
		path, err := exec.LookPath("ipmitool")
		if err != nil {
			return fmt.Errorf("unable to locate ipmitool in path")
		}

		resp, err := c.driver.MachineIPMIGet(machineID)
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
	parsedurl, err := url.Parse(c.driverURL)
	if err != nil {
		return err
	}
	authContext, err := getAuthContext(viper.GetString("kubeconfig"))
	if err != nil {
		return err
	}
	err = os.Setenv("LC_METAL_STACK_OIDC_TOKEN", authContext.IDToken)
	if err != nil {
		return err
	}
	err = SSHClient(machineID, key, parsedurl.Host, bmcConsolePort)
	if err != nil {
		return fmt.Errorf("machine console error:%w", err)
	}
	return nil
}

func (c *config) machineIpmi(args []string) error {
	if len(args) == 1 {
		machineID, err := c.getMachineID(args)
		if err != nil {
			return err
		}
		resp, err := c.driver.MachineIPMIGet(machineID)
		if err != nil {
			return err
		}

		hidden := "<hidden>"
		resp.Machine.Ipmi.Password = &hidden
		return output.NewDetailer().Detail(resp.Machine)
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
	if filterOpts.Mac != "" {
		mfr.NicsMacAddresses = []string{filterOpts.Mac}
	}
	if len(filterOpts.Tags) > 0 {
		mfr.Tags = filterOpts.Tags
	}
	resp, err := c.driver.MachineIPMIList(mfr)
	if err != nil {
		return err
	}
	return output.New().Print(resp.Machines)
}

func (c *config) machineIssues() error {
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
	resp, err := c.driver.MachineIPMIList(mfr)
	if err != nil {
		return err
	}

	only := viper.GetStringSlice("only")
	omit := viper.GetStringSlice("omit")

	var (
		res      = api.MachineIssues{}
		asnMap   = map[int64][]models.V1MachineIPMIResponse{}
		bmcIPMap = map[string][]models.V1MachineIPMIResponse{}

		conditionalAppend = func(issues api.Issues, issue api.Issue) api.Issues {
			for _, o := range omit {
				if issue.ShortName == o {
					return issues
				}
			}

			if len(only) > 0 {
				for _, o := range only {
					if issue.ShortName == o {
						return append(issues, issue)
					}
				}
				return issues
			}

			return append(issues, issue)
		}
	)

	for _, m := range resp.Machines {
		var issues api.Issues

		if m.Partition == nil {
			issues = conditionalAppend(issues, api.IssueNoPartition)
		}

		if m.Liveliness != nil {
			switch *m.Liveliness {
			case "Alive":
			case "Dead":
				issues = conditionalAppend(issues, api.IssueLivelinessDead)
			case "Unknown":
				issues = conditionalAppend(issues, api.IssueLivelinessUnknown)
			default:
				issues = conditionalAppend(issues, api.IssueLivelinessNotAvailable)
			}
		} else {
			issues = conditionalAppend(issues, api.IssueLivelinessNotAvailable)
		}

		if m.Allocation == nil && len(m.Events.Log) > 0 && *m.Events.Log[0].Event == "Phoned Home" {
			issues = conditionalAppend(issues, api.IssueFailedMachineReclaim)
		}

		if m.Events.IncompleteProvisioningCycles != nil &&
			*m.Events.IncompleteProvisioningCycles != "" &&
			*m.Events.IncompleteProvisioningCycles != "0" {
			if m.Events != nil && len(m.Events.Log) > 0 && *m.Events.Log[0].Event == "Waiting" {
				// Machine which are waiting are not considered to have issues
			} else {
				issues = conditionalAppend(issues, api.IssueIncompleteCycles)
			}
		}

		if m.Ipmi != nil {
			if m.Ipmi.Mac == nil || *m.Ipmi.Mac == "" {
				issues = conditionalAppend(issues, api.IssueBMCWithoutMAC)
			}

			if m.Ipmi.Address == nil || *m.Ipmi.Address == "" {
				issues = conditionalAppend(issues, api.IssueBMCWithoutIP)
			} else {
				entries := bmcIPMap[*m.Ipmi.Address]
				entries = append(entries, *m)
				bmcIPMap[*m.Ipmi.Address] = entries
			}
		}

		if m.Allocation != nil && m.Allocation.Role != nil && *m.Allocation.Role == models.V1MachineAllocationRoleFirewall {
			// collecting ASN overlaps
			for _, n := range m.Allocation.Networks {
				if n.Asn == nil {
					continue
				}

				machines, ok := asnMap[*n.Asn]
				if !ok {
					machines = []models.V1MachineIPMIResponse{}
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
			res[*m.ID] = api.MachineWithIssues{
				Machine: *m,
				Issues:  issues,
			}
		}
	}

	includeASN := true
	for _, o := range omit {
		if o == api.IssueASNUniqueness.ShortName {
			includeASN = false
			break
		}
	}

	if includeASN {
		for asn, ms := range asnMap {
			if len(ms) < 2 {
				continue
			}

			for _, m := range ms {
				var sharedIDs []string
				for _, mm := range ms {
					if *m.ID == *mm.ID {
						continue
					}
					sharedIDs = append(sharedIDs, *mm.ID)
				}

				mWithIssues, ok := res[*m.ID]
				if !ok {
					mWithIssues = api.MachineWithIssues{
						Machine: m,
					}
				}
				issue := api.IssueASNUniqueness
				issue.Description = fmt.Sprintf("ASN (%d) not unique, shared with %s", asn, sharedIDs)
				mWithIssues.Issues = append(mWithIssues.Issues, issue)
				res[*m.ID] = mWithIssues
			}
		}
	}

	includeDistinctBMC := true
	for _, o := range omit {
		if o == api.IssueNonDistinctBMCIP.ShortName {
			includeDistinctBMC = false
			break
		}
	}

	if includeDistinctBMC {
		for ip, ms := range bmcIPMap {
			if len(ms) < 2 {
				continue
			}

			for _, m := range ms {
				var sharedIDs []string
				for _, mm := range ms {
					if *m.ID == *mm.ID {
						continue
					}
					sharedIDs = append(sharedIDs, *mm.ID)
				}

				mWithIssues, ok := res[*m.ID]
				if !ok {
					mWithIssues = api.MachineWithIssues{
						Machine: m,
					}
				}
				issue := api.IssueNonDistinctBMCIP
				issue.Description = fmt.Sprintf("BMC IP (%s) not unique, shared with %s", ip, sharedIDs)
				mWithIssues.Issues = append(mWithIssues.Issues, issue)
				res[*m.ID] = mWithIssues
			}
		}
	}

	return output.New().Print(res)
}

func (c *config) getMachineID(args []string) (string, error) {
	if len(args) < 1 {
		return "", fmt.Errorf("no machine ID given")
	}

	machineID := args[0]
	_, err := c.driver.MachineGet(machineID)
	if err != nil {
		return "", err
	}
	return machineID, nil
}

func (c *config) getMachine(args []string) (*models.V1MachineIPMIResponse, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("no machine ID given")
	}

	machineID := args[0]
	m, err := c.driver.MachineIPMIGet(machineID)
	if err != nil {
		return nil, err
	}
	return m.Machine, nil
}
