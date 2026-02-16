package tableprinters

import (
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/metal-stack/metal-go/api/models"
	"github.com/metal-stack/metal-lib/pkg/testcommon"
)

func Test_switchInterfaceNameLessFunc(t *testing.T) {
	tests := []struct {
		name  string
		conns []*models.V1SwitchConnection
		want  []*models.V1SwitchConnection
	}{
		{
			name: "sorts interface names for cumulus-like interface names",
			conns: []*models.V1SwitchConnection{
				{Nic: &models.V1SwitchNic{Name: new("swp10")}},
				{Nic: &models.V1SwitchNic{Name: new("swp1s4")}},
				{Nic: &models.V1SwitchNic{Name: new("swp1s3")}},
				{Nic: &models.V1SwitchNic{Name: new("swp1s1")}},
				{Nic: &models.V1SwitchNic{Name: new("swp1s2")}},
				{Nic: &models.V1SwitchNic{Name: new("swp9")}},
			},
			want: []*models.V1SwitchConnection{
				{Nic: &models.V1SwitchNic{Name: new("swp1s1")}},
				{Nic: &models.V1SwitchNic{Name: new("swp1s2")}},
				{Nic: &models.V1SwitchNic{Name: new("swp1s3")}},
				{Nic: &models.V1SwitchNic{Name: new("swp1s4")}},
				{Nic: &models.V1SwitchNic{Name: new("swp9")}},
				{Nic: &models.V1SwitchNic{Name: new("swp10")}},
			},
		},
		{
			name: "sorts interface names for sonic-like interface names",
			conns: []*models.V1SwitchConnection{
				{Nic: &models.V1SwitchNic{Name: new("Ethernet3")}},
				{Nic: &models.V1SwitchNic{Name: new("Ethernet49")}},
				{Nic: &models.V1SwitchNic{Name: new("Ethernet10")}},
				{Nic: &models.V1SwitchNic{Name: new("Ethernet2")}},
				{Nic: &models.V1SwitchNic{Name: new("Ethernet1")}},
				{Nic: &models.V1SwitchNic{Name: new("Ethernet11")}},
			},
			want: []*models.V1SwitchConnection{
				{Nic: &models.V1SwitchNic{Name: new("Ethernet1")}},
				{Nic: &models.V1SwitchNic{Name: new("Ethernet2")}},
				{Nic: &models.V1SwitchNic{Name: new("Ethernet3")}},
				{Nic: &models.V1SwitchNic{Name: new("Ethernet10")}},
				{Nic: &models.V1SwitchNic{Name: new("Ethernet11")}},
				{Nic: &models.V1SwitchNic{Name: new("Ethernet49")}},
			},
		},
		{
			name: "sorts interface names edge cases",
			conns: []*models.V1SwitchConnection{
				{Nic: &models.V1SwitchNic{Name: new("123")}},
				{Nic: &models.V1SwitchNic{Name: new("")}},
				{Nic: &models.V1SwitchNic{Name: new("Ethernet1")}},
				{Nic: &models.V1SwitchNic{Name: new("swp1s4w5")}},
				{Nic: &models.V1SwitchNic{Name: new("foo")}},
				{Nic: &models.V1SwitchNic{Name: new("swp1s3w3")}},
				{Nic: &models.V1SwitchNic{Name: new("Ethernet100")}},
				{Nic: &models.V1SwitchNic{Name: new("swp1s4w6")}},
				{Nic: &models.V1SwitchNic{Name: nil}},
			},
			want: []*models.V1SwitchConnection{
				{Nic: &models.V1SwitchNic{Name: new("swp1s3w3")}},
				{Nic: &models.V1SwitchNic{Name: new("swp1s4w5")}},
				{Nic: &models.V1SwitchNic{Name: new("swp1s4w6")}},
				{Nic: &models.V1SwitchNic{Name: new("Ethernet1")}},
				{Nic: &models.V1SwitchNic{Name: new("Ethernet100")}},
				{Nic: &models.V1SwitchNic{Name: new("")}},
				{Nic: &models.V1SwitchNic{Name: nil}},
				{Nic: &models.V1SwitchNic{Name: new("123")}},
				{Nic: &models.V1SwitchNic{Name: new("foo")}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sort.Slice(tt.conns, switchInterfaceNameLessFunc(tt.conns))

			if diff := cmp.Diff(tt.conns, tt.want, testcommon.StrFmtDateComparer()); diff != "" {
				t.Errorf("diff (+got -want):\n %s", diff)
			}
		})
	}
}
