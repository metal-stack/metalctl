package tableprinters

import (
	"fmt"

	"github.com/metal-stack/metal-go/api/models"
)

func ToHeaderAndRows(data interface{}, wide bool) ([]string, [][]string, error) {
	switch d := data.(type) {
	case *models.V1SizeImageConstraintResponse:
		return SizeImageConstraintTable([]*models.V1SizeImageConstraintResponse{d}, wide)
	case []*models.V1SizeImageConstraintResponse:
		return SizeImageConstraintTable(d, wide)
	case *models.V1SizeResponse:
		return SizeTable([]*models.V1SizeResponse{d}, wide)
	case []*models.V1SizeResponse:
		return SizeTable(d, wide)
	default:
		return nil, nil, fmt.Errorf("unknown table printer for type: %T", d)
	}
}
