package cmd

import (
	metalgo "github.com/metal-stack/metal-go"
	"github.com/metal-stack/metal-lib/pkg/genericcli"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// TODO: API responses are much different from the rest and it does not work well with generic cli

type firmwareTask int

const (
	upload firmwareTask = iota
	remove
)

func newFirmwareCmd(c *config) *cobra.Command {
	firmwareCmd := &cobra.Command{
		Use:   "firmware",
		Short: "manage firmwares",
		Long:  "list, upload and remove firmwares.",
	}

	firmwareListCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "list firmwares",
		Long:    "lists all available firmwares matching the given criteria.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.firmwareList()
		},
		PreRun: bindPFlags,
	}

	firmwareUploadCmd := &cobra.Command{
		Use:   "upload",
		Short: "upload a firmware",
	}

	firmwareUploadBiosCmd := &cobra.Command{
		Use:   "bios <file>",
		Short: "upload a BIOS firmware",
		Long:  "the given BIOS firmware file will be uploaded and tagged as given revision.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.firmwareUploadBios(args)
		},
		PreRun: bindPFlags,
	}

	firmwareUploadBmcCmd := &cobra.Command{
		Use:   "bmc <file>",
		Short: "upload a BMC firmware",
		Long:  "the given BMC firmware file will be uploaded and tagged as given revision.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.firmwareUploadBmc(args)
		},
		PreRun: bindPFlags,
	}

	firmwareRemoveCmd := &cobra.Command{
		Use:     "delete",
		Aliases: []string{"destroy", "rm", "remove"},
		Short:   "delete a firmware",
		Long:    "deletes the specified firmware.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.firmwareRemove(args)
		},
		PreRun: bindPFlags,
	}

	firmwareListCmd.Flags().StringP("kind", "", "", "the firmware kind [bmc|bios]")
	firmwareListCmd.Flags().StringP("vendor", "", "", "the vendor")
	firmwareListCmd.Flags().StringP("board", "", "", "the board type")
	firmwareListCmd.Flags().StringP("machineid", "", "", "the machine id (ignores vendor and board flags)")
	must(firmwareListCmd.RegisterFlagCompletionFunc("kind", c.comp.FirmwareKindCompletion))
	must(firmwareListCmd.RegisterFlagCompletionFunc("vendor", c.comp.FirmwareVendorCompletion))
	must(firmwareListCmd.RegisterFlagCompletionFunc("board", c.comp.FirmwareBoardCompletion))
	firmwareCmd.AddCommand(firmwareListCmd)

	firmwareUploadBiosCmd.Flags().StringP("vendor", "", "", "the vendor (required)")
	firmwareUploadBiosCmd.Flags().StringP("board", "", "", "the board type (required)")
	firmwareUploadBiosCmd.Flags().StringP("revision", "", "", "the BIOS firmware revision (required)")
	must(firmwareUploadBiosCmd.MarkFlagRequired("vendor"))
	must(firmwareUploadBiosCmd.MarkFlagRequired("board"))
	must(firmwareUploadBiosCmd.MarkFlagRequired("revision"))
	must(firmwareUploadBiosCmd.RegisterFlagCompletionFunc("vendor", c.comp.FirmwareVendorCompletion))
	must(firmwareUploadBiosCmd.RegisterFlagCompletionFunc("board", c.comp.FirmwareBoardCompletion))
	firmwareUploadCmd.AddCommand(firmwareUploadBiosCmd)

	firmwareUploadBmcCmd.Flags().StringP("vendor", "", "", "the vendor (required)")
	firmwareUploadBmcCmd.Flags().StringP("board", "", "", "the board type (required)")
	firmwareUploadBmcCmd.Flags().StringP("revision", "", "", "the BMC firmware revision (required)")
	must(firmwareUploadBmcCmd.MarkFlagRequired("vendor"))
	must(firmwareUploadBmcCmd.MarkFlagRequired("board"))
	must(firmwareUploadBmcCmd.MarkFlagRequired("revision"))
	must(firmwareUploadBmcCmd.RegisterFlagCompletionFunc("vendor", c.comp.FirmwareVendorCompletion))
	must(firmwareUploadBmcCmd.RegisterFlagCompletionFunc("board", c.comp.FirmwareBoardCompletion))
	firmwareUploadCmd.AddCommand(firmwareUploadBmcCmd)

	firmwareCmd.AddCommand(firmwareUploadCmd)

	firmwareRemoveCmd.Flags().StringP("kind", "", "", "the firmware kind [bmc|bios] (required)")
	firmwareRemoveCmd.Flags().StringP("vendor", "", "", "the vendor (required)")
	firmwareRemoveCmd.Flags().StringP("board", "", "", "the board type (required)")
	firmwareRemoveCmd.Flags().StringP("revision", "", "", "the firmware revision (required)")
	must(firmwareRemoveCmd.MarkFlagRequired("kind"))
	must(firmwareRemoveCmd.MarkFlagRequired("vendor"))
	must(firmwareRemoveCmd.MarkFlagRequired("board"))
	must(firmwareRemoveCmd.MarkFlagRequired("revision"))
	must(firmwareRemoveCmd.RegisterFlagCompletionFunc("kind", c.comp.FirmwareKindCompletion))
	must(firmwareRemoveCmd.RegisterFlagCompletionFunc("vendor", c.comp.FirmwareVendorCompletion))
	must(firmwareRemoveCmd.RegisterFlagCompletionFunc("board", c.comp.FirmwareBoardCompletion))
	must(firmwareRemoveCmd.RegisterFlagCompletionFunc("revision", c.comp.FirmwareRevisionCompletion))
	firmwareCmd.AddCommand(firmwareRemoveCmd)

	return firmwareCmd
}

func (c *config) firmwareList() error {
	var err error
	var resp *metalgo.FirmwaresResponse

	kind := metalgo.FirmwareKind(viper.GetString("kind"))
	id := viper.GetString("machineid")

	switch id {
	case "":
		vendor := viper.GetString("vendor")
		board := viper.GetString("board")
		resp, err = c.driver.ListFirmwares(kind, vendor, board)
	default:
		resp, err = c.driver.MachineListFirmwares(kind, id)
	}
	if err != nil {
		return err
	}

	return newPrinterFromCLI().Print(resp.Firmwares)
}

func (c *config) firmwareUploadBios(args []string) error {
	return c.manageFirmware(upload, metalgo.Bios, args)
}

func (c *config) firmwareUploadBmc(args []string) error {
	return c.manageFirmware(upload, metalgo.Bmc, args)
}

func (c *config) firmwareRemove(args []string) error {
	return c.manageFirmware(remove, metalgo.Bios, args)
}

func (c *config) manageFirmware(task firmwareTask, kind metalgo.FirmwareKind, args []string) error {
	revision := viper.GetString("revision")
	vendor := viper.GetString("vendor")
	board := viper.GetString("board")

	var err error
	switch task {
	case upload:
		var file string
		file, err = genericcli.GetExactlyOneArg(args)
		if err != nil {
			return err
		}
		_, err = c.driver.UploadFirmware(kind, vendor, board, revision, file)
	case remove:
		_, err = c.driver.RemoveFirmware(kind, vendor, board, revision)
	}
	return err
}
