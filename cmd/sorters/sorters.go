package sorters

import (
	"fmt"
	"strings"

	"github.com/metal-stack/metal-lib/pkg/multisort"
	"github.com/spf13/viper"
)

func MustKeysFromCLIOrDefaults(defaultKeys multisort.Keys) multisort.Keys {
	if !viper.IsSet("order") {
		return defaultKeys
	}

	var keys multisort.Keys

	for _, col := range viper.GetStringSlice("order") {
		col = strings.ToLower(strings.TrimSpace(col))

		var descending bool

		id, directionRaw, found := strings.Cut(col, ":")
		if found {
			switch directionRaw {
			case "asc", "ascending":
				descending = false
			case "desc", "descending":
				descending = true
			default:
				panic(fmt.Errorf("unsupported sort direction: %s", directionRaw))
			}
		}

		keys = append(keys, multisort.Key{ID: id, Descending: descending})
	}

	return keys
}
