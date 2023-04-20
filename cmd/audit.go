package cmd

import (
	"fmt"

	"github.com/metal-stack/metal-go/api/client/audit"
	"github.com/metal-stack/metal-go/api/models"
	"github.com/metal-stack/metal-lib/pkg/genericcli"
	"github.com/metal-stack/metal-lib/pkg/genericcli/printers"
	"github.com/metal-stack/metalctl/cmd/sorters"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type auditCmd struct {
	*config
}

func newAuditCmd(c *config) *cobra.Command {
	w := auditCmd{
		config: c,
	}

	cmdsConfig := &genericcli.CmdsConfig[any, any, *models.V1AuditResponse]{
		BinaryName:      binaryName,
		GenericCLI:      genericcli.NewGenericCLI[any, any, *models.V1AuditResponse](w).WithFS(c.fs),
		Singular:        "audit trace",
		Plural:          "audit traces",
		Description:     "show audit traces of the api. feature must be enabled on server-side.",
		Sorter:          sorters.AuditSorter(),
		DescribePrinter: func() printers.Printer { return c.describePrinter },
		ListPrinter:     func() printers.Printer { return c.listPrinter },
		ListCmdMutateFn: func(cmd *cobra.Command) {
			cmd.Flags().StringP("query", "q", "", "filters audit trace payloads for the given text.")
			cmd.Flags().String("path", "", "api path of the audit trace.")
			cmd.Flags().String("phase", "", "phase of the audit trace.")
			cmd.Flags().String("request-id", "", "request id of the audit trace.")
			cmd.Flags().String("tenant", "", "tenant of the audit trace.")
			cmd.Flags().String("user", "", "user of the audit trace.")
			cmd.Flags().String("from", "", "start time of the audit trace.")
			cmd.Flags().String("to", "", "end time of the audit trace.")
			cmd.Flags().String("detail", "", "detail of the audit trace.")
			cmd.Flags().String("component", "", "component of the audit trace.")
			cmd.Flags().Int64("limit", 100, "limit the number of audit traces.")
		},
		OnlyCmds: genericcli.OnlyCmds(
			genericcli.ListCmd,
			genericcli.DescribeCmd,
		),
	}

	return genericcli.NewCmds(cmdsConfig)
}

func (c auditCmd) Get(id string) (*models.V1AuditResponse, error) {
	traces, err := c.client.Audit().FindAuditTraces(audit.NewFindAuditTracesParams().WithBody(&models.V1AuditFindRequest{
		Rqid: id,
	}), nil)
	if err != nil {
		return nil, err
	}
	if len(traces.Payload) == 0 {
		return nil, fmt.Errorf("no audit trace found with request id %s", id)
	}
	return traces.Payload[0], nil
}

func (c auditCmd) List() ([]*models.V1AuditResponse, error) {
	resp, err := c.client.Audit().FindAuditTraces(audit.NewFindAuditTracesParams().WithBody(&models.V1AuditFindRequest{
		Body:         viper.GetString("query"),
		Component:    viper.GetString("component"),
		Detail:       viper.GetString("detail"),
		Error:        viper.GetString("error"),
		ForwardedFor: viper.GetString("forwarded-for"),
		// From:         strfmt.DateTime{},
		Limit:      viper.GetInt64("limit"),
		Path:       viper.GetString("path"),
		Phase:      viper.GetString("phase"),
		RemoteAddr: viper.GetString("remote-addr"),
		Rqid:       viper.GetString("request-id"),
		StatusCode: viper.GetInt32("status-code"),
		Tenant:     viper.GetString("tenant"),
		// To:           strfmt.DateTime{},
		Type: viper.GetString("type"),
		User: viper.GetString("user"),
	}), nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c auditCmd) Delete(id string) (*models.V1AuditResponse, error) {
	return nil, fmt.Errorf("not implemented for audit traces")
}

func (c auditCmd) Create(_ any) (*models.V1AuditResponse, error) {
	return nil, fmt.Errorf("not implemented for audit traces")
}

func (c auditCmd) Update(_ any) (*models.V1AuditResponse, error) {
	return nil, fmt.Errorf("not implemented for audit traces")
}

func (c auditCmd) ToCreate(_ *models.V1AuditResponse) (any, error) {
	return nil, fmt.Errorf("not implemented for audit traces")
}

func (c auditCmd) ToUpdate(_ *models.V1AuditResponse) (any, error) {
	return nil, fmt.Errorf("not implemented for audit traces")
}
