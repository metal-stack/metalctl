package cmd

import (
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/go-openapi/strfmt"
	"github.com/metal-stack/metal-go/api/client/audit"
	"github.com/metal-stack/metal-go/api/models"
	"github.com/metal-stack/metal-go/test/client"
	"github.com/metal-stack/metal-lib/pkg/pointer"
	"github.com/metal-stack/metal-lib/pkg/testcommon"

	"github.com/stretchr/testify/mock"
)

var (
	auditTrace1 = &models.V1AuditResponse{
		Body:         `{"a": "b"}`,
		Component:    "example",
		Detail:       "GET",
		Error:        "",
		ForwardedFor: "192.168.2.2",
		Path:         "/v1/audit",
		Phase:        "response",
		RemoteAddr:   "192.168.2.1",
		Rqid:         "c40ad996-e1fd-4511-a7bf-418219cb8d91",
		StatusCode:   http.StatusOK,
		Tenant:       "a-tenant",
		Timestamp:    strfmt.DateTime(truncateToSeconds(testTime)),
		Type:         "http",
		User:         "a-user",
	}
	auditTrace2 = &models.V1AuditResponse{
		Body:         `{"c": "d"}`,
		Component:    "test",
		Detail:       "POST",
		Error:        "",
		ForwardedFor: "192.168.2.4",
		Path:         "/v1/audit",
		Phase:        "request",
		RemoteAddr:   "192.168.2.3",
		Rqid:         "b5817ef7-980a-41ef-9ed3-741a143870b0",
		StatusCode:   http.StatusForbidden,
		Tenant:       "b-tenant",
		Timestamp:    strfmt.DateTime(truncateToSeconds(testTime.Add(1 * time.Minute))),
		Type:         "http",
		User:         "b-user",
	}
)

func truncateToSeconds(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), 0, t.Location())
}

func Test_AuditCmd_MultiResult(t *testing.T) {
	tests := []*test[[]*models.V1AuditResponse]{
		{
			name: "list",
			cmd: func(want []*models.V1AuditResponse) []string {
				return []string{"audit", "list"}
			},
			mocks: &client.MetalMockFns{
				Audit: func(mock *mock.Mock) {
					beforeOneHour := strfmt.DateTime(testTime.Add(-1 * time.Hour))
					mock.On("FindAuditTraces", testcommon.MatchIgnoreContext(t, audit.NewFindAuditTracesParams().
						WithBody(&models.V1AuditFindRequest{Limit: 100, From: beforeOneHour})), nil).
						Return(&audit.FindAuditTracesOK{
							Payload: []*models.V1AuditResponse{
								auditTrace2,
								auditTrace1,
							},
						}, nil)
				},
			},
			want: []*models.V1AuditResponse{
				auditTrace2,
				auditTrace1,
			},
			wantTable: pointer.Pointer(`
TIME                  REQUEST ID                             DETAIL   PATH        CODE   TENANT     USER     PHASE    
May 19 01:03:03.000   b5817ef7-980a-41ef-9ed3-741a143870b0   POST     /v1/audit   403    b-tenant   b-user   request
May 19 01:02:03.000   c40ad996-e1fd-4511-a7bf-418219cb8d91   GET      /v1/audit   200    a-tenant   a-user   response 
`),
			wantWideTable: pointer.Pointer(`
TIME                  REQUEST ID                             DETAIL   PATH        CODE   TENANT     USER     PHASE      BODY
May 19 01:03:03.000   b5817ef7-980a-41ef-9ed3-741a143870b0   POST     /v1/audit   403    b-tenant   b-user   request    {"c": "d"}
May 19 01:02:03.000   c40ad996-e1fd-4511-a7bf-418219cb8d91   GET      /v1/audit   200    a-tenant   a-user   response   {"a": "b"}
`),
			template: pointer.Pointer(`{{ date "02/01/2006" .timestamp }} {{ .rqid }}`),
			wantTemplate: pointer.Pointer(`
19/05/2022 b5817ef7-980a-41ef-9ed3-741a143870b0
19/05/2022 c40ad996-e1fd-4511-a7bf-418219cb8d91
`),
			wantMarkdown: pointer.Pointer(`
|        TIME         |              REQUEST ID              | DETAIL |   PATH    | CODE |  TENANT  |  USER  |  PHASE   |
|---------------------|--------------------------------------|--------|-----------|------|----------|--------|----------|
| May 19 01:03:03.000 | b5817ef7-980a-41ef-9ed3-741a143870b0 | POST   | /v1/audit |  403 | b-tenant | b-user | request  |
| May 19 01:02:03.000 | c40ad996-e1fd-4511-a7bf-418219cb8d91 | GET    | /v1/audit |  200 | a-tenant | a-user | response |
`),
		},
		{
			name: "list with filters",
			cmd: func(want []*models.V1AuditResponse) []string {
				args := []string{"audit", "list",
					"--query", want[0].Body,
					"--path", want[0].Path,
					"--phase", want[0].Phase,
					"--request-id", want[0].Rqid,
					"--tenant", want[0].Tenant,
					"--user", want[0].User,
					"--from", want[0].Timestamp.String(),
					"--to", want[0].Timestamp.String(),
					"--error", want[0].Error,
					"--forwarded-for", want[0].ForwardedFor,
					"--remote-addr", want[0].RemoteAddr,
					"--detail", want[0].Detail,
					"--type", want[0].Type,
					"--component", want[0].Component,
					"--status-code", strconv.Itoa(int(want[0].StatusCode)),
					"--limit", "100",
				}
				assertExhaustiveArgs(t, args, "sort-by")
				return args
			},
			mocks: &client.MetalMockFns{
				Audit: func(mock *mock.Mock) {
					mock.On("FindAuditTraces", testcommon.MatchIgnoreContext(t, audit.NewFindAuditTracesParams().WithBody(&models.V1AuditFindRequest{
						Body:         auditTrace1.Body,
						Component:    auditTrace1.Component,
						Detail:       auditTrace1.Detail,
						Error:        auditTrace1.Error,
						ForwardedFor: auditTrace1.ForwardedFor,
						From:         auditTrace1.Timestamp,
						Limit:        100,
						Path:         auditTrace1.Path,
						Phase:        auditTrace1.Phase,
						RemoteAddr:   auditTrace1.RemoteAddr,
						Rqid:         auditTrace1.Rqid,
						StatusCode:   auditTrace1.StatusCode,
						Tenant:       auditTrace1.Tenant,
						To:           auditTrace1.Timestamp,
						Type:         auditTrace1.Type,
						User:         auditTrace1.User,
					})), nil).Return(&audit.FindAuditTracesOK{
						Payload: []*models.V1AuditResponse{
							auditTrace1,
						},
					}, nil)
				},
			},
			want: []*models.V1AuditResponse{
				auditTrace1,
			},
			wantTable: pointer.Pointer(`
TIME                  REQUEST ID                             DETAIL   PATH        CODE   TENANT     USER     PHASE    
May 19 01:02:03.000   c40ad996-e1fd-4511-a7bf-418219cb8d91   GET      /v1/audit   200    a-tenant   a-user   response		
`),
			wantWideTable: pointer.Pointer(`
TIME                  REQUEST ID                             DETAIL   PATH        CODE   TENANT     USER     PHASE      BODY       
May 19 01:02:03.000   c40ad996-e1fd-4511-a7bf-418219cb8d91   GET      /v1/audit   200    a-tenant   a-user   response   {"a": "b"}
		`),
			template: pointer.Pointer(`{{ date "02/01/2006" .timestamp }} {{ .rqid }}`),
			wantTemplate: pointer.Pointer(`
19/05/2022 c40ad996-e1fd-4511-a7bf-418219cb8d91
`),
			wantMarkdown: pointer.Pointer(`
|        TIME         |              REQUEST ID              | DETAIL |   PATH    | CODE |  TENANT  |  USER  |  PHASE   |
|---------------------|--------------------------------------|--------|-----------|------|----------|--------|----------|
| May 19 01:02:03.000 | c40ad996-e1fd-4511-a7bf-418219cb8d91 | GET    | /v1/audit |  200 | a-tenant | a-user | response |
`),
		},
	}
	for _, tt := range tests {
		tt.testCmd(t)
	}
}
