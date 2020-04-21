package cmd

import (
	"fmt"

	metalgo "github.com/metal-stack/metal-go"
	"github.com/spf13/cobra"
)

var (
	firewallCmd = &cobra.Command{
		Use:   "firewall",
		Short: "manage firewalls",
		Long:  "metal firewalls are bare metal firewalls.",
	}

	firewallCreateCmd = &cobra.Command{
		Use:   "create",
		Short: "create a firewall",
		Long:  `create a new firewall connected to the given networks.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return firewallCreate(driver)
		},
		PreRun: bindPFlags,
	}

	firewallListCmd = &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "list all firewalls",
		Long:    "list all firewalls with almost all properties in tabular form.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return firewallList(driver)
		},
		PreRun: bindPFlags,
	}

	firewallDescribeCmd = &cobra.Command{
		Use:   "describe <firewall ID>",
		Short: "describe a firewall",
		Long:  "describe a firewall in a very detailed form with all properties.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return firewallDescribe(driver, args)
		},
		PreRun: bindPFlags,
	}

	firewallDestroyCmd = &cobra.Command{
		Use:     "destroy <firewall ID>",
		Aliases: []string{"delete", "rm"},
		Short:   "destroy a firewall",
		Long: `destroy a firewall and destroy all data stored on the local disks. Once destroyed it is back for usage by other projects.
A destroyed firewall can not restored anymore`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return firewallDestroy(driver, args)
		},
		PreRun: bindPFlags,
	}

	firewallReserveCmd = &cobra.Command{
		Use:   "reserve <firewall ID>",
		Short: "reserve a firewall",
		Long: `reserve a firewall for exclusive usage, this firewall will no longer be picked by other allocations.
This is useful for maintenance of the firewall or testing. After the reservation is not needed anymore, the reservation
should be removed with --remove.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return machineReserve(driver, args)
		},
		PreRun: bindPFlags,
	}
)

func init() {
	addMachineCreateFlags(firewallCreateCmd, "firewall")
	firewallListCmd.Flags().StringP("partition", "", "", "partition to filter [optional]")
	firewallListCmd.Flags().StringP("project", "", "", "project to filter [optional]")

	firewallCmd.AddCommand(firewallCreateCmd)
	firewallCmd.AddCommand(firewallListCmd)
	firewallCmd.AddCommand(firewallDestroyCmd)
	firewallCmd.AddCommand(firewallDescribeCmd)
	firewallCmd.AddCommand(firewallReserveCmd)
}

func firewallCreate(driver *metalgo.Driver) error {
	mcr, err := machineCreateRequest()
	if err != nil {
		return fmt.Errorf("firewall create error:%v", err)
	}

	fcr := &metalgo.FirewallCreateRequest{
		MachineCreateRequest: *mcr,
	}
	resp, err := driver.FirewallCreate(fcr)
	if err != nil {
		return fmt.Errorf("firewall create error:%v", err)
	}
	return printer.Print(resp.Firewall)
}

func firewallList(driver *metalgo.Driver) error {
	var resp *metalgo.FirewallListResponse
	var err error
	if atLeastOneViperStringFlagGiven("id", "partition", "size", "name", "project", "image", "hostname") ||
		atLeastOneViperStringSliceFlagGiven("tags") {
		ffr := &metalgo.FirewallFindRequest{
			MachineFindRequest: metalgo.MachineFindRequest{
				ID:                 viperString("id"),
				PartitionID:        viperString("partition"),
				SizeID:             viperString("size"),
				AllocationName:     viperString("name"),
				AllocationProject:  viperString("project"),
				AllocationImageID:  viperString("image"),
				AllocationHostname: viperString("hostname"),
				Tags:               viperStringSlice("tags"),
			},
		}
		if atLeastOneViperStringFlagGiven("mac") {
			ffr.NicsMacAddresses = []string{*viperString("mac")}
		}
		resp, err = driver.FirewallFind(ffr)
	} else {
		resp, err = driver.FirewallList()
	}
	if err != nil {
		return fmt.Errorf("firewall find error:%v", err)
	}
	return printer.Print(resp.Firewalls)
}

func firewallDescribe(driver *metalgo.Driver, args []string) error {
	firewallID, err := getMachineID(args)
	if err != nil {
		return err
	}
	resp, err := driver.FirewallGet(firewallID)
	if err != nil {
		return fmt.Errorf("firewall describe error:%v", err)
	}
	return detailer.Detail(resp.Firewall)
}

func firewallDestroy(driver *metalgo.Driver, args []string) error {
	firewallID, err := getMachineID(args)
	if err != nil {
		return err
	}

	resp, err := driver.MachineDelete(firewallID)
	if err != nil {
		return fmt.Errorf("firewall destroy error:%v", err)
	}
	return printer.Print(resp.Machine)
}
