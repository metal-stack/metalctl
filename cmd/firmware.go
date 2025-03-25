package cmd

import (
	"os"

	"github.com/go-openapi/runtime"
	"github.com/metal-stack/metal-go/api/client/firmware"
	"github.com/metal-stack/metal-go/api/models"
	"github.com/metal-stack/metal-lib/pkg/genericcli"
	"github.com/metal-stack/metalctl/pkg/api"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// TODO: API responses are much different from the rest and it does not work well with generic cli

type firmwareCmd struct {
	c *api.Config
}

func newFirmwareCmd(c *api.Config) *cobra.Command {
	w := &firmwareCmd{
		c: c,
	}

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
			return w.firmwareList()
		},
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
			return w.firmwareUploadBios(args)
		},
	}

	firmwareUploadBmcCmd := &cobra.Command{
		Use:   "bmc <file>",
		Short: "upload a BMC firmware",
		Long:  "the given BMC firmware file will be uploaded and tagged as given revision.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return w.firmwareUploadBmc(args)
		},
	}

	firmwareRemoveCmd := &cobra.Command{
		Use:     "delete",
		Aliases: []string{"destroy", "rm", "remove"},
		Short:   "delete a firmware",
		Long:    "deletes the specified firmware.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return w.firmwareRemove()
		},
	}

	firmwareListCmd.Flags().StringP("kind", "", "", "the firmware kind [bmc|bios]")
	firmwareListCmd.Flags().StringP("vendor", "", "", "the vendor")
	firmwareListCmd.Flags().StringP("board", "", "", "the board type")
	firmwareListCmd.Flags().StringP("machineid", "", "", "the machine id (ignores vendor and board flags)")
	genericcli.Must(firmwareListCmd.RegisterFlagCompletionFunc("kind", c.Comp.FirmwareKindCompletion))
	genericcli.Must(firmwareListCmd.RegisterFlagCompletionFunc("vendor", c.Comp.FirmwareVendorCompletion))
	genericcli.Must(firmwareListCmd.RegisterFlagCompletionFunc("board", c.Comp.FirmwareBoardCompletion))
	genericcli.Must(firmwareListCmd.RegisterFlagCompletionFunc("machineid", c.Comp.MachineListCompletion))
	firmwareCmd.AddCommand(firmwareListCmd)

	firmwareUploadBiosCmd.Flags().StringP("vendor", "", "", "the vendor (required)")
	firmwareUploadBiosCmd.Flags().StringP("board", "", "", "the board type (required)")
	firmwareUploadBiosCmd.Flags().StringP("revision", "", "", "the BIOS firmware revision (required)")
	genericcli.Must(firmwareUploadBiosCmd.MarkFlagRequired("vendor"))
	genericcli.Must(firmwareUploadBiosCmd.MarkFlagRequired("board"))
	genericcli.Must(firmwareUploadBiosCmd.MarkFlagRequired("revision"))
	genericcli.Must(firmwareUploadBiosCmd.RegisterFlagCompletionFunc("vendor", c.Comp.FirmwareVendorCompletion))
	genericcli.Must(firmwareUploadBiosCmd.RegisterFlagCompletionFunc("board", c.Comp.FirmwareBoardCompletion))
	firmwareUploadCmd.AddCommand(firmwareUploadBiosCmd)

	firmwareUploadBmcCmd.Flags().StringP("vendor", "", "", "the vendor (required)")
	firmwareUploadBmcCmd.Flags().StringP("board", "", "", "the board type (required)")
	firmwareUploadBmcCmd.Flags().StringP("revision", "", "", "the BMC firmware revision (required)")
	genericcli.Must(firmwareUploadBmcCmd.MarkFlagRequired("vendor"))
	genericcli.Must(firmwareUploadBmcCmd.MarkFlagRequired("board"))
	genericcli.Must(firmwareUploadBmcCmd.MarkFlagRequired("revision"))
	genericcli.Must(firmwareUploadBmcCmd.RegisterFlagCompletionFunc("vendor", c.Comp.FirmwareVendorCompletion))
	genericcli.Must(firmwareUploadBmcCmd.RegisterFlagCompletionFunc("board", c.Comp.FirmwareBoardCompletion))
	firmwareUploadCmd.AddCommand(firmwareUploadBmcCmd)

	firmwareCmd.AddCommand(firmwareUploadCmd)

	firmwareRemoveCmd.Flags().StringP("kind", "", "", "the firmware kind [bmc|bios] (required)")
	firmwareRemoveCmd.Flags().StringP("vendor", "", "", "the vendor (required)")
	firmwareRemoveCmd.Flags().StringP("board", "", "", "the board type (required)")
	firmwareRemoveCmd.Flags().StringP("revision", "", "", "the firmware revision (required)")
	genericcli.Must(firmwareRemoveCmd.MarkFlagRequired("kind"))
	genericcli.Must(firmwareRemoveCmd.MarkFlagRequired("vendor"))
	genericcli.Must(firmwareRemoveCmd.MarkFlagRequired("board"))
	genericcli.Must(firmwareRemoveCmd.MarkFlagRequired("revision"))
	genericcli.Must(firmwareRemoveCmd.RegisterFlagCompletionFunc("kind", c.Comp.FirmwareKindCompletion))
	genericcli.Must(firmwareRemoveCmd.RegisterFlagCompletionFunc("vendor", c.Comp.FirmwareVendorCompletion))
	genericcli.Must(firmwareRemoveCmd.RegisterFlagCompletionFunc("board", c.Comp.FirmwareBoardCompletion))
	genericcli.Must(firmwareRemoveCmd.RegisterFlagCompletionFunc("revision", c.Comp.FirmwareRevisionCompletion))
	firmwareCmd.AddCommand(firmwareRemoveCmd)

	return firmwareCmd
}

func (c *firmwareCmd) firmwareList() error {
	kind := viper.GetString("kind")
	board := viper.GetString("board")
	vendor := viper.GetString("vendor")
	id := viper.GetString("machineid")

	resp, err := c.c.Client.Firmware().ListFirmwares(firmware.NewListFirmwaresParams().WithKind(&kind).WithBoard(&board).WithVendor(&vendor).WithMachineID(&id), nil)
	if err != nil {
		return err
	}

	return c.c.ListPrinter.Print(resp.Payload)
}

func (c *firmwareCmd) firmwareUploadBios(args []string) error {
	return c.uploadFirmware(models.V1MachineUpdateFirmwareRequestKindBios, args)
}

func (c *firmwareCmd) firmwareUploadBmc(args []string) error {
	return c.uploadFirmware(models.V1MachineUpdateFirmwareRequestKindBmc, args)
}

func (c *firmwareCmd) firmwareRemove() error {
	kind := viper.GetString("kind")
	revision := viper.GetString("revision")
	vendor := viper.GetString("vendor")
	board := viper.GetString("board")

	_, err := c.c.Client.Firmware().RemoveFirmware(firmware.NewRemoveFirmwareParams().
		WithKind(kind).
		WithBoard(board).
		WithVendor(vendor).
		WithRevision(revision), nil)
	if err != nil {
		return err
	}

	return nil
}

func (c *firmwareCmd) uploadFirmware(kind string, args []string) error {
	revision := viper.GetString("revision")
	vendor := viper.GetString("vendor")
	board := viper.GetString("board")

	var err error

	var file string
	file, err = genericcli.GetExactlyOneArg(args)
	if err != nil {
		return err
	}

	reader, err := os.Open(file)
	if err != nil {
		return err
	}

	_, err = c.c.Client.Firmware().UploadFirmware(firmware.NewUploadFirmwareParams().
		WithKind(kind).
		WithBoard(board).
		WithVendor(vendor).
		WithRevision(revision).
		WithFile(runtime.NamedReader(revision, reader)), nil)

	return err

}
