package cmd

import (
	"testing"

	"github.com/metal-stack/metal-go/api/models"
	"github.com/metal-stack/metal-lib/pkg/pointer"

	"github.com/stretchr/testify/assert"
)

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
