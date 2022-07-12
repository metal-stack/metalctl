package completion

import (
	metalgo "github.com/metal-stack/metal-go"
	"github.com/metal-stack/metal-go/api/client/firmware"
	"github.com/metal-stack/metal-go/api/client/machine"
	"github.com/metal-stack/metal-lib/pkg/pointer"
	"github.com/spf13/cobra"
)

func (c *Completion) FirmwareKindCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return []string{string(metalgo.Bmc), string(metalgo.Bios)}, cobra.ShellCompDirectiveNoFileComp
}

func (c *Completion) FirmwareVendorCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	resp, err := c.client.Firmware().ListFirmwares(firmware.NewListFirmwaresParams(), nil)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	var vendors []string
	for _, vv := range resp.Payload.Revisions {
		for v := range vv.VendorRevisions {
			vendors = append(vendors, v)
		}
	}
	return vendors, cobra.ShellCompDirectiveNoFileComp
}

func (c *Completion) FirmwareBoardCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	resp, err := c.client.Firmware().ListFirmwares(firmware.NewListFirmwaresParams(), nil)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	var boards []string
	for _, vv := range resp.Payload.Revisions {
		for _, bb := range vv.VendorRevisions {
			for b := range bb.BoardRevisions {
				boards = append(boards, b)
			}
		}
	}
	return boards, cobra.ShellCompDirectiveNoFileComp
}
func (c *Completion) FirmwareRevisionCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return c.firmwareRevisions("", "")
}
func (c *Completion) FirmwareBiosRevisionCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) != 1 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	return c.firmwareRevisions(args[0], metalgo.Bios)
}
func (c *Completion) FirmwareBmcRevisionCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) != 1 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	return c.firmwareRevisions(args[0], metalgo.Bmc)
}

func (c *Completion) firmwareRevisions(machineID string, kind metalgo.FirmwareKind) ([]string, cobra.ShellCompDirective) {
	vendor := ""
	board := ""
	if machineID != "" {
		m, err := c.client.Machine().FindIPMIMachine(machine.NewFindIPMIMachineParams().WithID(machineID), nil)
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}
		board = m.Payload.Ipmi.Fru.BoardPartNumber
		vendor = m.Payload.Ipmi.Fru.BoardMfg
	}
	resp, err := c.client.Firmware().ListFirmwares(firmware.NewListFirmwaresParams().WithKind(pointer.Pointer(string(kind))).WithVendor(&vendor).WithBoard(&board), nil)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	var revisions []string
	for _, vv := range resp.Payload.Revisions {
		for _, bb := range vv.VendorRevisions {
			for _, rr := range bb.BoardRevisions {
				revisions = append(revisions, rr...)
			}
		}
	}
	return revisions, cobra.ShellCompDirectiveNoFileComp
}
