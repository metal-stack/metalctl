package cmd

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/go-openapi/strfmt"
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
		GenericCLI:      genericcli.NewGenericCLI(w).WithFS(c.fs),
		Singular:        "audit trace",
		Plural:          "audit traces",
		Description:     "show audit traces of the api. feature must be enabled on server-side.",
		Sorter:          sorters.AuditSorter(),
		DescribePrinter: func() printers.Printer { return c.describePrinter },
		ListPrinter:     func() printers.Printer { return c.listPrinter },
		ListCmdMutateFn: func(cmd *cobra.Command) {
			cmd.Flags().StringP("query", "q", "", "filters audit trace body payloads for the given text.")

			cmd.Flags().String("from", "1h", "start of range of the audit traces. e.g. 1h, 10m, 2006-01-02 15:04:05")
			cmd.Flags().String("to", "", "end of range of the audit traces. e.g. 1h, 10m, 2006-01-02 15:04:05")

			cmd.Flags().String("component", "", "component of the audit trace.")
			cmd.Flags().String("request-id", "", "request id of the audit trace.")
			cmd.Flags().String("type", "", "type of the audit trace. One of [http, grpc, event].")

			cmd.Flags().String("user", "", "user of the audit trace.")
			cmd.Flags().String("tenant", "", "tenant of the audit trace.")

			cmd.Flags().String("detail", "", "detail of the audit trace. An HTTP method, unary or stream")
			cmd.Flags().String("phase", "", "phase of the audit trace. One of [request, response, single, error, opened, closed]")

			cmd.Flags().String("path", "", "api path of the audit trace.")
			cmd.Flags().String("forwarded-for", "", "forwarded for of the audit trace.")
			cmd.Flags().String("remote-addr", "", "remote address of the audit trace.")

			cmd.Flags().String("error", "", "error of the audit trace.")
			cmd.Flags().Int32("status-code", 0, "HTTP status code of the audit trace.")

			cmd.Flags().Int64("limit", 100, "limit the number of audit traces.")

			genericcli.Must(cmd.RegisterFlagCompletionFunc("type", c.comp.AuditTypeCompletion))
			genericcli.Must(cmd.RegisterFlagCompletionFunc("phase", c.comp.AuditPhaseCompletion))
		},
		DescribeCmdMutateFn: func(cmd *cobra.Command) {
			cmd.Flags().Bool("prettify-body", false, "attempts to interpret the body as json and prettifies it")
			cmd.Flags().String("phase", "response", "phase of the audit trace. One of [request, response, single, error, opened, closed]")

			genericcli.Must(cmd.RegisterFlagCompletionFunc("phase", c.comp.AuditPhaseCompletion))
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
		Rqid:  id,
		Phase: viper.GetString("phase"),
	}), nil)
	if err != nil {
		return nil, err
	}
	if len(traces.Payload) == 0 {
		return nil, fmt.Errorf("no audit trace found with request id %s", id)
	}

	trace := traces.Payload[0]

	if viper.GetBool("prettify-body") {
		trimmed := strings.Trim(trace.Body, `"`)
		body := map[string]any{}
		err = json.Unmarshal([]byte(trimmed), &body)
		if err == nil {
			if pretty, err := json.MarshalIndent(body, "", "    "); err == nil {
				trace.Body = string(pretty)
			}
		}
	}

	return trace, nil
}

func (c auditCmd) List() ([]*models.V1AuditResponse, error) {
	fromDateTime, err := eventuallyRelativeDateTime(viper.GetString("from"))
	if err != nil {
		return nil, err
	}
	toDateTime, err := eventuallyRelativeDateTime(viper.GetString("to"))
	if err != nil {
		return nil, err
	}
	resp, err := c.client.Audit().FindAuditTraces(audit.NewFindAuditTracesParams().WithBody(&models.V1AuditFindRequest{
		Body:         viper.GetString("query"),
		From:         fromDateTime,
		To:           toDateTime,
		Component:    viper.GetString("component"),
		Rqid:         viper.GetString("request-id"),
		Type:         viper.GetString("type"),
		User:         viper.GetString("user"),
		Tenant:       viper.GetString("tenant"),
		Detail:       viper.GetString("detail"),
		Phase:        viper.GetString("phase"),
		Path:         viper.GetString("path"),
		ForwardedFor: viper.GetString("forwarded-for"),
		RemoteAddr:   viper.GetString("remote-addr"),
		Error:        viper.GetString("error"),
		StatusCode:   viper.GetInt32("status-code"),
		Limit:        viper.GetInt64("limit"),
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

func (c auditCmd) Convert(r *models.V1AuditResponse) (string, any, any, error) {
	return "", nil, nil, fmt.Errorf("not implemented for audit traces")
}

func eventuallyRelativeDateTime(s string) (strfmt.DateTime, error) {
	if s == "" {
		return strfmt.DateTime{}, nil
	}
	duration, err := strfmt.ParseDuration(s)
	if err == nil {
		return strfmt.DateTime(time.Now().Add(-duration)), nil
	}
	return strfmt.ParseDateTime(s)
}
