package cmd

import (
	"github.com/spf13/cobra"

	"github.com/metal-stack/updater"
)

func newUpdateCmd() *cobra.Command {
	updateCmd := &cobra.Command{
		Use:   "update",
		Short: "update the program",
	}
	updateCheckCmd := &cobra.Command{
		Use:   "check",
		Short: "check for update of the program",
		RunE: func(cmd *cobra.Command, args []string) error {
			u, err := updater.New("metal-stack", binaryName, binaryName)
			if err != nil {
				return err
			}
			return u.Check()
		},
	}
	updateDoCmd := &cobra.Command{
		Use:   "do",
		Short: "do the update of the program",
		RunE: func(cmd *cobra.Command, args []string) error {
			u, err := updater.New("metal-stack", binaryName, binaryName)
			if err != nil {
				return err
			}
			return u.Do()
		},
	}
	updateCmd.AddCommand(updateCheckCmd)
	updateCmd.AddCommand(updateDoCmd)
	return updateCmd
}
