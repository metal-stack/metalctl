package cmd

import (
	"fmt"
	"log"

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

	firewallListCmd.Flags().StringVarP(&filterOpts.ID, "id", "", "", "ID to filter [optional]")
	firewallListCmd.Flags().StringVarP(&filterOpts.Partition, "partition", "", "", "partition to filter [optional]")
	firewallListCmd.Flags().StringVarP(&filterOpts.Size, "size", "", "", "size to filter [optional]")
	firewallListCmd.Flags().StringVarP(&filterOpts.Name, "name", "", "", "allocation name to filter [optional]")
	firewallListCmd.Flags().StringVarP(&filterOpts.Project, "project", "", "", "allocation project to filter [optional]")
	firewallListCmd.Flags().StringVarP(&filterOpts.Image, "image", "", "", "allocation image to filter [optional]")
	firewallListCmd.Flags().StringVarP(&filterOpts.Hostname, "hostname", "", "", "allocation hostname to filter [optional]")
	firewallListCmd.Flags().StringVarP(&filterOpts.Mac, "mac", "", "", "mac to filter [optional]")
	firewallListCmd.Flags().StringSliceVar(&filterOpts.Tags, "tags", []string{}, "tags to filter, use it like: --tags \"tag1,tag2\" or --tags \"tag3\".")
	err := firewallListCmd.RegisterFlagCompletionFunc("partition", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return partitionListCompletion(driver)
	})
	if err != nil {
		log.Fatal(err.Error())
	}
	err = firewallListCmd.RegisterFlagCompletionFunc("size", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return sizeListCompletion(driver)
	})
	if err != nil {
		log.Fatal(err.Error())
	}
	err = firewallListCmd.RegisterFlagCompletionFunc("project", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return projectListCompletion(driver)
	})
	if err != nil {
		log.Fatal(err.Error())
	}
	err = firewallListCmd.RegisterFlagCompletionFunc("id", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return machineListCompletion(driver)
	})
	if err != nil {
		log.Fatal(err.Error())
	}
	err = firewallListCmd.RegisterFlagCompletionFunc("image", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return imageListCompletion(driver)
	})
	if err != nil {
		log.Fatal(err.Error())
	}

	firewallCmd.AddCommand(firewallCreateCmd)
	firewallCmd.AddCommand(firewallListCmd)
	firewallCmd.AddCommand(firewallDestroyCmd)
	firewallCmd.AddCommand(firewallDescribeCmd)
	firewallCmd.AddCommand(firewallReserveCmd)
}

func firewallCreate(driver *metalgo.Driver) error {
	mcr, err := machineCreateRequest()
	if err != nil {
		return fmt.Errorf("firewall create error:%w", err)
	}

	fcr := &metalgo.FirewallCreateRequest{
		MachineCreateRequest: *mcr,
	}
	resp, err := driver.FirewallCreate(fcr)
	if err != nil {
		return err
	}
	return printer.Print(resp.Firewall)
}

func firewallList(driver *metalgo.Driver) error {
	var resp *metalgo.FirewallListResponse
	var err error
	if atLeastOneViperStringFlagGiven("id", "partition", "size", "name", "project", "image", "hostname") ||
		atLeastOneViperStringSliceFlagGiven("tags") {
		ffr := &metalgo.FirewallFindRequest{}
		if filterOpts.ID != "" {
			ffr.ID = &filterOpts.ID
		}
		if filterOpts.Partition != "" {
			ffr.PartitionID = &filterOpts.Partition
		}
		if filterOpts.Size != "" {
			ffr.SizeID = &filterOpts.Size
		}
		if filterOpts.Name != "" {
			ffr.AllocationName = &filterOpts.Name
		}
		if filterOpts.Project != "" {
			ffr.AllocationProject = &filterOpts.Project
		}
		if filterOpts.Image != "" {
			ffr.AllocationImageID = &filterOpts.Image
		}
		if filterOpts.Hostname != "" {
			ffr.AllocationHostname = &filterOpts.Hostname
		}
		if filterOpts.Hostname != "" {
			ffr.AllocationHostname = &filterOpts.Hostname
		}
		if filterOpts.Mac != "" {
			ffr.NicsMacAddresses = []string{filterOpts.Mac}
		}
		if len(filterOpts.Tags) > 0 {
			ffr.Tags = filterOpts.Tags
		}
		resp, err = driver.FirewallFind(ffr)
	} else {
		resp, err = driver.FirewallList()
	}
	if err != nil {
		return err
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
		return err
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
		return err
	}
	return printer.Print(resp.Machine)
}
