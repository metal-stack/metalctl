package tableprinters

import (
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/metal-stack/metal-go/api/models"
	"github.com/metal-stack/metal-lib/pkg/pointer"
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
				{Nic: &models.V1SwitchNic{Name: pointer.Pointer("swp10")}},
				{Nic: &models.V1SwitchNic{Name: pointer.Pointer("swp1s4")}},
				{Nic: &models.V1SwitchNic{Name: pointer.Pointer("swp1s3")}},
				{Nic: &models.V1SwitchNic{Name: pointer.Pointer("swp1s1")}},
				{Nic: &models.V1SwitchNic{Name: pointer.Pointer("swp1s2")}},
				{Nic: &models.V1SwitchNic{Name: pointer.Pointer("swp9")}},
			},
			want: []*models.V1SwitchConnection{
				{Nic: &models.V1SwitchNic{Name: pointer.Pointer("swp1s1")}},
				{Nic: &models.V1SwitchNic{Name: pointer.Pointer("swp1s2")}},
				{Nic: &models.V1SwitchNic{Name: pointer.Pointer("swp1s3")}},
				{Nic: &models.V1SwitchNic{Name: pointer.Pointer("swp1s4")}},
				{Nic: &models.V1SwitchNic{Name: pointer.Pointer("swp9")}},
				{Nic: &models.V1SwitchNic{Name: pointer.Pointer("swp10")}},
			},
		},
		{
			name: "sorts interface names for sonic-like interface names",
			conns: []*models.V1SwitchConnection{
				{Nic: &models.V1SwitchNic{Name: pointer.Pointer("Ethernet3")}},
				{Nic: &models.V1SwitchNic{Name: pointer.Pointer("Ethernet49")}},
				{Nic: &models.V1SwitchNic{Name: pointer.Pointer("Ethernet10")}},
				{Nic: &models.V1SwitchNic{Name: pointer.Pointer("Ethernet2")}},
				{Nic: &models.V1SwitchNic{Name: pointer.Pointer("Ethernet1")}},
				{Nic: &models.V1SwitchNic{Name: pointer.Pointer("Ethernet11")}},
			},
			want: []*models.V1SwitchConnection{
				{Nic: &models.V1SwitchNic{Name: pointer.Pointer("Ethernet1")}},
				{Nic: &models.V1SwitchNic{Name: pointer.Pointer("Ethernet2")}},
				{Nic: &models.V1SwitchNic{Name: pointer.Pointer("Ethernet3")}},
				{Nic: &models.V1SwitchNic{Name: pointer.Pointer("Ethernet10")}},
				{Nic: &models.V1SwitchNic{Name: pointer.Pointer("Ethernet11")}},
				{Nic: &models.V1SwitchNic{Name: pointer.Pointer("Ethernet49")}},
			},
		},
		{
			name: "sorts interface names edge cases",
			conns: []*models.V1SwitchConnection{
				{Nic: &models.V1SwitchNic{Name: pointer.Pointer("123")}},
				{Nic: &models.V1SwitchNic{Name: pointer.Pointer("")}},
				{Nic: &models.V1SwitchNic{Name: pointer.Pointer("Ethernet1")}},
				{Nic: &models.V1SwitchNic{Name: pointer.Pointer("swp1s4w5")}},
				{Nic: &models.V1SwitchNic{Name: pointer.Pointer("foo")}},
				{Nic: &models.V1SwitchNic{Name: pointer.Pointer("swp1s3w3")}},
				{Nic: &models.V1SwitchNic{Name: pointer.Pointer("Ethernet100")}},
				{Nic: &models.V1SwitchNic{Name: pointer.Pointer("swp1s4w6")}},
				{Nic: &models.V1SwitchNic{Name: nil}},
			},
			want: []*models.V1SwitchConnection{
				{Nic: &models.V1SwitchNic{Name: pointer.Pointer("swp1s3w3")}},
				{Nic: &models.V1SwitchNic{Name: pointer.Pointer("swp1s4w5")}},
				{Nic: &models.V1SwitchNic{Name: pointer.Pointer("swp1s4w6")}},
				{Nic: &models.V1SwitchNic{Name: pointer.Pointer("Ethernet1")}},
				{Nic: &models.V1SwitchNic{Name: pointer.Pointer("Ethernet100")}},
				{Nic: &models.V1SwitchNic{Name: pointer.Pointer("")}},
				{Nic: &models.V1SwitchNic{Name: nil}},
				{Nic: &models.V1SwitchNic{Name: pointer.Pointer("123")}},
				{Nic: &models.V1SwitchNic{Name: pointer.Pointer("foo")}},
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
