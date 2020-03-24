package cmd

import (
	"fmt"

	metalgo "github.com/metal-stack/metal-go"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	projectCmd = &cobra.Command{
		Use:   "project",
		Short: "manage projects",
		Long:  "a project groups multiple networks for a tenant",
	}

	projectListCmd = &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "list all projects",
		RunE: func(cmd *cobra.Command, args []string) error {
			return projectList(driver)
		},
	}
	projectDescribeCmd = &cobra.Command{
		Use:   "describe <projectID>",
		Short: "describe a project",
		RunE: func(cmd *cobra.Command, args []string) error {
			return projectDescribe(driver, args)
		},
	}
)

func init() {
	projectListCmd.Flags().StringP("name", "", "", "Name of the project.")
	projectListCmd.Flags().StringP("id", "", "", "ID of the project.")
	projectListCmd.Flags().StringP("tenant", "", "", "tenant of this project.")

	projectCmd.AddCommand(projectListCmd)
	projectCmd.AddCommand(projectDescribeCmd)
}

func projectList(driver *metalgo.Driver) error {
	if atLeastOneViperStringFlagGiven("id", "name", "tenant") {
		resp, err := driver.ProjectFind(metalgo.ProjectFindRequest{
			ID:     viper.GetString("id"),
			Name:   viper.GetString("name"),
			Tenant: viper.GetString("tenant"),
		})
		if err != nil {
			return fmt.Errorf("project list error:%v", err)
		}
		return printer.Print(resp.Project)
	}
	resp, err := driver.ProjectList()
	if err != nil {
		return fmt.Errorf("project list error:%v", err)
	}
	return printer.Print(resp.Project)
}

func projectDescribe(driver *metalgo.Driver, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("no project ID given")
	}
	projectID := args[0]
	resp, err := driver.ProjectGet(projectID)
	if err != nil {
		return fmt.Errorf("project describe error:%v", err)
	}
	return detailer.Detail(resp.Project)
}
