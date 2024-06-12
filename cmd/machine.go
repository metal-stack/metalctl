package cmd

import (
	"encoding/base64"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"fmt"

	"net/url"

	"slices"

	"github.com/metal-stack/metal-go/api/client/firmware"
	"github.com/metal-stack/metal-go/api/client/machine"
	"github.com/metal-stack/metal-go/api/models"
	"github.com/metal-stack/metal-lib/pkg/genericcli"
	"github.com/metal-stack/metal-lib/pkg/genericcli/printers"
	"github.com/metal-stack/metal-lib/pkg/pointer"
	"github.com/metal-stack/metalctl/cmd/sorters"
	"github.com/metal-stack/metalctl/cmd/tableprinters"
	"github.com/metal-stack/metalctl/pkg/api"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	// Port open on our control-plane to connect via ssh to get machine console access.
	bmcConsolePort = 5222
	forceFlag      = "yes-i-really-mean-it"
)

type machineCmd struct {
	*config
}

func (c *machineCmd) listCmdFlags(cmd *cobra.Command, lastEventErrorThresholdDefault time.Duration) {
	listFlagCompletions := []struct {
		flagName string
		f        func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective)
	}{
		{flagName: "partition", f: c.comp.PartitionListCompletion},
		{flagName: "size", f: c.comp.SizeListCompletion},
		{flagName: "project", f: c.comp.ProjectListCompletion},
		{flagName: "rack", f: c.comp.MachineRackListCompletion},
		{flagName: "id", f: c.comp.MachineListCompletion},
		{flagName: "image", f: c.comp.ImageListCompletion},
		{flagName: "state", f: cobra.FixedCompletions([]string{
			// empty does not work:
			// models.V1FirewallFindRequestStateValueEmpty,
			models.V1FirewallFindRequestStateValueLOCKED,
			models.V1MachineFindRequestStateValueRESERVED,
		}, cobra.ShellCompDirectiveDefault)},
		{flagName: "role", f: cobra.FixedCompletions([]string{
			models.V1MachineAllocationRoleFirewall,
			models.V1MachineAllocationRoleMachine,
		}, cobra.ShellCompDirectiveDefault)},
		{flagName: "manufacturer", f: c.comp.MachineManufacturerCompletion},
		{flagName: "product-part-number", f: c.comp.MachineProductPartNumberCompletion},
		{flagName: "product-serial", f: c.comp.MachineProductSerialCompletion},
		{flagName: "board-part-number", f: c.comp.MachineBoardPartNumberCompletion},
		{flagName: "network-destination-prefixes", f: c.comp.NetworkDestinationPrefixesCompletion},
		{flagName: "network-ids", f: c.comp.NetworkListCompletion},
	}

	cmd.Flags().String("id", "", "ID to filter [optional]")
	cmd.Flags().String("partition", "", "partition to filter [optional]")
	cmd.Flags().String("size", "", "size to filter [optional]")
	cmd.Flags().String("rack", "", "rack to filter [optional]")
	cmd.Flags().String("state", "", "state to filter [optional]")
	cmd.Flags().String("name", "", "allocation name to filter [optional]")
	cmd.Flags().String("project", "", "allocation project to filter [optional]")
	cmd.Flags().String("image", "", "allocation image to filter [optional]")
	cmd.Flags().String("hostname", "", "allocation hostname to filter [optional]")
	cmd.Flags().String("mac", "", "mac to filter [optional]")
	cmd.Flags().StringSlice("tags", []string{}, "tags to filter, use it like: --tags \"tag1,tag2\" or --tags \"tag3\".")
	cmd.Flags().Duration("last-event-error-threshold", lastEventErrorThresholdDefault, "the duration up to how long in the past a machine last event error will be counted as an issue [optional]")
	cmd.Flags().String("role", "", "allocation role to filter [optional]")
	cmd.Flags().String("board-part-number", "", "fru board part number to filter [optional]")
	cmd.Flags().String("manufacturer", "", "fru manufacturer to filter [optional]")
	cmd.Flags().String("product-part-number", "", "fru product part number to filter [optional]")
	cmd.Flags().String("product-serial", "", "fru product serial to filter [optional]")
	cmd.Flags().String("bmc-address", "", "bmc ipmi address (needs to include port) to filter [optional]")
	cmd.Flags().String("bmc-mac", "", "bmc mac address to filter [optional]")
	cmd.Flags().String("network-destination-prefixes", "", "network destination prefixes to filter [optional]")
	cmd.Flags().String("network-ids", "", "network ids to filter [optional]")
	cmd.Flags().String("network-ips", "", "network ips to filter [optional]")

	for _, c := range listFlagCompletions {
		c := c
		genericcli.Must(cmd.RegisterFlagCompletionFunc(c.flagName, c.f))
	}

	cmd.Long = cmd.Short + "\n" + api.EmojiHelpText()
}

func newMachineCmd(c *config) *cobra.Command {
	w := machineCmd{
		config: c,
	}

	cmdsConfig := &genericcli.CmdsConfig[*models.V1MachineAllocateRequest, *models.V1MachineUpdateRequest, *models.V1MachineResponse]{
		BinaryName:           binaryName,
		GenericCLI:           genericcli.NewGenericCLI[*models.V1MachineAllocateRequest, *models.V1MachineUpdateRequest, *models.V1MachineResponse](w).WithFS(c.fs),
		Singular:             "machine",
		Plural:               "machines",
		Description:          "a machine is a bare metal server provisioned through metal-stack that is intended to run user workload.",
		Aliases:              []string{"ms"},
		CreateRequestFromCLI: w.createRequestFromCLI,
		UpdateRequestFromCLI: w.updateRequestFromCLI,
		Sorter:               sorters.MachineSorter(),
		ValidArgsFn:          c.comp.MachineListCompletion,
		DescribePrinter:      func() printers.Printer { return c.describePrinter },
		ListPrinter:          func() printers.Printer { return c.listPrinter },
		CreateCmdMutateFn: func(cmd *cobra.Command) {
			c.addMachineCreateFlags(cmd, "machine")
			cmd.Aliases = []string{"allocate"}
			cmd.Example = `machine create can be done in two different ways:

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
`
		},
		DeleteCmdMutateFn: func(cmd *cobra.Command) {
			cmd.Long = `delete a machine and destroy all data stored on the local disks. Once destroyed it is back for usage by other projects. A destroyed machine can not restored anymore`
			cmd.Flags().Bool("remove-from-database", false, "remove given machine from the database, is only required for maintenance reasons [optional] (admin only).")
		},
		ListCmdMutateFn: func(cmd *cobra.Command) {
			w.listCmdFlags(cmd, 1*time.Hour)
		},
		UpdateCmdMutateFn: func(cmd *cobra.Command) {
			cmd.Flags().String("description", "", "the description of the machine [optional]")
			cmd.Flags().StringSlice("add-tags", []string{}, "tags to be added to the machine [optional]")
			cmd.Flags().StringSlice("remove-tags", []string{}, "tags to be removed from the machine [optional]")
		},
	}

	machineConsolePasswordCmd := &cobra.Command{
		Use:   "consolepassword <machine ID>",
		Short: "fetch the consolepassword for a machine",
		RunE: func(cmd *cobra.Command, args []string) error {
			return w.machineConsolePassword(args)
		},
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
			return w.machinePowerOn(args)
		},
		ValidArgsFunction: c.comp.MachineListCompletion,
	}

	machinePowerOffCmd := &cobra.Command{
		Use:   "off <machine ID>",
		Short: "power off a machine",
		Long: `set the machine to power off state, if the machine already was off nothing happens.
It will usually take some time to power off the machine, depending on the machine type.
Power on will therefore not work if the machine is in the powering off phase.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return w.machinePowerOff(args)
		},
		ValidArgsFunction: c.comp.MachineListCompletion,
	}

	machinePowerResetCmd := &cobra.Command{
		Use:   "reset <machine ID>",
		Short: "power reset a machine",
		Long:  "(hard) reset the machine power.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return w.machinePowerReset(args)
		},
		ValidArgsFunction: c.comp.MachineListCompletion,
	}

	machinePowerCycleCmd := &cobra.Command{
		Use:   "cycle <machine ID>",
		Short: "power cycle a machine (graceful shutdown)",
		Long:  "(soft) cycle the machine power.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return w.machinePowerCycle(args)
		},
		ValidArgsFunction: c.comp.MachineListCompletion,
	}

	machineUpdateFirmwareCmd := &cobra.Command{
		Use:     "update-firmware",
		Aliases: []string{"firmware-update"},
		Short:   "update a machine firmware",
	}

	machineUpdateBiosCmd := &cobra.Command{
		Use:   "bios <machine ID>",
		Short: "update a machine BIOS",
		Long:  "the machine BIOS will be updated to given revision. If revision flag is not specified an update plan will be printed instead.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return w.machineUpdateBios(args)
		},
		ValidArgsFunction: c.comp.MachineListCompletion,
	}

	machineUpdateBmcCmd := &cobra.Command{
		Use:   "bmc <machine ID>",
		Short: "update a machine BMC",
		Long:  "the machine BMC will be updated to given revision. If revision flag is not specified an update plan will be printed instead.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return w.machineUpdateBmc(args)
		},
		ValidArgsFunction: c.comp.MachineListCompletion,
	}

	machineBootBiosCmd := &cobra.Command{
		Use:   "bios <machine ID>",
		Short: "boot a machine into BIOS",
		Long:  "the machine will boot into bios. (machine does not reboot automatically)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return w.machineBootBios(args)
		},
		ValidArgsFunction: c.comp.MachineListCompletion,
	}

	machineBootPxeCmd := &cobra.Command{
		Use:   "pxe <machine ID>",
		Short: "boot a machine from PXE",
		Long:  "the machine will boot from PXE. (machine does not reboot automatically)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return w.machineBootPxe(args)
		},
		ValidArgsFunction: c.comp.MachineListCompletion,
	}

	machineBootDiskCmd := &cobra.Command{
		Use:   "disk <machine ID>",
		Short: "boot a machine from disk",
		Long:  "the machine will boot from disk. (machine does not reboot automatically)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return w.machineBootDisk(args)
		},
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
			return w.machineIdentifyOn(args)
		},
		ValidArgsFunction: c.comp.MachineListCompletion,
	}

	machineIdentifyOffCmd := &cobra.Command{
		Use:   "off <machine ID>",
		Short: "power off the machine chassis identify LED",
		Long:  `set the machine chassis identify LED to off state`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return w.machineIdentifyOff(args)
		},
		ValidArgsFunction: c.comp.MachineListCompletion,
	}

	machineReserveCmd := &cobra.Command{
		Use:   "reserve <machine ID>",
		Short: "reserve a machine",
		Long: `reserve a machine for exclusive usage, this machine will no longer be picked by other allocations.
This is useful for maintenance of the machine or testing. After the reservation is not needed anymore, the reservation
should be removed with --remove.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return w.machineReserve(args)
		},
		ValidArgsFunction: c.comp.MachineListCompletion,
	}

	machineLockCmd := &cobra.Command{
		Use:   "lock <machine ID>",
		Short: "lock a machine",
		Long:  `when a machine is locked, it can not be destroyed, to destroy a machine you must first remove the lock from that machine with --remove`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return w.machineLock(args)
		},
		ValidArgsFunction: c.comp.MachineListCompletion,
	}

	machineReinstallCmd := &cobra.Command{
		Use:   "reinstall <machine ID>",
		Short: "reinstalls an already allocated machine",
		Long: `reinstalls an already allocated machine. If it is not yet allocated, nothing happens, otherwise only the machine's primary disk
is wiped and the new image will subsequently be installed on that device`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return w.machineReinstall(args)
		},
		ValidArgsFunction: c.comp.MachineListCompletion,
	}

	machineConsoleCmd := &cobra.Command{
		Use:   "console <machine ID>",
		Short: `console access to a machine`,
		Long: `console access to a machine, machine must be created with a ssh public key, authentication is done with your private key.
In case the machine did not register properly a direct ipmi console access is available via the --ipmi flag. This is only for administrative access.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return w.machineConsole(args)
		},
		ValidArgsFunction: c.comp.MachineListCompletion,
	}
	machineIpmiCmd := &cobra.Command{
		Use:   "ipmi [<machine ID>]",
		Short: `display ipmi details of the machine, if no machine ID is given all ipmi addresses are returned.`,
		Long:  `display ipmi details of the machine, if no machine ID is given all ipmi addresses are returned.` + "\n" + api.EmojiHelpText(),
		RunE: func(cmd *cobra.Command, args []string) error {
			return w.machineIpmi(args)
		},
		ValidArgsFunction: c.comp.MachineListCompletion,
	}
	machineIssuesCmd := &cobra.Command{
		Use:   "issues [<machine ID>]",
		Short: `display machines which are in a potential bad state`,
		Long:  `display machines which are in a potential bad state` + "\n" + api.EmojiHelpText(),
		RunE: func(cmd *cobra.Command, args []string) error {
			return w.machineIssuesEvaluate(args)
		},
		ValidArgsFunction: c.comp.MachineListCompletion,
	}
	machineIssuesListCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   `list all machine issues that the metal-api can evaluate`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return w.machineIssuesList()
		},
	}
	machineLogsCmd := &cobra.Command{
		Use:     "logs <machine ID>",
		Aliases: []string{"log"},
		Short:   `display machine provisioning logs`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return w.machineLogs(args)
		},
		ValidArgsFunction: c.comp.MachineListCompletion,
	}
	machineIpmiEventsCmd := &cobra.Command{
		Use:     "events <machine ID>",
		Aliases: []string{"event"},
		Short:   `display machine hardware events`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return w.machineIpmiEvents(args)
		},
		ValidArgsFunction: c.comp.MachineListCompletion,
	}

	w.listCmdFlags(machineIpmiCmd, 1*time.Hour)
	genericcli.AddSortFlag(machineIpmiCmd, sorters.MachineIPMISorter())

	w.listCmdFlags(machineIssuesCmd, 0)
	genericcli.AddSortFlag(machineIssuesCmd, sorters.MachineIPMISorter())

	machineIssuesCmd.AddCommand(machineIssuesListCmd)
	genericcli.AddSortFlag(machineIssuesListCmd, sorters.MachineIssueSorter())

	machineIssuesCmd.Flags().StringSlice("only", nil, "issue types to include [optional]")
	machineIssuesCmd.Flags().StringSlice("omit", nil, "issue types to omit [optional]")
	machineIssuesCmd.Flags().String("severity", "", "issue severity to include [optional]")

	genericcli.Must(machineIssuesCmd.RegisterFlagCompletionFunc("severity", c.comp.IssueSeverityCompletion))
	genericcli.Must(machineIssuesCmd.RegisterFlagCompletionFunc("omit", c.comp.IssueTypeCompletion))
	genericcli.Must(machineIssuesCmd.RegisterFlagCompletionFunc("only", c.comp.IssueTypeCompletion))

	machineLogsCmd.Flags().Duration("last-event-error-threshold", 7*24*time.Hour, "the duration up to how long in the past a machine last event error will be counted as an issue [optional]")

	machineConsolePasswordCmd.Flags().StringP("reason", "", "", "a short description why access to the consolepassword is required")

	machineUpdateBiosCmd.Flags().StringP("revision", "", "", "the BIOS revision")
	machineUpdateBiosCmd.Flags().StringP("description", "", "", "the reason why the BIOS should be updated")
	genericcli.Must(machineUpdateBiosCmd.RegisterFlagCompletionFunc("revision", c.comp.FirmwareBiosRevisionCompletion))
	machineUpdateFirmwareCmd.AddCommand(machineUpdateBiosCmd)

	machineUpdateBmcCmd.Flags().StringP("revision", "", "", "the BMC revision")
	machineUpdateBmcCmd.Flags().StringP("description", "", "", "the reason why the BMC should be updated")
	genericcli.Must(machineUpdateBmcCmd.RegisterFlagCompletionFunc("revision", c.comp.FirmwareBmcRevisionCompletion))
	machineUpdateFirmwareCmd.AddCommand(machineUpdateBmcCmd)

	machinePowerCmd.AddCommand(machinePowerOnCmd)
	machinePowerCmd.AddCommand(machinePowerOffCmd)
	machinePowerCmd.AddCommand(machinePowerResetCmd)
	machinePowerCmd.AddCommand(machinePowerCycleCmd)
	machinePowerCmd.AddCommand(machineBootBiosCmd)
	machinePowerCmd.AddCommand(machineBootDiskCmd)
	machinePowerCmd.AddCommand(machineBootPxeCmd)

	machineIdentifyOnCmd.Flags().StringP("description", "d", "", "description of the reason for chassis identify LED turn-on.")
	machineIdentifyCmd.AddCommand(machineIdentifyOnCmd)

	machineIdentifyOffCmd.Flags().StringP("description", "d", "Triggered by metalctl", "description of the reason for chassis identify LED turn-off.")
	machineIdentifyCmd.AddCommand(machineIdentifyOffCmd)

	machineReserveCmd.Flags().StringP("description", "d", "", "description of the reason for the reservation.")
	machineReserveCmd.Flags().BoolP("remove", "r", false, "remove the reservation.")

	machineLockCmd.Flags().StringP("description", "d", "", "description of the reason for the lock.")
	machineLockCmd.Flags().BoolP("remove", "r", false, "remove the lock.")

	machineReinstallCmd.Flags().StringP("image", "", "", "id of the image to get installed. [required]")
	machineReinstallCmd.Flags().StringP("description", "d", "", "description of the reinstallation. [optional]")
	genericcli.Must(machineReinstallCmd.MarkFlagRequired("image"))

	machineConsoleCmd.Flags().StringP("sshidentity", "p", "", "SSH key file, if not given the default ssh key will be used if present [optional].")
	machineConsoleCmd.Flags().BoolP("ipmi", "", false, "use ipmitool with direct network access (admin only).")
	machineConsoleCmd.Flags().BoolP("admin", "", false, "authenticate as admin (admin only).")
	machineConsoleCmd.Flags().StringP("ipmiuser", "", "", "overwrite ipmi user (admin only).")
	machineConsoleCmd.Flags().StringP("ipmipassword", "", "", "overwrite ipmi password (admin only).")

	machineIpmiEventsCmd.Flags().StringP("ipmiuser", "", "", "overwrite ipmi user (admin only).")
	machineIpmiEventsCmd.Flags().StringP("ipmipassword", "", "", "overwrite ipmi password (admin only).")
	machineIpmiEventsCmd.Flags().StringP("last", "n", "10", "show last <n> log entries.")
	machineIpmiCmd.AddCommand(machineIpmiEventsCmd)

	return genericcli.NewCmds(
		cmdsConfig,
		machineConsolePasswordCmd,
		machineConsoleCmd,
		machineIpmiCmd,
		machineIssuesCmd,
		machineLogsCmd,
		machineUpdateFirmwareCmd,
		machinePowerCmd,
		machineIdentifyCmd,
		machineReserveCmd,
		machineLockCmd,
		machineReinstallCmd,
	)
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
	default:
		log.Fatal(fmt.Errorf("illegal name: %s. Must be one of (machine, firewall)", name))
	}

	cmd.MarkFlagsMutuallyExclusive("file", "project")
	cmd.MarkFlagsRequiredTogether("project", "networks", "hostname", "image")
	cmd.MarkFlagsRequiredTogether("size", "partition")

	// Completion for arguments
	genericcli.Must(cmd.RegisterFlagCompletionFunc("networks", c.comp.NetworkListCompletion))
	genericcli.Must(cmd.RegisterFlagCompletionFunc("ips", c.comp.IpListCompletion))
	genericcli.Must(cmd.RegisterFlagCompletionFunc("partition", c.comp.PartitionListCompletion))
	genericcli.Must(cmd.RegisterFlagCompletionFunc("size", c.comp.SizeListCompletion))
	genericcli.Must(cmd.RegisterFlagCompletionFunc("project", c.comp.ProjectListCompletion))
	genericcli.Must(cmd.RegisterFlagCompletionFunc("id", c.comp.MachineListCompletion))
	genericcli.Must(cmd.RegisterFlagCompletionFunc("image", c.comp.ImageListCompletion))
	genericcli.Must(cmd.RegisterFlagCompletionFunc("filesystemlayout", c.comp.FilesystemLayoutListCompletion))
}

func (c machineCmd) Get(id string) (*models.V1MachineResponse, error) {
	resp, err := c.client.Machine().FindMachine(machine.NewFindMachineParams().WithID(id), nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c machineCmd) List() ([]*models.V1MachineResponse, error) {
	resp, err := c.client.Machine().FindMachines(machine.NewFindMachinesParams().WithBody(machineFindRequestFromCLI()), nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func machineFindRequestFromCLI() *models.V1MachineFindRequest {
	var macs []string
	if viper.IsSet("mac") {
		macs = pointer.WrapInSlice(viper.GetString("mac"))
	}

	return &models.V1MachineFindRequest{
		AllocationHostname:         viper.GetString("hostname"),
		AllocationImageID:          viper.GetString("image"),
		AllocationProject:          viper.GetString("project"),
		AllocationRole:             viper.GetString("role"),
		FruBoardPartNumber:         viper.GetString("board-part-number"),
		FruProductManufacturer:     viper.GetString("manufacturer"),
		FruProductPartNumber:       viper.GetString("product-part-number"),
		FruProductSerial:           viper.GetString("product-serial"),
		ID:                         viper.GetString("id"),
		IpmiAddress:                viper.GetString("bmc-address"),
		IpmiMacAddress:             viper.GetString("bmc-mac"),
		Name:                       viper.GetString("name"),
		NetworkDestinationPrefixes: viper.GetStringSlice("network-destination-prefixes"),
		NetworkIds:                 viper.GetStringSlice("network-ids"),
		NetworkIps:                 viper.GetStringSlice("network-ips"),
		NicsMacAddresses:           macs,
		PartitionID:                viper.GetString("partition"),
		Rackid:                     viper.GetString("rack"),
		Sizeid:                     viper.GetString("size"),
		StateValue:                 viper.GetString("state"),
		Tags:                       viper.GetStringSlice("tags"),
	}
}

func (c machineCmd) Delete(id string) (*models.V1MachineResponse, error) {
	if viper.GetBool("remove-from-database") {
		if !viper.GetBool(forceFlag) {
			return nil, fmt.Errorf("remove-from-database is set but you forgot to add --%s", forceFlag)
		}

		resp, err := c.client.Machine().DeleteMachine(machine.NewDeleteMachineParams().WithID(id), nil)
		if err != nil {
			return nil, err
		}

		return resp.Payload, nil
	}

	resp, err := c.client.Machine().FreeMachine(machine.NewFreeMachineParams().WithID(id), nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c machineCmd) Create(rq *models.V1MachineAllocateRequest) (*models.V1MachineResponse, error) {
	resp, err := c.client.Machine().AllocateMachine(machine.NewAllocateMachineParams().WithBody(rq), nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c machineCmd) Update(rq *models.V1MachineUpdateRequest) (*models.V1MachineResponse, error) {
	resp, err := c.client.Machine().UpdateMachine(machine.NewUpdateMachineParams().WithBody(rq), nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c machineCmd) Convert(r *models.V1MachineResponse) (string, *models.V1MachineAllocateRequest, *models.V1MachineUpdateRequest, error) {
	if r.ID == nil {
		return "", nil, nil, fmt.Errorf("ipaddress is nil")
	}
	return *r.ID, machineResponseToCreate(r), machineResponseToUpdate(r), nil
}

func machineResponseToCreate(r *models.V1MachineResponse) *models.V1MachineAllocateRequest {
	var (
		ips        []string
		networks   []*models.V1MachineAllocationNetwork
		allocation = pointer.SafeDeref(r.Allocation)
	)
	for _, s := range allocation.Networks {
		ips = append(ips, s.Ips...)
		networks = append(networks, &models.V1MachineAllocationNetwork{
			Autoacquire: pointer.Pointer(len(s.Ips) == 0),
			Networkid:   s.Networkid,
		})
	}

	return &models.V1MachineAllocateRequest{
		Description:        allocation.Description,
		Filesystemlayoutid: pointer.SafeDeref(pointer.SafeDeref(allocation.Filesystemlayout).ID),
		Hostname:           pointer.SafeDeref(allocation.Hostname),
		Imageid:            pointer.SafeDeref(allocation.Image).ID,
		Ips:                ips,
		Name:               r.Name,
		Networks:           networks,
		Partitionid:        pointer.SafeDeref(r.Partition).ID,
		Projectid:          allocation.Project,
		Sizeid:             pointer.SafeDeref(r.Size).ID,
		SSHPubKeys:         allocation.SSHPubKeys,
		Tags:               r.Tags,
		UserData:           base64.StdEncoding.EncodeToString([]byte(allocation.UserData)),
		UUID:               pointer.SafeDeref(r.ID),
	}
}

func machineResponseToUpdate(r *models.V1MachineResponse) *models.V1MachineUpdateRequest {
	// SSHPublicKeys should can not be updated by metalctl
	// nolint:exhaustruct
	return &models.V1MachineUpdateRequest{
		Description: pointer.PointerOrNil(pointer.SafeDeref(r.Allocation).Description),
		ID:          r.ID,
		Tags:        r.Tags,
	}
}

func (c *machineCmd) createRequestFromCLI() (*models.V1MachineAllocateRequest, error) {
	mcr, err := machineCreateRequest()
	if err != nil {
		return nil, fmt.Errorf("machine create error:%w", err)
	}

	return mcr, nil
}

func machineCreateRequest() (*models.V1MachineAllocateRequest, error) {
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

	mcr := &models.V1MachineAllocateRequest{
		Description: viper.GetString("description"),
		Partitionid: pointer.Pointer(viper.GetString("partition")),
		Hostname:    viper.GetString("hostname"),
		Imageid:     pointer.Pointer(viper.GetString("image")),
		Name:        viper.GetString("name"),
		UUID:        viper.GetString("id"),
		Projectid:   pointer.Pointer(viper.GetString("project")),
		Sizeid:      pointer.Pointer(viper.GetString("size")),
		SSHPubKeys:  keys,
		Tags:        viper.GetStringSlice("tags"),
		UserData:    userDataArgument,
		Networks:    networks,
		Ips:         viper.GetStringSlice("ips"),
	}

	if viper.IsSet("filesystemlayout") {
		mcr.Filesystemlayoutid = viper.GetString("filesystemlayout")
	}

	return mcr, nil
}

func (c *machineCmd) updateRequestFromCLI(args []string) (*models.V1MachineUpdateRequest, error) {
	id, err := genericcli.GetExactlyOneArg(args)
	if err != nil {
		return nil, err
	}

	resp, err := c.Get(id)
	if err != nil {
		return nil, err
	}

	addTags := viper.GetStringSlice("add-tags")
	removeTags := viper.GetStringSlice("remove-tags")

	for _, removeTag := range removeTags {
		if !slices.Contains(resp.Tags, removeTag) {
			return nil, fmt.Errorf("cannot remove tag because it is currently not present: %s", removeTag)
		}
	}

	newTags := addTags
	for _, t := range resp.Tags {
		if slices.Contains(removeTags, t) {
			continue
		}
		newTags = append(newTags, t)
	}

	// SSHPublicKeys should can not be updated by metalctl
	// nolint:exhaustruct
	return &models.V1MachineUpdateRequest{
		ID:          pointer.Pointer(id),
		Description: pointer.Pointer(viper.GetString("description")),
		Tags:        newTags,
	}, nil
}

// non-generic command handling

func (c *machineCmd) machineConsolePassword(args []string) error {
	id, err := genericcli.GetExactlyOneArg(args)
	if err != nil {
		return err
	}

	resp, err := c.client.Machine().GetMachineConsolePassword(machine.NewGetMachineConsolePasswordParams().WithBody(&models.V1MachineConsolePasswordRequest{
		ID:     &id,
		Reason: pointer.Pointer(viper.GetString("reason")),
	}), nil)
	if err != nil {
		return err
	}

	fmt.Fprintf(c.out, "%s\n", pointer.SafeDeref(resp.Payload.ConsolePassword))

	return nil
}

func (c *machineCmd) machinePowerOn(args []string) error {
	id, err := genericcli.GetExactlyOneArg(args)
	if err != nil {
		return err
	}

	resp, err := c.client.Machine().MachineOn(machine.NewMachineOnParams().WithID(id).WithBody(emptyBody), nil)
	if err != nil {
		return err
	}

	return c.listPrinter.Print(resp.Payload)
}

func (c *machineCmd) machinePowerOff(args []string) error {
	id, err := genericcli.GetExactlyOneArg(args)
	if err != nil {
		return err
	}

	resp, err := c.client.Machine().MachineOff(machine.NewMachineOffParams().WithID(id).WithBody(emptyBody), nil)
	if err != nil {
		return err
	}

	return c.listPrinter.Print(resp.Payload)
}

func (c *machineCmd) machinePowerReset(args []string) error {
	id, err := genericcli.GetExactlyOneArg(args)
	if err != nil {
		return err
	}

	resp, err := c.client.Machine().MachineReset(machine.NewMachineResetParams().WithID(id).WithBody(emptyBody), nil)
	if err != nil {
		return err
	}

	return c.listPrinter.Print(resp.Payload)
}

func (c *machineCmd) machinePowerCycle(args []string) error {
	id, err := genericcli.GetExactlyOneArg(args)
	if err != nil {
		return err
	}

	resp, err := c.client.Machine().MachineCycle(machine.NewMachineCycleParams().WithID(id).WithBody(emptyBody), nil)
	if err != nil {
		return err
	}

	return c.listPrinter.Print(resp.Payload)
}

func (c *machineCmd) machineUpdateBios(args []string) error {
	m, vendor, board, err := c.firmwareData(args)
	if err != nil {
		return err
	}

	revision := viper.GetString("revision")
	currentVersion := ""
	if m.Bios != nil && m.Bios.Version != nil {
		currentVersion = *m.Bios.Version
	}

	return c.machineUpdateFirmware(models.V1MachineUpdateFirmwareRequestKindBios, *m.ID, vendor, board, revision, currentVersion)
}

func (c *machineCmd) machineUpdateBmc(args []string) error {
	m, vendor, board, err := c.firmwareData(args)
	if err != nil {
		return err
	}
	revision := viper.GetString("revision")
	currentVersion := ""
	if m.Ipmi != nil && m.Ipmi.Bmcversion != nil {
		currentVersion = *m.Ipmi.Bmcversion
	}

	return c.machineUpdateFirmware(models.V1MachineUpdateFirmwareRequestKindBmc, *m.ID, vendor, board, revision, currentVersion)
}

func (c *machineCmd) firmwareData(args []string) (*models.V1MachineIPMIResponse, string, string, error) {
	id, err := genericcli.GetExactlyOneArg(args)
	if err != nil {
		return nil, "", "", err
	}

	resp, err := c.client.Machine().FindIPMIMachine(machine.NewFindIPMIMachineParams().WithID(id), nil)
	if err != nil {
		return nil, "", "", err
	}

	m := resp.Payload
	if m.Ipmi == nil {
		return nil, "", "", fmt.Errorf("no ipmi data available of machine %s", id)
	}

	fru := *m.Ipmi.Fru
	vendor := strings.ToLower(fru.BoardMfg)
	board := strings.ToUpper(fru.BoardPartNumber)

	return m, vendor, board, nil
}

func (c *machineCmd) machineUpdateFirmware(kind string, machineID, vendor, board, revision, currentVersion string) error {
	firmwareResp, err := c.client.Firmware().ListFirmwares(firmware.NewListFirmwaresParams().WithKind(&kind), nil)
	if err != nil {
		return err
	}

	var rr []string
	revisionAvailable, containsCurrentVersion := false, false
	vv, ok := firmwareResp.Payload.Revisions[string(kind)]
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
		fmt.Fprintln(c.out, "Available:")
		for _, rev := range rr {
			if rev == currentVersion {
				fmt.Fprintf(c.out, "%s (current)\n", rev)
			} else {
				fmt.Fprintln(c.out, rev)
			}
		}
		if !containsCurrentVersion {
			fmt.Fprintf(c.out, "---\nCurrent %s version: %s\n", strings.ToUpper(string(kind)), currentVersion)
		}
	}

	if revision == "" {
		return nil
	}
	if !revisionAvailable {
		return fmt.Errorf("specified revision %s not available", revision)
	}

	switch kind {
	case models.V1MachineUpdateFirmwareRequestKindBios:
		fmt.Fprintln(c.out, "It is recommended to power off the machine before updating the BIOS. This command will power on your machine automatically after the update or trigger a reboot.\n\nThe update may take a couple of minutes (up to ~10 minutes). Please wait until the machine powers on / reboots automatically as otherwise the update is still progressing or an error occurred during the update.")
	case models.V1MachineUpdateFirmwareRequestKindBmc:
		fmt.Fprintln(c.out, "The update may take a couple of minutes (up to ~10 minutes). You can look up the result through the server's BMC interface.")
	default:
		return fmt.Errorf("unsupported firmware kind: %s", kind)
	}

	if !viper.GetBool("yes-i-really-mean-it") {
		err = genericcli.Prompt()
		if err != nil {
			return err
		}
	}

	description := viper.GetString("description")
	if description == "" {
		description = "unknown"
	}

	kindString := string(kind)

	resp, err := c.client.Machine().UpdateFirmware(machine.NewUpdateFirmwareParams().WithID(machineID).WithBody(&models.V1MachineUpdateFirmwareRequest{
		Description: &description,
		Kind:        &kindString,
		Revision:    &revision,
	}), nil)
	if err != nil {
		return err
	}

	return c.listPrinter.Print(resp.Payload)
}

func (c *machineCmd) machineBootBios(args []string) error {
	id, err := genericcli.GetExactlyOneArg(args)
	if err != nil {
		return err
	}

	resp, err := c.client.Machine().MachineBios(machine.NewMachineBiosParams().WithID(id).WithBody(emptyBody), nil)
	if err != nil {
		return err
	}

	return c.listPrinter.Print(resp.Payload)
}

func (c *machineCmd) machineBootDisk(args []string) error {
	id, err := genericcli.GetExactlyOneArg(args)
	if err != nil {
		return err
	}

	resp, err := c.client.Machine().MachineDisk(machine.NewMachineDiskParams().WithID(id).WithBody(emptyBody), nil)
	if err != nil {
		return err
	}

	return c.listPrinter.Print(resp.Payload)
}

func (c *machineCmd) machineBootPxe(args []string) error {
	id, err := genericcli.GetExactlyOneArg(args)
	if err != nil {
		return err
	}

	resp, err := c.client.Machine().MachinePxe(machine.NewMachinePxeParams().WithID(id).WithBody(emptyBody), nil)
	if err != nil {
		return err
	}

	return c.listPrinter.Print(resp.Payload)
}

func (c *machineCmd) machineIdentifyOn(args []string) error {
	id, err := genericcli.GetExactlyOneArg(args)
	if err != nil {
		return err
	}

	description := pointer.Pointer(viper.GetString("description"))
	resp, err := c.client.Machine().ChassisIdentifyLEDOn(machine.NewChassisIdentifyLEDOnParams().WithID(id).WithBody(emptyBody).WithDescription(description), nil)
	if err != nil {
		return err
	}

	return c.listPrinter.Print(resp.Payload)
}

func (c *machineCmd) machineIdentifyOff(args []string) error {
	id, err := genericcli.GetExactlyOneArg(args)
	if err != nil {
		return err
	}

	description := pointer.Pointer(viper.GetString("description"))
	resp, err := c.client.Machine().ChassisIdentifyLEDOff(machine.NewChassisIdentifyLEDOffParams().WithID(id).WithBody(emptyBody).WithDescription(description), nil)
	if err != nil {
		return err
	}

	return c.listPrinter.Print(resp.Payload)
}

func (c *machineCmd) machineReserve(args []string) error {
	id, err := genericcli.GetExactlyOneArg(args)
	if err != nil {
		return err
	}

	if viper.GetBool("remove") {
		resp, err := c.client.Machine().SetMachineState(machine.NewSetMachineStateParams().WithID(id).WithBody(&models.V1MachineState{
			Description: pointer.Pointer(""),
			Value:       pointer.Pointer(models.V1MachineStateValueEmpty),
		}), nil)
		if err != nil {
			return err
		}

		return c.listPrinter.Print(resp.Payload)
	}

	resp, err := c.client.Machine().SetMachineState(machine.NewSetMachineStateParams().WithID(id).WithBody(&models.V1MachineState{
		Description: pointer.Pointer(viper.GetString("description")),
		Value:       pointer.Pointer(models.V1MachineStateValueRESERVED),
	}), nil)
	if err != nil {
		return err
	}

	return c.listPrinter.Print(resp.Payload)
}

func (c *machineCmd) machineLock(args []string) error {
	id, err := genericcli.GetExactlyOneArg(args)
	if err != nil {
		return err
	}

	if viper.GetBool("remove") {
		resp, err := c.client.Machine().SetMachineState(machine.NewSetMachineStateParams().WithID(id).WithBody(&models.V1MachineState{
			Description: pointer.Pointer(""),
			Value:       pointer.Pointer(models.V1MachineStateValueEmpty),
		}), nil)
		if err != nil {
			return err
		}

		return c.listPrinter.Print(resp.Payload)
	}

	resp, err := c.client.Machine().SetMachineState(machine.NewSetMachineStateParams().WithID(id).WithBody(&models.V1MachineState{
		Description: pointer.Pointer(viper.GetString("description")),
		Value:       pointer.Pointer(models.V1MachineStateValueLOCKED),
	}), nil)
	if err != nil {
		return err
	}

	return c.listPrinter.Print(resp.Payload)
}

func (c *machineCmd) machineReinstall(args []string) error {
	id, err := genericcli.GetExactlyOneArg(args)
	if err != nil {
		return err
	}

	resp, err := c.client.Machine().ReinstallMachine(machine.NewReinstallMachineParams().WithID(id).WithBody(&models.V1MachineReinstallRequest{
		ID:          pointer.Pointer(id),
		Description: viper.GetString("description"),
		Imageid:     pointer.Pointer(viper.GetString("image")),
	}), nil)
	if err != nil {
		return err
	}

	return c.listPrinter.Print(resp.Payload)
}

func (c *machineCmd) machineLogs(args []string) error {
	id, err := genericcli.GetExactlyOneArg(args)
	if err != nil {
		return err
	}

	// FIXME add ipmi sel as well
	resp, err := c.Get(id)
	if err != nil {
		return err
	}

	err = c.listPrinter.Print(pointer.SafeDeref(resp.Events).Log)
	if err != nil {
		return err
	}

	if pointer.SafeDeref(resp.Events).LastErrorEvent != nil {
		timeSince := time.Since(time.Time(resp.Events.LastErrorEvent.Time))
		if timeSince > viper.GetDuration("last-event-error-threshold") {
			return nil
		}

		fmt.Fprintln(c.out)
		fmt.Fprintf(c.out, "Recent last error (%s ago):\n", timeSince.String())
		fmt.Fprintln(c.out)

		return c.listPrinter.Print(resp.Events.LastErrorEvent)
	}

	return nil
}

func (c *machineCmd) machineConsole(args []string) error {
	id, err := genericcli.GetExactlyOneArg(args)
	if err != nil {
		return err
	}

	useIpmi := viper.GetBool("ipmi")
	if useIpmi {
		path, err := exec.LookPath("ipmitool")
		if err != nil {
			return fmt.Errorf("unable to locate ipmitool in path")
		}

		resp, err := c.client.Machine().FindIPMIMachine(machine.NewFindIPMIMachineParams().WithID(id), nil)
		if err != nil {
			return err
		}

		ipmi := resp.Payload.Ipmi
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
			fmt.Fprintf(c.out, "no ipmi user stored, please specify with --ipmiuser\n")
		}
		ipmiuser := viper.GetString("ipmiuser")
		if ipmiuser != "" {
			usr = ipmiuser
		}
		password := *ipmi.Password
		if *ipmi.Password == "" {
			fmt.Fprintf(c.out, "no ipmi password stored, please specify with --ipmipassword\n")
		}

		ipmipassword := viper.GetString("ipmipassword")
		if ipmipassword != "" {
			password = ipmipassword
		}

		err = os.Setenv("IPMITOOL_PASSWORD", password)
		if err != nil {
			return err
		}
		defer func() {
			_ = os.Unsetenv("IPMITOOL_PASSWORD")
		}()

		args := []string{"-I", intf, "-H", hostAndPort[0], "-p", hostAndPort[1], "-U", usr, "-E", "sol", "activate"}
		fmt.Fprintf(c.out, "connecting to console with:\n%s %s\nExit with ~.\n\n", path, strings.Join(args, " "))
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
	err = sshClient(id, key, parsedurl.Host, bmcConsolePort, &authContext.IDToken, viper.GetBool("admin"))
	if err != nil {
		return fmt.Errorf("machine console error:%w", err)
	}

	return nil
}

func (c *machineCmd) machineIpmi(args []string) error {
	if len(args) > 0 {
		id, err := genericcli.GetExactlyOneArg(args)
		if err != nil {
			return err
		}

		resp, err := c.client.Machine().FindIPMIMachine(machine.NewFindIPMIMachineParams().WithID(id), nil)
		if err != nil {
			return err
		}

		return c.describePrinter.Print(resp.Payload)
	}

	sortKeys, err := genericcli.ParseSortFlags()
	if err != nil {
		return err
	}

	resp, err := c.client.Machine().FindIPMIMachines(machine.NewFindIPMIMachinesParams().WithBody(machineFindRequestFromCLI()), nil)
	if err != nil {
		return err
	}

	err = sorters.MachineIPMISorter().SortBy(resp.Payload, sortKeys...)
	if err != nil {
		return err
	}

	return c.listPrinter.Print(resp.Payload)
}

func (c *machineCmd) machineIssuesList() error {
	issuesResp, err := c.client.Machine().ListIssues(machine.NewListIssuesParams(), nil)
	if err != nil {
		return err
	}

	sortKeys, err := genericcli.ParseSortFlags()
	if err != nil {
		return err
	}

	err = sorters.MachineIssueSorter().SortBy(issuesResp.Payload, sortKeys...)
	if err != nil {
		return err
	}

	return c.listPrinter.Print(issuesResp.Payload)
}

func (c *machineCmd) machineIssuesEvaluate(args []string) error {
	var id string
	if len(args) > 0 {
		var err error
		id, err = genericcli.GetExactlyOneArg(args)
		if err != nil {
			return err
		}
	}

	issuesResp, err := c.client.Machine().ListIssues(machine.NewListIssuesParams(), nil)
	if err != nil {
		return err
	}

	var macs []string
	if viper.IsSet("mac") {
		macs = pointer.WrapInSlice(viper.GetString("mac"))
	}

	evalResp, err := c.client.Machine().Issues(machine.NewIssuesParams().WithBody(&models.V1MachineIssuesRequest{
		AllocationHostname:         viper.GetString("hostname"),
		AllocationImageID:          viper.GetString("image"),
		AllocationProject:          viper.GetString("project"),
		AllocationRole:             viper.GetString("role"),
		FruBoardPartNumber:         viper.GetString("board-part-number"),
		FruProductManufacturer:     viper.GetString("manufacturer"),
		FruProductPartNumber:       viper.GetString("product-part-number"),
		FruProductSerial:           viper.GetString("product-serial"),
		ID:                         id,
		IpmiAddress:                viper.GetString("bmc-address"),
		IpmiMacAddress:             viper.GetString("bmc-mac"),
		Name:                       viper.GetString("name"),
		NetworkDestinationPrefixes: viper.GetStringSlice("network-destination-prefixes"),
		NetworkIds:                 viper.GetStringSlice("network-ids"),
		NetworkIps:                 viper.GetStringSlice("network-ips"),
		NicsMacAddresses:           macs,
		PartitionID:                viper.GetString("partition"),
		Rackid:                     viper.GetString("rack"),
		Sizeid:                     viper.GetString("size"),
		StateValue:                 viper.GetString("state"),
		Tags:                       viper.GetStringSlice("tags"),

		LastErrorThreshold: pointer.PointerOrNil(int64(viper.GetDuration("last-event-error-threshold"))),
		Omit:               viper.GetStringSlice("omit"),
		Only:               viper.GetStringSlice("only"),
		Severity:           pointer.PointerOrNil(viper.GetString("severity")),
	}), nil)
	if err != nil {
		return err
	}

	var machines []*models.V1MachineIPMIResponse
	if len(args) > 0 {
		machineResp, err := c.client.Machine().FindIPMIMachine(machine.NewFindIPMIMachineParams().WithID(id), nil)
		if err != nil {
			return err
		}

		machines = append(machines, machineResp.Payload)
	} else {
		machinesResp, err := c.client.Machine().FindIPMIMachines(machine.NewFindIPMIMachinesParams().WithBody(machineFindRequestFromCLI()), nil)
		if err != nil {
			return err
		}

		machines = machinesResp.Payload
	}

	sortKeys, err := genericcli.ParseSortFlags()
	if err != nil {
		return err
	}

	err = sorters.MachineIPMISorter().SortBy(machines, sortKeys...)
	if err != nil {
		return err
	}

	return c.listPrinter.Print(&tableprinters.MachinesAndIssues{
		Machines:         machines,
		Issues:           issuesResp.Payload,
		EvaluationResult: evalResp.Payload,
	})
}

func (c *machineCmd) machineIpmiEvents(args []string) error {
	id, err := genericcli.GetExactlyOneArg(args)
	if err != nil {
		return err
	}

	path, err := exec.LookPath("ipmitool")
	if err != nil {
		return fmt.Errorf("unable to locate ipmitool in path")
	}

	resp, err := c.client.Machine().FindIPMIMachine(machine.NewFindIPMIMachineParams().WithID(id), nil)
	if err != nil {
		return err
	}

	ipmi := resp.Payload.Ipmi
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
		fmt.Fprintf(c.out, "no ipmi user stored, please specify with --ipmiuser\n")
	}
	ipmiuser := viper.GetString("ipmiuser")
	if ipmiuser != "" {
		usr = ipmiuser
	}

	password := *ipmi.Password
	if *ipmi.Password == "" {
		fmt.Fprintf(c.out, "no ipmi password stored, please specify with --ipmipassword\n")
	}
	ipmipassword := viper.GetString("ipmipassword")
	if ipmipassword != "" {
		password = ipmipassword
	}
	err = os.Setenv("IPMITOOL_PASSWORD", password)
	if err != nil {
		return err
	}
	defer func() {
		_ = os.Unsetenv("IPMITOOL_PASSWORD")
	}()

	cmdArgs := []string{"-I", intf, "-H", hostAndPort[0], "-p", hostAndPort[1], "-U", usr, "-E", "sel", "list", "last", viper.GetString("last")}
	cmd := exec.Command(path, cmdArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stdout

	return cmd.Run()
}
