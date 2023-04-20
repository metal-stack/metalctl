package cmd

import (
	"net/http"
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
		Body:          `{"a": "b"}`,
		Code:          http.StatusOK,
		ForwardedFor:  "192.168.2.2",
		Path:          "/v1/audit",
		Phase:         "response",
		RemoteAddress: "192.168.2.1",
		Rqid:          "c40ad996-e1fd-4511-a7bf-418219cb8d91",
		Tenant:        "a-tenant",
		Timestamp:     strfmt.DateTime(testTime),
		User:          "a-user",
	}
	auditTrace2 = &models.V1AuditResponse{
		Body:          `{"c": "d"}`,
		Code:          http.StatusForbidden,
		ForwardedFor:  "192.168.2.4",
		Path:          "/v1/audit",
		Phase:         "request",
		RemoteAddress: "192.168.2.3",
		Rqid:          "b5817ef7-980a-41ef-9ed3-741a143870b0",
		Tenant:        "b-tenant",
		Timestamp:     strfmt.DateTime(testTime.Add(1 * time.Minute)),
		User:          "b-user",
	}
)

func Test_AuditCmd_MultiResult(t *testing.T) {
	tests := []*test[[]*models.V1AuditResponse]{
		{
			name: "list",
			cmd: func(want []*models.V1AuditResponse) []string {
				return []string{"audit", "list"}
			},
			mocks: &client.MetalMockFns{
				Audit: func(mock *mock.Mock) {
					mock.On("FindAuditTraces", testcommon.MatchIgnoreContext(t, audit.NewFindAuditTracesParams().
						WithBody(&models.V1AuditFindRequest{Limit: 100})), nil).
						Return(&audit.FindAuditTracesOK{
							Payload: []*models.V1AuditResponse{
								auditTrace2,
								auditTrace1,
							},
						}, nil)
				},
			},
			want: []*models.V1AuditResponse{
				auditTrace1,
				auditTrace2,
			},
			wantTable: pointer.Pointer(`
TIME                  REQUEST ID                             PATH        CODE   TENANT     USER     PHASE
May 19 01:02:03.000   c40ad996-e1fd-4511-a7bf-418219cb8d91   /v1/audit   200    a-tenant   a-user   response
May 19 01:03:03.000   b5817ef7-980a-41ef-9ed3-741a143870b0   /v1/audit   403    b-tenant   b-user   request
`),
			wantWideTable: pointer.Pointer(`
TIME                  REQUEST ID                             PATH        CODE   TENANT     USER     PHASE      BODY
May 19 01:02:03.000   c40ad996-e1fd-4511-a7bf-418219cb8d91   /v1/audit   200    a-tenant   a-user   response   {"a": "b"}
May 19 01:03:03.000   b5817ef7-980a-41ef-9ed3-741a143870b0   /v1/audit   403    b-tenant   b-user   request    {"c": "d"}
`),
			template: pointer.Pointer(`{{ date "02/01/2006" .timestamp }} {{ .rqid }}`),
			wantTemplate: pointer.Pointer(`
19/05/2022 c40ad996-e1fd-4511-a7bf-418219cb8d91
19/05/2022 b5817ef7-980a-41ef-9ed3-741a143870b0
`),
			wantMarkdown: pointer.Pointer(`
|        TIME         |              REQUEST ID              |   PATH    | CODE |  TENANT  |  USER  |  PHASE   |
|---------------------|--------------------------------------|-----------|------|----------|--------|----------|
| May 19 01:02:03.000 | c40ad996-e1fd-4511-a7bf-418219cb8d91 | /v1/audit |  200 | a-tenant | a-user | response |
| May 19 01:03:03.000 | b5817ef7-980a-41ef-9ed3-741a143870b0 | /v1/audit |  403 | b-tenant | b-user | request  |
`),
		},
		{
			name: "list with filters",
			cmd: func(want []*models.V1AuditResponse) []string {
				args := []string{"audit", "list",
					"--path", want[0].Path,
					"--query", want[0].Body,
					"--phase", want[0].Phase,
					"--request-id", want[0].Rqid,
					"--tenant", want[0].Tenant,
					"--user", want[0].User,
				}
				assertExhaustiveArgs(t, args, "sort-by")
				return args
			},
			mocks: &client.MetalMockFns{
				Audit: func(mock *mock.Mock) {
					mock.On("FindAuditTraces", testcommon.MatchIgnoreContext(t, audit.NewFindAuditTracesParams().WithBody(&models.V1AuditFindRequest{
						Body: auditTrace1.Body,
						// From:   strfmt.DateTime{},
						Path:   auditTrace1.Path,
						Phase:  auditTrace1.Phase,
						Rqid:   auditTrace1.Rqid,
						Tenant: auditTrace1.Tenant,
						// To:     strfmt.DateTime{},
						User: auditTrace1.User,
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
TIME                  REQUEST ID                             PATH        CODE   TENANT     USER     PHASE
May 19 01:02:03.000   c40ad996-e1fd-4511-a7bf-418219cb8d91   /v1/audit   200    a-tenant   a-user   response
		`),
			wantWideTable: pointer.Pointer(`
TIME                  REQUEST ID                             PATH        CODE   TENANT     USER     PHASE      BODY
May 19 01:02:03.000   c40ad996-e1fd-4511-a7bf-418219cb8d91   /v1/audit   200    a-tenant   a-user   response   {"a": "b"}
		`),
			template: pointer.Pointer(`{{ date "02/01/2006" .timestamp }} {{ .rqid }}`),
			wantTemplate: pointer.Pointer(`
19/05/2022 c40ad996-e1fd-4511-a7bf-418219cb8d91
`),
			wantMarkdown: pointer.Pointer(`
|        TIME         |              REQUEST ID              |   PATH    | CODE |  TENANT  |  USER  |  PHASE   |
|---------------------|--------------------------------------|-----------|------|----------|--------|----------|
| May 19 01:02:03.000 | c40ad996-e1fd-4511-a7bf-418219cb8d91 | /v1/audit |  200 | a-tenant | a-user | response |
		`),
		},
	}
	for _, tt := range tests {
		tt.testCmd(t)
	}
}
