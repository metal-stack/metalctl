package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/metal-stack/metal-go/api/client/switch_operations"
	"github.com/metal-stack/metal-go/api/models"
	"github.com/metal-stack/metal-lib/pkg/genericcli"
	"github.com/metal-stack/metal-lib/pkg/genericcli/printers"
	"github.com/metal-stack/metalctl/cmd/sorters"
	"github.com/metal-stack/metalctl/cmd/tableprinters"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type switchCmd struct {
	*config
}

func newSwitchCmd(c *config) *cobra.Command {
	w := switchCmd{
		config: c,
	}

	cmdsConfig := &genericcli.CmdsConfig[any, *models.V1SwitchUpdateRequest, *models.V1SwitchResponse]{
		BinaryName: binaryName,
		GenericCLI: genericcli.NewGenericCLI[any, *models.V1SwitchUpdateRequest, *models.V1SwitchResponse](w).WithFS(c.fs),
		OnlyCmds: genericcli.OnlyCmds(
			genericcli.ListCmd,
			genericcli.DescribeCmd,
			genericcli.UpdateCmd,
			genericcli.DeleteCmd,
			genericcli.EditCmd,
		),
		Aliases:         []string{"sw"},
		Singular:        "switch",
		Plural:          "switches",
		Description:     "switch are the leaf switches in the data center that are controlled by metal-stack.",
		Sorter:          sorters.SwitchSorter(),
		ValidArgsFn:     c.comp.SwitchListCompletion,
		DescribePrinter: func() printers.Printer { return c.describePrinter },
		ListPrinter:     func() printers.Printer { return c.listPrinter },
		ListCmdMutateFn: func(cmd *cobra.Command) {
			cmd.Flags().String("id", "", "ID of the switch.")
			cmd.Flags().String("name", "", "Name of the switch.")
			cmd.Flags().String("os-vendor", "", "OS vendor of this switch.")
			cmd.Flags().String("os-version", "", "OS version of this switch.")
			cmd.Flags().String("partition", "", "Partition of this switch.")
			cmd.Flags().String("rack", "", "Rack of this switch.")

			must(cmd.RegisterFlagCompletionFunc("id", c.comp.SwitchListCompletion))
			must(cmd.RegisterFlagCompletionFunc("name", c.comp.SwitchNameListCompletion))
			must(cmd.RegisterFlagCompletionFunc("partition", c.comp.PartitionListCompletion))
			must(cmd.RegisterFlagCompletionFunc("rack", c.comp.SwitchRackListCompletion))
			must(cmd.RegisterFlagCompletionFunc("os-vendor", c.comp.SwitchOSVendorListCompletion))
			must(cmd.RegisterFlagCompletionFunc("os-version", c.comp.SwitchOSVersionListCompletion))
		},
	}

	switchDetailCmd := &cobra.Command{
		Use:   "detail",
		Short: "switch details",
		RunE: func(cmd *cobra.Command, args []string) error {
			return w.switchDetail()
		},
		ValidArgsFunction: c.comp.SwitchListCompletion,
	}

	switchDetailCmd.Flags().String("id", "", "ID of the switch.")
	switchDetailCmd.Flags().String("name", "", "Name of the switch.")
	switchDetailCmd.Flags().String("os-vendor", "", "OS vendor of this switch.")
	switchDetailCmd.Flags().String("os-version", "", "OS version of this switch.")
	switchDetailCmd.Flags().String("partition", "", "Partition of this switch.")
	switchDetailCmd.Flags().String("rack", "", "Rack of this switch.")

	must(switchDetailCmd.RegisterFlagCompletionFunc("id", c.comp.SwitchListCompletion))
	must(switchDetailCmd.RegisterFlagCompletionFunc("name", c.comp.SwitchNameListCompletion))
	must(switchDetailCmd.RegisterFlagCompletionFunc("partition", c.comp.PartitionListCompletion))
	must(switchDetailCmd.RegisterFlagCompletionFunc("rack", c.comp.SwitchRackListCompletion))

	switchReplaceCmd := &cobra.Command{
		Use:   "replace <switchID>",
		Short: "put a leaf switch into replace mode in preparation for physical replacement. For a description of the steps involved see the long help.",
		Long: `Put a leaf switch into replace mode in preparation for physical replacement

Operational steps to replace a switch:

- Put the switch that needs to be replaced in replace mode with this command
- Replace the switch MAC address in the metal-stack deployment configuration
- Make sure that interfaces on the new switch do not get connected to the PXE-bridge immediately by setting the interfaces list of the respective leaf switch to [] in the metal-stack deployment configuration
- Deploy the management servers so that the dhcp servers will serve the right address and DHCP options to the new switch
- Replace the switch physically. Be careful to ensure that the cabling mirrors the remaining leaf exactly because the new switch information will be cloned from the remaining switch! Also make sure to have console access to the switch so you can start and monitor the install process
- If the switch is not in onie install mode but already has an operating system installed, put it into install mode with "sudo onie-select -i -f -v" and reboot it. Now the switch should be provisioned with a management IP from a management server, install itself with the right software image and receive license and ssh keys through ZTP. You can check whether that process has completed successfully with the command "sudo ztp -s". The ZTP state should be disabled and the result should be success.
- Deploy the switch plane and metal-core through metal-stack deployment CI job
- The switch will now register with its metal-api, and the metal-core service will receive the cloned interface and routing information. You can verify successful switch replacement by checking the interface and BGP configuration, and checking the switch status with "metalctl switch ls -o wide"; it should now be operational again`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return w.switchReplace(args)
		},
		ValidArgsFunction: c.comp.SwitchListCompletion,
	}
	switchSSHCmd := &cobra.Command{
		Use:   "ssh <switchID>",
		Short: "connect to the switch via ssh",
		Long:  "this requires a network connectivity to the management ip address of the switch.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return w.switchSSH(args)
		},
		ValidArgsFunction: c.comp.SwitchListCompletion,
	}
	switchConsoleCmd := &cobra.Command{
		Use:   "console <switchID>",
		Short: "connect to the switch console",
		Long:  "this requires a network connectivity to the ip address of the console server this switch is connected to.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return w.switchConsole(args)
		},
		ValidArgsFunction: c.comp.SwitchListCompletion,
	}

	return genericcli.NewCmds(cmdsConfig, switchDetailCmd, switchReplaceCmd, switchSSHCmd, switchConsoleCmd)
}

func (c switchCmd) Get(id string) (*models.V1SwitchResponse, error) {
	resp, err := c.client.SwitchOperations().FindSwitch(switch_operations.NewFindSwitchParams().WithID(id), nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c switchCmd) List() ([]*models.V1SwitchResponse, error) {
	resp, err := c.client.SwitchOperations().FindSwitches(switch_operations.NewFindSwitchesParams().WithBody(&models.V1SwitchFindRequest{
		ID:          viper.GetString("id"),
		Name:        viper.GetString("name"),
		Osvendor:    viper.GetString("os-vendor"),
		Osversion:   viper.GetString("os-version"),
		Partitionid: viper.GetString("partition"),
		Rackid:      viper.GetString("rack"),
	}), nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c switchCmd) Delete(id string) (*models.V1SwitchResponse, error) {
	resp, err := c.client.SwitchOperations().DeleteSwitch(switch_operations.NewDeleteSwitchParams().WithID(id), nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c switchCmd) Create(rq any) (*models.V1SwitchResponse, error) {
	return nil, fmt.Errorf("switch entity does not support create operation")
}

func (c switchCmd) Update(rq *models.V1SwitchUpdateRequest) (*models.V1SwitchResponse, error) {
	resp, err := c.client.SwitchOperations().UpdateSwitch(switch_operations.NewUpdateSwitchParams().WithBody(rq), nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c switchCmd) ToCreate(r *models.V1SwitchResponse) (any, error) {
	return nil, fmt.Errorf("switch entity does not support create operation")
}

func (c switchCmd) ToUpdate(r *models.V1SwitchResponse) (*models.V1SwitchUpdateRequest, error) {
	return switchResponseToUpdate(r), nil
}

func switchResponseToUpdate(r *models.V1SwitchResponse) *models.V1SwitchUpdateRequest {
	switchOS := &models.V1SwitchOS{}
	if r.Os != nil {
		switchOS.Vendor = r.Os.Vendor
		switchOS.Vendor = r.Os.Version
	}

	return &models.V1SwitchUpdateRequest{
		ConsoleCommand: r.ConsoleCommand,
		Description:    r.Description,
		ID:             r.ID,
		ManagementIP:   "",
		ManagementUser: "",
		Mode:           r.Mode,
		Name:           r.Name,
		Os:             switchOS,
		RackID:         r.RackID,
	}
}

// non-generic command handling

func (c *switchCmd) switchDetail() error {
	resp, err := c.List()
	if err != nil {
		return err
	}

	var result []*tableprinters.SwitchDetail
	for _, s := range resp {
		result = append(result, &tableprinters.SwitchDetail{V1SwitchResponse: s})
	}

	return c.listPrinter.Print(result)
}

func (c *switchCmd) switchReplace(args []string) error {
	id, err := genericcli.GetExactlyOneArg(args)
	if err != nil {
		return err
	}

	resp, err := c.Get(id)
	if err != nil {
		return err
	}

	switchOS := &models.V1SwitchOS{}
	if resp.Os != nil {
		switchOS.Vendor = resp.Os.Vendor
		switchOS.Vendor = resp.Os.Version
	}

	uresp, err := c.Update(&models.V1SwitchUpdateRequest{
		ConsoleCommand: "",
		Description:    resp.Description,
		ID:             resp.ID,
		ManagementIP:   "",
		ManagementUser: "",
		Mode:           "replace",
		Name:           resp.Name,
		Os:             switchOS,
		RackID:         resp.RackID,
	})
	if err != nil {
		return err
	}

	return c.describePrinter.Print(uresp)
}

func (c *switchCmd) switchSSH(args []string) error {
	id, err := genericcli.GetExactlyOneArg(args)
	if err != nil {
		return err
	}

	resp, err := c.Get(id)
	if err != nil {
		return err
	}
	if resp.ManagementIP == "" || resp.ManagementUser == "" {
		return fmt.Errorf("unable to connect to switch by ssh because no ip and user was stored for this switch, please restart metal-core on this switch")
	}

	// nolint: gosec
	cmd := exec.Command("ssh", fmt.Sprintf("%s@%s", resp.ManagementUser, resp.ManagementIP))
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stdout
	return cmd.Run()
}

func (c *switchCmd) switchConsole(args []string) error {
	id, err := genericcli.GetExactlyOneArg(args)
	if err != nil {
		return err
	}

	resp, err := c.Get(id)
	if err != nil {
		return err
	}

	if resp.ConsoleCommand == "" {
		return fmt.Errorf(`
unable to connect to console because no console_command was specified for this switch
You can add a working console_command to every switch with metalctl switch edit
A sample would look like:

telnet console-server 7008`)
	}
	parts := strings.Fields(resp.ConsoleCommand)

	// nolint: gosec
	cmd := exec.Command(parts[0])
	if len(parts) > 1 {
		// nolint: gosec
		cmd = exec.Command(parts[0], parts[1:]...)
	}
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stdout
	return cmd.Run()
}
