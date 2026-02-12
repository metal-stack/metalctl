package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

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
			desired, err := getMinimumClientVersion(c)
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
			var desired *string

			if !viper.IsSet("version") {
				var err error
				desired, err = getMinimumClientVersion(c)
				if err != nil {
					return err
				}
			}

			if viper.IsSet("version") && viper.GetString("version") != "latest" {
				desired = new(viper.GetString("version"))
			}

			u, err := updater.New("metal-stack", binaryName, binaryName, desired)
			if err != nil {
				return err
			}

			return u.Do()
		},
	}

	updateDoCmd.Flags().StringP("version", "v", "", `the version to update to, by default updates to the supported version, use "latest" to update to latest version`)

	updateCmd.AddCommand(updateCheckCmd)
	updateCmd.AddCommand(updateDoCmd)

	return updateCmd
}

func getMinimumClientVersion(c *config) (*string, error) {
	resp, err := c.client.Version().Info(version.NewInfoParams(), clientNoAuth())
	if err != nil {
		return nil, err
	}
	if resp.Payload != nil && resp.Payload.MinClientVersion != nil {
		return resp.Payload.MinClientVersion, nil
	}
	return nil, nil
}
