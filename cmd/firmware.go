package cmd

import (
	"fmt"
	"log"

	metalgo "github.com/metal-stack/metal-go"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type firmwareTask int

const (
	upload firmwareTask = iota
	remove
)

var (
	firmwareCmd = &cobra.Command{
		Use:   "firmware",
		Short: "manage firmwares",
		Long:  "list, upload and remove firmwares.",
	}

	firmwareListCmd = &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "list firmwares",
		Long:    "lists all available firmwares matching the given criteria.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return firmwareList(driver)
		},
		PreRun: bindPFlags,
	}

	firmwareUploadCmd = &cobra.Command{
		Use:   "upload",
		Short: "upload a firmware",
	}

	firmwareUploadBiosCmd = &cobra.Command{
		Use:   "bios <file>",
		Short: "upload a BIOS firmware",
		Long:  "the given BIOS firmware file will be uploaded and tagged as given revision.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return firmwareUploadBios(driver, args)
		},
		PreRun: bindPFlags,
	}

	firmwareUploadBmcCmd = &cobra.Command{
		Use:   "bmc <file>",
		Short: "upload a BMC firmware",
		Long:  "the given BMC firmware file will be uploaded and tagged as given revision.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return firmwareUploadBmc(driver, args)
		},
		PreRun: bindPFlags,
	}

	firmwareRemoveCmd = &cobra.Command{
		Use:     "remove",
		Aliases: []string{"rm", "delete", "del"},
		Short:   "remove a firmware",
		Long:    "removes the specified firmware.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return firmwareRemove(driver, args)
		},
		PreRun: bindPFlags,
	}
)

func init() {
	firmwareListCmd.Flags().StringP("kind", "", "", "the firmware kind [bmc|bios]")
	firmwareListCmd.Flags().StringP("vendor", "", "", "the vendor")
	firmwareListCmd.Flags().StringP("board", "", "", "the board type")
	firmwareListCmd.Flags().StringP("machineid", "", "", "the machine id (ignores vendor and board flags)")
	err := firmwareListCmd.RegisterFlagCompletionFunc("kind", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return firmwareKindCompletion(driver)
	})
	if err != nil {
		log.Fatal(err.Error())
	}
	err = firmwareListCmd.RegisterFlagCompletionFunc("vendor", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return firmwareVendorCompletion(driver)
	})
	if err != nil {
		log.Fatal(err.Error())
	}
	err = firmwareListCmd.RegisterFlagCompletionFunc("board", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return firmwareBoardCompletion(driver)
	})
	if err != nil {
		log.Fatal(err.Error())
	}
	firmwareCmd.AddCommand(firmwareListCmd)

	firmwareUploadBiosCmd.Flags().StringP("vendor", "", "", "the vendor (required)")
	firmwareUploadBiosCmd.Flags().StringP("board", "", "", "the board type (required)")
	firmwareUploadBiosCmd.Flags().StringP("revision", "", "", "the BIOS firmware revision (required)")
	err = firmwareUploadBiosCmd.MarkFlagRequired("vendor")
	if err != nil {
		log.Fatal(err.Error())
	}
	err = firmwareUploadBiosCmd.MarkFlagRequired("board")
	if err != nil {
		log.Fatal(err.Error())
	}
	err = firmwareUploadBiosCmd.MarkFlagRequired("revision")
	if err != nil {
		log.Fatal(err.Error())
	}
	err = firmwareUploadBiosCmd.RegisterFlagCompletionFunc("vendor", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return firmwareVendorCompletion(driver)
	})
	if err != nil {
		log.Fatal(err.Error())
	}
	err = firmwareUploadBiosCmd.RegisterFlagCompletionFunc("board", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return firmwareBoardCompletion(driver)
	})
	if err != nil {
		log.Fatal(err.Error())
	}
	firmwareUploadCmd.AddCommand(firmwareUploadBiosCmd)

	firmwareUploadBmcCmd.Flags().StringP("vendor", "", "", "the vendor (required)")
	firmwareUploadBmcCmd.Flags().StringP("board", "", "", "the board type (required)")
	firmwareUploadBmcCmd.Flags().StringP("revision", "", "", "the BMC firmware revision (required)")
	err = firmwareUploadBmcCmd.MarkFlagRequired("vendor")
	if err != nil {
		log.Fatal(err.Error())
	}
	err = firmwareUploadBmcCmd.MarkFlagRequired("board")
	if err != nil {
		log.Fatal(err.Error())
	}
	err = firmwareUploadBmcCmd.MarkFlagRequired("revision")
	if err != nil {
		log.Fatal(err.Error())
	}
	err = firmwareUploadBmcCmd.RegisterFlagCompletionFunc("vendor", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return firmwareVendorCompletion(driver)
	})
	if err != nil {
		log.Fatal(err.Error())
	}
	err = firmwareUploadBmcCmd.RegisterFlagCompletionFunc("board", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return firmwareBoardCompletion(driver)
	})
	if err != nil {
		log.Fatal(err.Error())
	}
	firmwareUploadCmd.AddCommand(firmwareUploadBmcCmd)

	firmwareCmd.AddCommand(firmwareUploadCmd)

	firmwareRemoveCmd.Flags().StringP("kind", "", "", "the firmware kind [bmc|bios] (required)")
	firmwareRemoveCmd.Flags().StringP("vendor", "", "", "the vendor (required)")
	firmwareRemoveCmd.Flags().StringP("board", "", "", "the board type (required)")
	firmwareRemoveCmd.Flags().StringP("revision", "", "", "the firmware revision (required)")
	err = firmwareRemoveCmd.MarkFlagRequired("kind")
	if err != nil {
		log.Fatal(err.Error())
	}
	err = firmwareRemoveCmd.MarkFlagRequired("vendor")
	if err != nil {
		log.Fatal(err.Error())
	}
	err = firmwareRemoveCmd.MarkFlagRequired("board")
	if err != nil {
		log.Fatal(err.Error())
	}
	err = firmwareRemoveCmd.MarkFlagRequired("revision")
	if err != nil {
		log.Fatal(err.Error())
	}
	err = firmwareRemoveCmd.RegisterFlagCompletionFunc("kind", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return firmwareKindCompletion(driver)
	})
	if err != nil {
		log.Fatal(err.Error())
	}
	err = firmwareRemoveCmd.RegisterFlagCompletionFunc("vendor", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return firmwareVendorCompletion(driver)
	})
	if err != nil {
		log.Fatal(err.Error())
	}
	err = firmwareRemoveCmd.RegisterFlagCompletionFunc("board", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return firmwareBoardCompletion(driver)
	})
	if err != nil {
		log.Fatal(err.Error())
	}
	err = firmwareRemoveCmd.RegisterFlagCompletionFunc("revision", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return firmwareRevisionCompletion(driver)
	})
	if err != nil {
		log.Fatal(err.Error())
	}
	firmwareCmd.AddCommand(firmwareRemoveCmd)
}

func firmwareList(driver *metalgo.Driver) error {
	var err error
	var resp *metalgo.FirmwaresResponse

	kind := metalgo.FirmwareKind(viper.GetString("kind"))
	id := viper.GetString("machineid")

	switch id {
	case "":
		vendor := viper.GetString("vendor")
		board := viper.GetString("board")
		resp, err = driver.ListFirmwares(kind, vendor, board)
	default:
		resp, err = driver.MachineListFirmwares(kind, id)
	}
	if err != nil {
		return err
	}

	return printer.Print(resp.Firmwares)
}

func firmwareUploadBios(driver *metalgo.Driver, args []string) error {
	return manageFirmware(upload, driver, metalgo.Bios, args)
}

func firmwareUploadBmc(driver *metalgo.Driver, args []string) error {
	return manageFirmware(upload, driver, metalgo.Bmc, args)
}

func firmwareRemove(driver *metalgo.Driver, args []string) error {
	return manageFirmware(remove, driver, metalgo.Bios, args)
}

func manageFirmware(task firmwareTask, driver *metalgo.Driver, kind metalgo.FirmwareKind, args []string) error {
	revision := viper.GetString("revision")
	vendor := viper.GetString("vendor")
	board := viper.GetString("board")

	var err error
	switch task {
	case upload:
		if len(args) < 1 {
			return fmt.Errorf("no firmware file given")
		}
		_, err = driver.UploadFirmware(kind, vendor, board, revision, args[0])
	case remove:
		_, err = driver.RemoveFirmware(kind, vendor, board, revision)
	}
	return err
}

func firmwareKindCompletion(driver *metalgo.Driver) ([]string, cobra.ShellCompDirective) {
	return []string{string(metalgo.Bmc), string(metalgo.Bios)}, cobra.ShellCompDirectiveNoFileComp
}

func firmwareVendorCompletion(driver *metalgo.Driver) ([]string, cobra.ShellCompDirective) {
	resp, err := driver.ListFirmwares("", "", "")
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	var vendors []string
	for _, vv := range resp.Firmwares.Revisions {
		for v := range vv.VendorRevisions {
			vendors = append(vendors, v)
		}
	}
	return vendors, cobra.ShellCompDirectiveNoFileComp
}

func firmwareBoardCompletion(driver *metalgo.Driver) ([]string, cobra.ShellCompDirective) {
	resp, err := driver.ListFirmwares("", "", "")
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	var boards []string
	for _, vv := range resp.Firmwares.Revisions {
		for _, bb := range vv.VendorRevisions {
			for b := range bb.BoardRevisions {
				boards = append(boards, b)
			}
		}
	}
	return boards, cobra.ShellCompDirectiveNoFileComp
}

func firmwareRevisionCompletion(driver *metalgo.Driver) ([]string, cobra.ShellCompDirective) {
	resp, err := driver.ListFirmwares("", "", "")
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	var revisions []string
	for _, vv := range resp.Firmwares.Revisions {
		for _, bb := range vv.VendorRevisions {
			for _, rr := range bb.BoardRevisions {
				revisions = append(revisions, rr...)
			}
		}
	}
	return revisions, cobra.ShellCompDirectiveNoFileComp
}
