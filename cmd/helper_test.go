package cmd

import (
	"testing"

	"github.com/metal-stack/metal-go/api/models"
	"github.com/metal-stack/metal-lib/pkg/pointer"

	"github.com/stretchr/testify/assert"
)

/*func Test_shortID(t *testing.T) {
	tests := []struct {
		name      string
		machineID string
		want      string
	}{
		{
			name:      "simple",
			machineID: "00000000-0000-0000-0000-ac1f6bb979ff",
			want:      "ac1f6bb979ff",
		},
		{
			name:      "trailing zero",
			machineID: "00000000-0000-0000-0000-ac1f6bb979f0",
			want:      "ac1f6bb979f0",
		},
		{
			name:      "more trailing zero",
			machineID: "00000000-0000-0000-0000-ac1f6bb97000",
			want:      "ac1f6bb97000",
		},
		{
			name:      "middle zero",
			machineID: "00000000-0000-0000-0000-ac1f600979f0",
			want:      "ac1f600979f0",
		},
		{
			name:      "leading zero",
			machineID: "00000000-0000-0000-0000-0c1f600979f0",
			want:      "0c1f600979f0",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := shortID(tt.machineID); got != tt.want {
				t.Errorf("shortID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_longID(t *testing.T) {
	tests := []struct {
		name    string
		shortID string
		want    string
	}{
		{
			name:    "simple",
			shortID: "ac1f600979f0",
			want:    "00000000-0000-0000-0000-ac1f600979f0",
		},
		{
			name:    "leading zero",
			shortID: "000f600979f0",
			want:    "00000000-0000-0000-0000-000f600979f0",
		},
		{
			name:    "half",
			shortID: "abc-000f600979f0",
			want:    "00000000-0000-0000-0abc-000f600979f0",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := longID(tt.shortID); got != tt.want {
				t.Errorf("longID() = %v, want %v", got, tt.want)
			}
		})
	}
}*/

func Test_parseNetworks(t *testing.T) {
	assert := assert.New(t)

	tests := []struct {
		possibleNetworks []string
		isError          bool
		expected         []*models.V1MachineAllocationNetwork
		name             string
	}{
		{
			name:             "empty networks",
			possibleNetworks: []string{},
			isError:          false,
			expected:         []*models.V1MachineAllocationNetwork{},
		},
		{
			name:             "single network",
			possibleNetworks: []string{"network"},
			isError:          false,
			expected: []*models.V1MachineAllocationNetwork{
				{Networkid: pointer.Pointer("network"), Autoacquire: pointer.Pointer(true)},
			},
		},
		{
			name:             "multiple networks",
			possibleNetworks: []string{"network1", "network2"},
			isError:          false,
			expected: []*models.V1MachineAllocationNetwork{
				{Networkid: pointer.Pointer("network1"), Autoacquire: pointer.Pointer(true)},
				{Networkid: pointer.Pointer("network2"), Autoacquire: pointer.Pointer(true)},
			},
		},
		{
			name:             "single network with auto flag",
			possibleNetworks: []string{"network:auto"},
			isError:          false,
			expected: []*models.V1MachineAllocationNetwork{
				{Networkid: pointer.Pointer("network"), Autoacquire: pointer.Pointer(true)},
			},
		},
		{
			name:             "multiple networks with auto, noauto and empty flag",
			possibleNetworks: []string{"network1:auto", "network2:noauto", "network3"},
			isError:          false,
			expected: []*models.V1MachineAllocationNetwork{
				{Networkid: pointer.Pointer("network1"), Autoacquire: pointer.Pointer(true)},
				{Networkid: pointer.Pointer("network2"), Autoacquire: pointer.Pointer(false)},
				{Networkid: pointer.Pointer("network3"), Autoacquire: pointer.Pointer(true)},
			},
		},
		{
			name:             "single network with invalid flag",
			possibleNetworks: []string{"network:gopher"},
			isError:          true,
			expected:         nil,
		},
		{
			name:             "single network with invalid flag separator",
			possibleNetworks: []string{"network::"},
			isError:          true,
			expected:         nil,
		},
		{
			name:             "single network with noauto",
			possibleNetworks: []string{"network:noauto"},
			isError:          false,
			expected: []*models.V1MachineAllocationNetwork{
				{Networkid: pointer.Pointer("network"), Autoacquire: pointer.Pointer(false)},
			},
		},
	}

	for _, test := range tests {
		actual, err := parseNetworks(test.possibleNetworks)
		if test.isError {
			assert.Error(err, "Name: %s", test.name)
			assert.Nil(actual, "Name: %s", test.name)
		} else {
			assert.NoError(err, "Name: %s", test.name)
			assert.NotNil(actual, "Name: %s", test.name)
			assert.ElementsMatch(actual, test.expected, "Name: %s", test.name)
		}
	}
}
