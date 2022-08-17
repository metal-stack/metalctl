package api

import (
	metalmodels "github.com/metal-stack/metal-go/api/models"
)

type Version struct {
	Client string                   `json:"client" yaml:"client"`
	Server *metalmodels.RestVersion `json:"server" yaml:"server,omitempty"`
}
