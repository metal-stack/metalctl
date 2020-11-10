package cmd

import (
	"fmt"
	"log"
	"net/http"

	v1 "github.com/metal-stack/masterdata-api/api/rest/v1"
	metalgo "github.com/metal-stack/metal-go"
	projectmodel "github.com/metal-stack/metal-go/api/client/project"
	"github.com/metal-stack/metal-go/api/models"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
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
	projectCreateCmd = &cobra.Command{
		Use:   "create",
		Short: "create a project",
		RunE: func(cmd *cobra.Command, args []string) error {
			return projectCreate()
		},
		PreRun: bindPFlags,
	}
	projectDeleteCmd = &cobra.Command{
		Use:     "remove <projectID>",
		Aliases: []string{"rm", "delete"},
		Short:   "delete a project",
		RunE: func(cmd *cobra.Command, args []string) error {
			return projectDelete(args)
		},
		PreRun: bindPFlags,
	}
	projectApplyCmd = &cobra.Command{
		Use:   "apply",
		Short: "create/update a project",
		RunE: func(cmd *cobra.Command, args []string) error {
			return projectApply()
		},
		PreRun: bindPFlags,
	}
	projectEditCmd = &cobra.Command{
		Use:   "edit <projectID>",
		Short: "edit a project",
		RunE: func(cmd *cobra.Command, args []string) error {
			return projectEdit(args)
		},
		PreRun: bindPFlags,
	}
)

func init() {
	projectCreateCmd.Flags().String("name", "", "name of the project, max 10 characters. [required]")
	projectCreateCmd.Flags().String("description", "", "description of the project. [required]")
	projectCreateCmd.Flags().String("tenant", "", "create project for given tenant")
	projectCreateCmd.Flags().StringSlice("label", nil, "add initial label, can be given multiple times to add multiple labels, e.g. --label=foo --label=bar")
	projectCreateCmd.Flags().StringSlice("annotation", nil, "add initial annotation, must be in the form of key=value, can be given multiple times to add multiple annotations, e.g. --annotation key=value --annotation foo=bar")
	projectCreateCmd.Flags().Int32("cluster-quota", 0, "cluster quota")
	projectCreateCmd.Flags().Int32("machine-quota", 0, "machine quota")
	projectCreateCmd.Flags().Int32("ip-quota", 0, "ip quota")
	err := projectCreateCmd.MarkFlagRequired("name")
	if err != nil {
		log.Fatal(err.Error())
	}

	projectApplyCmd.Flags().StringP("file", "f", "", `filename of the create or update request in yaml format, or - for stdin.
	Example project update:

	# cloudctl project describe project1 -o yaml > project1.yaml
	# vi project1.yaml
	## either via stdin
	# cat project1.yaml | cloudctl project apply -f -
	## or via file
	# cloudctl project apply -f project1.yaml
	`)
	projectListCmd.Flags().StringP("name", "", "", "Name of the project.")
	projectListCmd.Flags().StringP("id", "", "", "ID of the project.")
	projectListCmd.Flags().StringP("tenant", "", "", "tenant of this project.")

	projectCmd.AddCommand(projectCreateCmd)
	projectCmd.AddCommand(projectDescribeCmd)
	projectCmd.AddCommand(projectDeleteCmd)
	projectCmd.AddCommand(projectListCmd)
	projectCmd.AddCommand(projectApplyCmd)
	projectCmd.AddCommand(projectEditCmd)

	viper.BindPFlags(projectListCmd.Flags())
}

func projectList(driver *metalgo.Driver) error {
	if atLeastOneViperStringFlagGiven("id", "name", "tenant") {
		pfr := v1.ProjectFindRequest{}
		id := viper.GetString("id")
		name := viper.GetString("name")
		tenantID := viper.GetString("tenant")

		if id != "" {
			pfr.Id = &id
		}
		if name != "" {
			pfr.Name = &name
		}
		if tenantID != "" {
			pfr.TenantId = &tenantID
		}
		resp, err := driver.ProjectFind(pfr)
		if err != nil {
			return err
		}
		return printer.Print(resp.Project)
	}
	resp, err := driver.ProjectList()
	if err != nil {
		return err
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
		return err
	}
	return detailer.Detail(resp.Project)
}

func projectCreate() error {
	tenant := viper.GetString("tenant")
	name := viper.GetString("name")
	desc := viper.GetString("description")
	labels := viper.GetStringSlice("label")
	as := viper.GetStringSlice("annotation")
	var (
		clusterQuota, machineQuota, ipQuota *v1.Quota
	)
	if viper.IsSet("cluster-quota") {
		q := viper.GetInt32("cluster-quota")
		clusterQuota = &v1.Quota{Quota: &q}
	}
	if viper.IsSet("machine-quota") {
		q := viper.GetInt32("machine-quota")
		machineQuota = &v1.Quota{Quota: &q}
	}
	if viper.IsSet("ip-quota") {
		q := viper.GetInt32("ip-quota")
		ipQuota = &v1.Quota{Quota: &q}
	}

	annotations, err := annotationsAsMap(as)
	if err != nil {
		return err
	}

	p := v1.Project{
		Name:        name,
		Description: desc,
		TenantId:    tenant,
		Quotas: &v1.QuotaSet{
			Cluster: clusterQuota,
			Machine: machineQuota,
			Ip:      ipQuota,
		},
		Meta: &v1.Meta{
			Kind:        "Project",
			Apiversion:  "v1",
			Annotations: annotations,
			Labels:      labels,
		},
	}
	pcr := v1.ProjectCreateRequest{
		Project: p,
	}

	response, err := driver.ProjectCreate(pcr)
	if err != nil {
		return err
	}

	return printer.Print(response.Project)
}

func projectApply() error {
	var pars []v1.Project
	var par v1.Project
	err := readFrom(viper.GetString("file"), &par, func(data interface{}) {
		doc := data.(*v1.Project)
		pars = append(pars, *doc)
		// the request needs to be renewed as otherwise the pointers in the request struct will
		// always point to same last value in the multi-document loop
		par = v1.Project{}
	})
	if err != nil {
		return err
	}
	var response []*models.V1ProjectResponse
	for _, par := range pars {
		if par.Meta.Id == "" {
			resp, err := driver.ProjectCreate(v1.ProjectCreateRequest{Project: par})
			if err != nil {
				return err
			}
			response = append(response, resp.Project)
			continue
		}

		resp, err := driver.ProjectGet(par.Meta.Id)
		if err != nil {
			switch e := err.(type) {
			case *projectmodel.FindProjectDefault:
				if e.Code() != http.StatusNotFound {
					return err
				}
			default:
				return err
			}
		}
		if resp.Project == nil {
			resp, err := driver.ProjectCreate(v1.ProjectCreateRequest{Project: par})
			if err != nil {
				return err
			}
			response = append(response, resp.Project)
			continue
		}

		resp, err = driver.ProjectUpdate(v1.ProjectUpdateRequest{Project: par})
		if err != nil {
			return err
		}
		response = append(response, resp.Project)
	}
	return printer.Print(response)
}

func projectEdit(args []string) error {
	id, err := projectID("edit", args)
	if err != nil {
		return err
	}

	getFunc := func(id string) ([]byte, error) {
		resp, err := driver.ProjectGet(id)
		if err != nil {
			return nil, err
		}
		content, err := yaml.Marshal(resp.Project)
		if err != nil {
			return nil, err
		}
		return content, nil
	}
	updateFunc := func(filename string) error {
		purs, err := readProjectUpdateRequests(filename)
		if err != nil {
			return err
		}
		if len(purs) != 1 {
			return fmt.Errorf("project update error more or less than one project given:%d", len(purs))
		}
		uresp, err := driver.ProjectUpdate(v1.ProjectUpdateRequest{Project: purs[0]})
		if err != nil {
			return err
		}
		return printer.Print(uresp.Project)
	}

	return edit(id, getFunc, updateFunc)
}

func projectDelete(args []string) error {
	id, err := projectID("delete", args)
	if err != nil {
		return err
	}

	response, err := driver.ProjectDelete(id)
	if err != nil {
		return err
	}

	return printer.Print(response.Project)
}

func projectID(verb string, args []string) (string, error) {
	if len(args) == 0 {
		return "", fmt.Errorf("project %s requires projectID as argument", verb)
	}
	if len(args) == 1 {
		return args[0], nil
	}
	return "", fmt.Errorf("project %s requires exactly one projectID as argument", verb)
}

func readProjectUpdateRequests(filename string) ([]v1.Project, error) {
	var pcrs []v1.Project
	var pcr v1.Project
	err := readFrom(filename, &pcr, func(data interface{}) {
		doc := data.(*v1.Project)
		pcrs = append(pcrs, *doc)
	})
	if err != nil {
		return pcrs, err
	}
	if len(pcrs) != 1 {
		return pcrs, fmt.Errorf("project update error more or less than one project given:%d", len(pcrs))
	}
	return pcrs, nil
}
