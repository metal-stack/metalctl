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
			cmd.Flags().StringP("query", "q", "", "request id of the audit trace.")
			cmd.Flags().String("path", "", "request id of the audit trace.")
			cmd.Flags().String("phase", "", "request id of the audit trace.")
			cmd.Flags().String("request-id", "", "request id of the audit trace.")
			cmd.Flags().String("tenant", "", "request id of the audit trace.")
			cmd.Flags().String("user", "", "request id of the audit trace.")
		},
		OnlyCmds: genericcli.OnlyCmds(
			genericcli.ListCmd,
		),
	}

	return genericcli.NewCmds(cmdsConfig)
}

func (c auditCmd) Get(id string) (*models.V1AuditResponse, error) {
	return nil, fmt.Errorf("not implemented for audit traces")
}

func (c auditCmd) List() ([]*models.V1AuditResponse, error) {
	resp, err := c.client.Audit().FindAuditTraces(audit.NewFindAuditTracesParams().WithBody(&models.V1AuditFindRequest{
		Body: viper.GetString("query"),
		// From:   strfmt.DateTime{},
		Path:   viper.GetString("path"),
		Phase:  viper.GetString("phase"),
		Rqid:   viper.GetString("request-id"),
		Tenant: viper.GetString("tenant"),
		// To:     strfmt.DateTime{},
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
