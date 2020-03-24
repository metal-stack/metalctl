package cmd

import (
	"testing"

	"github.com/metal-stack/metal-go/api/models"
	"github.com/stretchr/testify/require"
)

func TestOrdering(t *testing.T) {
	// given
	mmtp := MetalMachineTablePrinter{TablePrinter{}}

	data := createTestData()
	require.Equal(t, "2", *data[0].ID)
	require.Equal(t, "3", *data[1].ID)
	require.Equal(t, "1", *data[2].ID)
	require.Equal(t, "4", *data[3].ID)
	require.Equal(t, "3", *data[4].ID)

	require.Nil(t, data[0].Size)
	require.Equal(t, "4", *data[1].Size.ID)
	require.Equal(t, "1", *data[2].Size.ID)
	require.Equal(t, "2", *data[3].Size.ID)
	require.Equal(t, "3", *data[4].Size.ID)

	require.Equal(t, "2", *data[0].Events.Log[0].Event)
	require.Equal(t, "1", *data[1].Events.Log[0].Event)
	require.Equal(t, "1", *data[2].Events.Log[0].Event)
	require.Equal(t, "3", *data[3].Events.Log[0].Event)
	require.Equal(t, "2", *data[4].Events.Log[0].Event)

	// when
	mmtp.order = "ID"
	mmtp.Order(data)

	// then
	require.Equal(t, "1", *data[0].ID)
	require.Equal(t, "2", *data[1].ID)
	require.Equal(t, "3", *data[2].ID)
	require.Equal(t, "3", *data[3].ID)
	require.Equal(t, "4", *data[4].ID)

	require.Equal(t, "1", *data[0].Size.ID)
	require.Nil(t, data[1].Size)
	require.Equal(t, "4", *data[2].Size.ID)
	require.Equal(t, "3", *data[3].Size.ID)
	require.Equal(t, "2", *data[4].Size.ID)

	require.Equal(t, "1", *data[0].Events.Log[0].Event)
	require.Equal(t, "2", *data[1].Events.Log[0].Event)
	require.Equal(t, "1", *data[2].Events.Log[0].Event)
	require.Equal(t, "2", *data[3].Events.Log[0].Event)
	require.Equal(t, "3", *data[4].Events.Log[0].Event)

	// when
	data = createTestData()
	mmtp.order = "id,size"
	mmtp.Order(data)

	// then
	require.Equal(t, "1", *data[0].ID)
	require.Equal(t, "2", *data[1].ID)
	require.Equal(t, "3", *data[2].ID)
	require.Equal(t, "3", *data[3].ID)
	require.Equal(t, "4", *data[4].ID)

	require.Equal(t, "1", *data[0].Size.ID)
	require.Nil(t, data[1].Size)
	require.Equal(t, "3", *data[2].Size.ID)
	require.Equal(t, "4", *data[3].Size.ID)
	require.Equal(t, "2", *data[4].Size.ID)

	require.Equal(t, "1", *data[0].Events.Log[0].Event)
	require.Equal(t, "2", *data[1].Events.Log[0].Event)
	require.Equal(t, "2", *data[2].Events.Log[0].Event)
	require.Equal(t, "1", *data[3].Events.Log[0].Event)
	require.Equal(t, "3", *data[4].Events.Log[0].Event)

	// when
	data = createTestData()
	mmtp.order = "id,event"
	mmtp.Order(data)

	// then
	require.Equal(t, "1", *data[0].ID)
	require.Equal(t, "2", *data[1].ID)
	require.Equal(t, "3", *data[2].ID)
	require.Equal(t, "3", *data[3].ID)
	require.Equal(t, "4", *data[4].ID)

	require.Equal(t, "1", *data[0].Size.ID)
	require.Nil(t, data[1].Size)
	require.Equal(t, "4", *data[2].Size.ID)
	require.Equal(t, "3", *data[3].Size.ID)
	require.Equal(t, "2", *data[4].Size.ID)

	require.Equal(t, "1", *data[0].Events.Log[0].Event)
	require.Equal(t, "2", *data[1].Events.Log[0].Event)
	require.Equal(t, "1", *data[2].Events.Log[0].Event)
	require.Equal(t, "2", *data[3].Events.Log[0].Event)
	require.Equal(t, "3", *data[4].Events.Log[0].Event)

	// when
	data = createTestData()
	mmtp.order = "id,size,event"
	mmtp.Order(data)

	// then
	require.Equal(t, "1", *data[0].ID)
	require.Equal(t, "2", *data[1].ID)
	require.Equal(t, "3", *data[2].ID)
	require.Equal(t, "3", *data[3].ID)
	require.Equal(t, "4", *data[4].ID)

	require.Equal(t, "1", *data[0].Size.ID)
	require.Nil(t, data[1].Size)
	require.Equal(t, "3", *data[2].Size.ID)
	require.Equal(t, "4", *data[3].Size.ID)
	require.Equal(t, "2", *data[4].Size.ID)

	require.Equal(t, "1", *data[0].Events.Log[0].Event)
	require.Equal(t, "2", *data[1].Events.Log[0].Event)
	require.Equal(t, "2", *data[2].Events.Log[0].Event)
	require.Equal(t, "1", *data[3].Events.Log[0].Event)
	require.Equal(t, "3", *data[4].Events.Log[0].Event)

	// when
	data = createTestData()
	mmtp.order = "event"
	mmtp.Order(data)

	// then
	require.Equal(t, "3", *data[0].ID)
	require.Equal(t, "1", *data[1].ID)
	require.Equal(t, "2", *data[2].ID)
	require.Equal(t, "3", *data[3].ID)
	require.Equal(t, "4", *data[4].ID)

	require.Equal(t, "4", *data[0].Size.ID)
	require.Equal(t, "1", *data[1].Size.ID)
	require.Nil(t, data[2].Size)
	require.Equal(t, "3", *data[3].Size.ID)
	require.Equal(t, "2", *data[4].Size.ID)

	require.Equal(t, "1", *data[0].Events.Log[0].Event)
	require.Equal(t, "1", *data[1].Events.Log[0].Event)
	require.Equal(t, "2", *data[2].Events.Log[0].Event)
	require.Equal(t, "2", *data[3].Events.Log[0].Event)
	require.Equal(t, "3", *data[4].Events.Log[0].Event)
}

func createTestData() []*models.V1MachineResponse {
	ids := []string{"4", "1", "2", "3"}
	sizes := []*models.V1SizeResponse{
		nil,
	}
	for i := range ids {
		sizes = append(sizes, &models.V1SizeResponse{ID: &ids[i]})
	}

	ee := []string{"2", "1", "1", "3", "2"}
	var events []*models.V1MachineRecentProvisioningEvents
	for i := range ee {
		events = append(events, &models.V1MachineRecentProvisioningEvents{
			Log: []*models.V1MachineProvisioningEvent{{Event: &ee[i]}},
		})
	}

	ids = []string{"2", "3", "1", "4", "3"}
	data := make([]*models.V1MachineResponse, len(ids))
	for i := range data {
		data[i] = &models.V1MachineResponse{
			ID:     &ids[i],
			Size:   sizes[i],
			Events: events[i],
		}
	}
	return data
}
