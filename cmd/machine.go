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

	"golang.org/x/exp/slices"

	"github.com/metal-stack/metal-go/api/client/firmware"
	"github.com/metal-stack/metal-go/api/client/machine"
	"github.com/metal-stack/metal-go/api/models"
	"github.com/metal-stack/metal-lib/pkg/genericcli"
	"github.com/metal-stack/metal-lib/pkg/pointer"
	"github.com/metal-stack/metalctl/cmd/printers"
	"github.com/metal-stack/metalctl/cmd/printers/tableprinters"
	"github.com/metal-stack/metalctl/cmd/sorters"
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

func (c *machineCmd) listCmdFlags(cmd *cobra.Command) {
	listFlagCompletions := []struct {
		flagName string
		f        func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective)
	}{
		{flagName: "partition", f: c.comp.PartitionListCompletion},
		{flagName: "size", f: c.comp.SizeListCompletion},
		{flagName: "project", f: c.comp.ProjectListCompletion},
		{flagName: "id", f: c.comp.MachineListCompletion},
		{flagName: "image", f: c.comp.ImageListCompletion},
	}

	cmd.Flags().String("id", "", "ID to filter [optional]")
	cmd.Flags().String("partition", "", "partition to filter [optional]")
	cmd.Flags().String("size", "", "size to filter [optional]")
	cmd.Flags().String("name", "", "allocation name to filter [optional]")
	cmd.Flags().String("project", "", "allocation project to filter [optional]")
	cmd.Flags().String("image", "", "allocation image to filter [optional]")
	cmd.Flags().String("hostname", "", "allocation hostname to filter [optional]")
	cmd.Flags().String("mac", "", "mac to filter [optional]")
	cmd.Flags().StringSlice("tags", []string{}, "tags to filter, use it like: --tags \"tag1,tag2\" or --tags \"tag3\".")
	for _, c := range listFlagCompletions {
		c := c
		must(cmd.RegisterFlagCompletionFunc(c.flagName, c.f))
	}
	cmd.Long = cmd.Short + "\n" + api.EmojiHelpText()
}

func newMachineCmd(c *config) *cobra.Command {
	w := machineCmd{
		config: c,
	}

	cmdsConfig := &genericcli.CmdsConfig[*models.V1MachineAllocateRequest, *models.V1MachineUpdateRequest, *models.V1MachineResponse]{
		BinaryName:           binaryName,
		GenericCLI:           genericcli.NewGenericCLI[*models.V1MachineAllocateRequest, *models.V1MachineUpdateRequest, *models.V1MachineResponse](w),
		Singular:             "machine",
		Plural:               "machines",
		Description:          "a machine is a bare metal server provisioned through metal-stack that is intended to run user workload.",
		Aliases:              []string{"ms"},
		CreateRequestFromCLI: w.createRequestFromCLI,
		UpdateRequestFromCLI: w.updateRequestFromCLI,
		AvailableSortKeys:    sorters.MachineSortKeys(),
		ValidArgsFn:          c.comp.MachineListCompletion,
		DescribePrinter:      printers.DefaultToYAMLPrinter(),
		ListPrinter:          printers.NewPrinterFromCLI(),
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
			w.listCmdFlags(cmd)
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
			return w.machinePowerOn(args)
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
			return w.machinePowerOff(args)
		},
		PreRun:            bindPFlags,
		ValidArgsFunction: c.comp.MachineListCompletion,
	}

	machinePowerResetCmd := &cobra.Command{
		Use:   "reset <machine ID>",
		Short: "power reset a machine",
		Long:  "(hard) reset the machine power.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return w.machinePowerReset(args)
		},
		PreRun:            bindPFlags,
		ValidArgsFunction: c.comp.MachineListCompletion,
	}

	machinePowerCycleCmd := &cobra.Command{
		Use:   "cycle <machine ID>",
		Short: "power cycle a machine",
		Long:  "(soft) cycle the machine power.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return w.machinePowerCycle(args)
		},
		PreRun:            bindPFlags,
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
		PreRun:            bindPFlags,
		ValidArgsFunction: c.comp.MachineListCompletion,
	}

	machineUpdateBmcCmd := &cobra.Command{
		Use:   "bmc <machine ID>",
		Short: "update a machine BMC",
		Long:  "the machine BMC will be updated to given revision. If revision flag is not specified an update plan will be printed instead.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return w.machineUpdateBmc(args)
		},
		PreRun:            bindPFlags,
		ValidArgsFunction: c.comp.MachineListCompletion,
	}

	machineBootBiosCmd := &cobra.Command{
		Use:   "bios <machine ID>",
		Short: "boot a machine into BIOS",
		Long:  "the machine will boot into bios. (machine does not reboot automatically)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return w.machineBootBios(args)
		},
		PreRun:            bindPFlags,
		ValidArgsFunction: c.comp.MachineListCompletion,
	}

	machineBootPxeCmd := &cobra.Command{
		Use:   "pxe <machine ID>",
		Short: "boot a machine from PXE",
		Long:  "the machine will boot from PXE. (machine does not reboot automatically)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return w.machineBootPxe(args)
		},
		PreRun:            bindPFlags,
		ValidArgsFunction: c.comp.MachineListCompletion,
	}

	machineBootDiskCmd := &cobra.Command{
		Use:   "disk <machine ID>",
		Short: "boot a machine from disk",
		Long:  "the machine will boot from disk. (machine does not reboot automatically)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return w.machineBootDisk(args)
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
			return w.machineIdentifyOn(args)
		},
		PreRun:            bindPFlags,
		ValidArgsFunction: c.comp.MachineListCompletion,
	}

	machineIdentifyOffCmd := &cobra.Command{
		Use:   "off <machine ID>",
		Short: "power off the machine chassis identify LED",
		Long:  `set the machine chassis identify LED to off state`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return w.machineIdentifyOff(args)
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
			return w.machineReserve(args)
		},
		PreRun:            bindPFlags,
		ValidArgsFunction: c.comp.MachineListCompletion,
	}

	machineLockCmd := &cobra.Command{
		Use:   "lock <machine ID>",
		Short: "lock a machine",
		Long:  `when a machine is locked, it can not be destroyed, to destroy a machine you must first remove the lock from that machine with --remove`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return w.machineLock(args)
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
			return w.machineReinstall(args)
		},
		PreRun:            bindPFlags,
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
		PreRun:            bindPFlags,
		ValidArgsFunction: c.comp.MachineListCompletion,
	}
	machineIpmiCmd := &cobra.Command{
		Use:   "ipmi [<machine ID>]",
		Short: `display ipmi details of the machine, if no machine ID is given all ipmi addresses are returned.`,
		Long:  `display ipmi details of the machine, if no machine ID is given all ipmi addresses are returned.` + "\n" + api.EmojiHelpText(),
		RunE: func(cmd *cobra.Command, args []string) error {
			return w.machineIpmi(args)
		},
		PreRun: bindPFlags,
	}
	machineIssuesCmd := &cobra.Command{
		Use:   "issues",
		Short: `display machines which are in a potential bad state`,
		Long:  `display machines which are in a potential bad state` + "\n" + api.EmojiHelpText(),
		RunE: func(cmd *cobra.Command, args []string) error {
			return w.machineIssues()
		},
		PreRun: bindPFlags,
	}
	machineLogsCmd := &cobra.Command{
		Use:     "logs <machine ID>",
		Aliases: []string{"log"},
		Short:   `display machine provisioning logs`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return w.machineLogs(args)
		},
		PreRun:            bindPFlags,
		ValidArgsFunction: c.comp.MachineListCompletion,
	}
	machineIpmiEventsCmd := &cobra.Command{
		Use:     "events <machine ID>",
		Aliases: []string{"event"},
		Short:   `display machine hardware events`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return w.machineIpmiEvents(args)
		},
		PreRun:            bindPFlags,
		ValidArgsFunction: c.comp.MachineListCompletion,
	}

	w.listCmdFlags(machineIpmiCmd)
	w.listCmdFlags(machineIssuesCmd)

	machineIssuesCmd.Flags().StringSliceP("only", "", []string{}, "issue types to include [optional]")
	machineIssuesCmd.Flags().StringSliceP("omit", "", []string{}, "issue types to omit [optional]")

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

	machineConsolePasswordCmd.Flags().StringP("reason", "", "", "a short description why access to the consolepassword is required")

	machineUpdateBiosCmd.Flags().StringP("revision", "", "", "the BIOS revision")
	machineUpdateBiosCmd.Flags().StringP("description", "", "", "the reason why the BIOS should be updated")
	must(machineUpdateBiosCmd.RegisterFlagCompletionFunc("revision", c.comp.FirmwareBiosRevisionCompletion))
	machineUpdateFirmwareCmd.AddCommand(machineUpdateBiosCmd)

	machineUpdateBmcCmd.Flags().StringP("revision", "", "", "the BMC revision")
	machineUpdateBmcCmd.Flags().StringP("description", "", "", "the reason why the BMC should be updated")
	must(machineUpdateBmcCmd.RegisterFlagCompletionFunc("revision", c.comp.FirmwareBmcRevisionCompletion))
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
	must(machineReinstallCmd.MarkFlagRequired("image"))

	machineConsoleCmd.Flags().StringP("sshidentity", "p", "", "SSH key file, if not given the default ssh key will be used if present [optional].")
	machineConsoleCmd.Flags().BoolP("ipmi", "", false, "use ipmitool with direct network access (admin only).")
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

	err = sorters.MachineSort(resp.Payload)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func machineFindRequestFromCLI() *models.V1MachineFindRequest {
	return &models.V1MachineFindRequest{
		ID:                 viper.GetString("id"),
		PartitionID:        viper.GetString("partition"),
		Sizeid:             viper.GetString("size"),
		Name:               viper.GetString("name"),
		AllocationProject:  viper.GetString("project"),
		AllocationImageID:  viper.GetString("image"),
		AllocationHostname: viper.GetString("hostname"),
		NicsMacAddresses:   viper.GetStringSlice("mac"),
		Tags:               viper.GetStringSlice("tags"),
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

	fmt.Printf("%s\n", pointer.SafeDeref(resp.Payload.ConsolePassword))

	return nil
}

func (c *machineCmd) machinePowerOn(args []string) error {
	id, err := genericcli.GetExactlyOneArg(args)
	if err != nil {
		return err
	}

	resp, err := c.client.Machine().MachineOn(machine.NewMachineOnParams().WithID(id), nil)
	if err != nil {
		return err
	}

	return printers.NewPrinterFromCLI().Print(resp.Payload)
}

func (c *machineCmd) machinePowerOff(args []string) error {
	id, err := genericcli.GetExactlyOneArg(args)
	if err != nil {
		return err
	}

	resp, err := c.client.Machine().MachineOff(machine.NewMachineOffParams().WithID(id), nil)
	if err != nil {
		return err
	}

	return printers.NewPrinterFromCLI().Print(resp.Payload)
}

func (c *machineCmd) machinePowerReset(args []string) error {
	id, err := genericcli.GetExactlyOneArg(args)
	if err != nil {
		return err
	}

	resp, err := c.client.Machine().MachineReset(machine.NewMachineResetParams().WithID(id), nil)
	if err != nil {
		return err
	}

	return printers.NewPrinterFromCLI().Print(resp.Payload)
}

func (c *machineCmd) machinePowerCycle(args []string) error {
	id, err := genericcli.GetExactlyOneArg(args)
	if err != nil {
		return err
	}

	resp, err := c.client.Machine().MachineCycle(machine.NewMachineCycleParams().WithID(id), nil)
	if err != nil {
		return err
	}

	return printers.NewPrinterFromCLI().Print(resp.Payload)
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
	case models.V1MachineUpdateFirmwareRequestKindBios:
		fmt.Println("It is recommended to power off the machine before updating the BIOS. This command will power on your machine automatically after the update or trigger a reboot.\n\nThe update may take a couple of minutes (up to ~10 minutes). Please wait until the machine powers on / reboots automatically as otherwise the update is still progressing or an error occurred during the update.")
	case models.V1MachineUpdateFirmwareRequestKindBmc:
		fmt.Println("The update may take a couple of minutes (up to ~10 minutes). You can look up the result through the server's BMC interface.")
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

	return printers.NewPrinterFromCLI().Print(resp.Payload)
}

func (c *machineCmd) machineBootBios(args []string) error {
	id, err := genericcli.GetExactlyOneArg(args)
	if err != nil {
		return err
	}

	resp, err := c.client.Machine().MachineBios(machine.NewMachineBiosParams().WithID(id), nil)
	if err != nil {
		return err
	}

	return printers.NewPrinterFromCLI().Print(resp.Payload)
}

func (c *machineCmd) machineBootDisk(args []string) error {
	id, err := genericcli.GetExactlyOneArg(args)
	if err != nil {
		return err
	}

	resp, err := c.client.Machine().MachineDisk(machine.NewMachineDiskParams().WithID(id), nil)
	if err != nil {
		return err
	}

	return printers.NewPrinterFromCLI().Print(resp.Payload)
}

func (c *machineCmd) machineBootPxe(args []string) error {
	id, err := genericcli.GetExactlyOneArg(args)
	if err != nil {
		return err
	}

	resp, err := c.client.Machine().MachinePxe(machine.NewMachinePxeParams().WithID(id), nil)
	if err != nil {
		return err
	}

	return printers.NewPrinterFromCLI().Print(resp.Payload)
}

func (c *machineCmd) machineIdentifyOn(args []string) error {
	id, err := genericcli.GetExactlyOneArg(args)
	if err != nil {
		return err
	}

	description := pointer.Pointer(viper.GetString("description"))
	resp, err := c.client.Machine().ChassisIdentifyLEDOn(machine.NewChassisIdentifyLEDOnParams().WithID(id).WithDescription(description), nil)
	if err != nil {
		return err
	}

	return printers.NewPrinterFromCLI().Print(resp.Payload)
}

func (c *machineCmd) machineIdentifyOff(args []string) error {
	id, err := genericcli.GetExactlyOneArg(args)
	if err != nil {
		return err
	}

	description := pointer.Pointer(viper.GetString("description"))
	resp, err := c.client.Machine().ChassisIdentifyLEDOff(machine.NewChassisIdentifyLEDOffParams().WithID(id).WithDescription(description), nil)
	if err != nil {
		return err
	}

	return printers.NewPrinterFromCLI().Print(resp.Payload)
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

		return printers.NewPrinterFromCLI().Print(resp.Payload)
	}

	resp, err := c.client.Machine().SetMachineState(machine.NewSetMachineStateParams().WithID(id).WithBody(&models.V1MachineState{
		Description: pointer.Pointer(viper.GetString("description")),
		Value:       pointer.Pointer(models.V1MachineStateValueRESERVED),
	}), nil)
	if err != nil {
		return err
	}

	return printers.NewPrinterFromCLI().Print(resp.Payload)
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

		return printers.NewPrinterFromCLI().Print(resp.Payload)
	}

	resp, err := c.client.Machine().SetMachineState(machine.NewSetMachineStateParams().WithID(id).WithBody(&models.V1MachineState{
		Description: pointer.Pointer(viper.GetString("description")),
		Value:       pointer.Pointer(models.V1MachineStateValueLOCKED),
	}), nil)
	if err != nil {
		return err
	}

	return printers.NewPrinterFromCLI().Print(resp.Payload)
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

	return printers.NewPrinterFromCLI().Print(resp.Payload)
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

	err = printers.NewPrinterFromCLI().Print(pointer.SafeDeref(resp.Events).Log)
	if err != nil {
		return err
	}

	if pointer.SafeDeref(resp.Events).LastErrorEvent != nil {
		timeSince := time.Since(time.Time(resp.Events.LastErrorEvent.Time))
		if timeSince > tableprinters.LastErrorEventRelevant {
			return nil
		}

		fmt.Println()
		fmt.Printf("Recent last error (%s ago):\n", timeSince.String())
		fmt.Println()

		return printers.NewPrinterFromCLI().Print(resp.Events.LastErrorEvent)
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
	err = SSHClient(id, key, parsedurl.Host, bmcConsolePort)
	if err != nil {
		return fmt.Errorf("machine console error:%w", err)
	}

	return nil
}

func (c *machineCmd) machineIpmi(args []string) error {
	if len(args) == 1 {
		id, err := genericcli.GetExactlyOneArg(args)
		if err != nil {
			return err
		}

		resp, err := c.client.Machine().FindIPMIMachine(machine.NewFindIPMIMachineParams().WithID(id), nil)
		if err != nil {
			return err
		}

		hidden := "<hidden>"
		resp.Payload.Ipmi.Password = &hidden

		return printers.DefaultToYAMLPrinter().Print(resp.Payload)
	}

	resp, err := c.client.Machine().FindIPMIMachines(machine.NewFindIPMIMachinesParams().WithBody(machineFindRequestFromCLI()), nil)
	if err != nil {
		return err
	}

	err = sorters.MachineIPMISort(resp.Payload)
	if err != nil {
		return err
	}

	return printers.NewPrinterFromCLI().Print(resp.Payload)
}

func (c *machineCmd) machineIssues() error {
	resp, err := c.client.Machine().FindIPMIMachines(machine.NewFindIPMIMachinesParams().WithBody(machineFindRequestFromCLI()), nil)
	if err != nil {
		return err
	}

	err = sorters.MachineIPMISort(resp.Payload)
	if err != nil {
		return err
	}

	var (
		only = viper.GetStringSlice("only")
		omit = viper.GetStringSlice("omit")

		res      = api.MachineIssues{}
		asnMap   = map[int64][]*models.V1MachineIPMIResponse{}
		bmcIPMap = map[string][]*models.V1MachineIPMIResponse{}

		includeThisIssue = func(issue api.Issue) bool {
			for _, o := range omit {
				if issue.ShortName == o {
					return false
				}
			}

			if len(only) > 0 {
				for _, o := range only {
					if issue.ShortName == o {
						return true
					}
				}
				return false
			}

			return true
		}

		addIssue = func(m *models.V1MachineIPMIResponse, issue api.Issue) {
			if m == nil {
				return
			}

			if !includeThisIssue(issue) {
				return
			}

			var mWithIssues *api.MachineWithIssues
			for _, machine := range res {
				machine := machine
				if pointer.SafeDeref(m.ID) == pointer.SafeDeref(machine.Machine.ID) {
					mWithIssues = &machine
					break
				}
			}
			if mWithIssues == nil {
				mWithIssues = &api.MachineWithIssues{
					Machine: *m,
				}
				res = append(res, *mWithIssues)
			}

			mWithIssues.Issues = append(mWithIssues.Issues, issue)
		}
	)

	for _, m := range resp.Payload {
		if m == nil {
			continue
		}

		if m.Partition == nil {
			addIssue(m, api.IssueNoPartition)
		}

		if m.Liveliness != nil {
			switch *m.Liveliness {
			case "Alive":
			case "Dead":
				addIssue(m, api.IssueLivelinessDead)
			case "Unknown":
				addIssue(m, api.IssueLivelinessUnknown)

			default:
				addIssue(m, api.IssueLivelinessNotAvailable)
			}
		} else {
			addIssue(m, api.IssueLivelinessNotAvailable)
		}

		if pointer.SafeDeref(pointer.SafeDeref(m.Events).FailedMachineReclaim) {
			addIssue(m, api.IssueFailedMachineReclaim)
		}

		if pointer.SafeDeref(pointer.SafeDeref(m.Events).CrashLoop) {
			if m.Events != nil && len(m.Events.Log) > 0 && *m.Events.Log[0].Event == "Waiting" {
				// Machine which are waiting are not considered to have issues
			} else {
				addIssue(m, api.IssueCrashLoop)
			}
		}

		if pointer.SafeDeref(m.Events).LastErrorEvent != nil {
			timeSince := time.Since(time.Time(m.Events.LastErrorEvent.Time))
			if timeSince < tableprinters.LastErrorEventRelevant {
				issue := api.IssueLastEventError
				issue.Description = fmt.Sprintf("%s (%s ago)", issue.Description, timeSince.String())
				addIssue(m, issue)
			}
		}

		if m.Ipmi != nil {
			if m.Ipmi.Mac == nil || *m.Ipmi.Mac == "" {
				addIssue(m, api.IssueBMCWithoutMAC)
			}

			if m.Ipmi.Address == nil || *m.Ipmi.Address == "" {
				addIssue(m, api.IssueBMCWithoutIP)
			} else {
				entries := bmcIPMap[*m.Ipmi.Address]
				entries = append(entries, m)
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
					machines = []*models.V1MachineIPMIResponse{}
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

				machines = append(machines, m)
				asnMap[*n.Asn] = machines
			}
		}
	}

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

			issue := api.IssueASNUniqueness
			issue.Description = fmt.Sprintf("ASN (%d) not unique, shared with %s", asn, sharedIDs)

			addIssue(m, issue)
		}
	}

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

			issue := api.IssueNonDistinctBMCIP
			issue.Description = fmt.Sprintf("BMC IP (%s) not unique, shared with %s", ip, sharedIDs)

			addIssue(m, issue)
		}
	}

	return printers.NewPrinterFromCLI().Print(res)
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

	cmdArgs := []string{"-I", intf, "-H", hostAndPort[0], "-p", hostAndPort[1], "-U", usr, "-P", password, "sel", "list", "last", viper.GetString("last")}
	cmd := exec.Command(path, cmdArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stdout

	return cmd.Run()
}
