package api

import (
	metalmodels "github.com/metal-stack/metal-go/api/models"
)

type Version struct {
	Client string                   `yaml:"client"`
	Server *metalmodels.RestVersion `yaml:"server,omitempty"`
}
