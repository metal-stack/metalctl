package cmd

import (
	"github.com/spf13/cobra"

	"github.com/metal-stack/metal-go/api/client/version"
	"github.com/metal-stack/updater"
)

func newUpdateCmd(c *config) *cobra.Command {
	updateCmd := &cobra.Command{
		Use:   "update",
		Short: "update the program",
	}
	updateCheckCmd := &cobra.Command{
		Use:   "check",
		Short: "check for update of the program",
		RunE: func(cmd *cobra.Command, args []string) error {
			desired, err := getDesiredVersion(c)
			if err != nil {
				return err
			}
			u, err := updater.New("metal-stack", binaryName, binaryName, desired)
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
			desired, err := getDesiredVersion(c)
			if err != nil {
				return err
			}
			u, err := updater.New("metal-stack", binaryName, binaryName, desired)
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

func getDesiredVersion(c *config) (*string, error) {
	resp, err := c.client.Version().Info(version.NewInfoParams(), nil)
	if err != nil {
		return nil, err
	}
	if resp.Payload != nil && resp.Payload.MinClientVersion != nil {
		return resp.Payload.MinClientVersion, nil
	}
	return nil, nil
}
