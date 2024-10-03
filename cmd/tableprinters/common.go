package tableprinters

import (
	"fmt"
	"math"
	"path/filepath"
	"strings"
	"time"
)

const (
	dot        = "●"
	nbr        = " "
	poweron    = "⏻"
	powersleep = "⏾"
)

func depth(path string) uint {
	var count uint = 0
	for p := filepath.Clean(path); p != "/"; count++ {
		p = filepath.Dir(p)
	}
	return count
}

func humanizeDuration(duration time.Duration) string {
	days := int64(duration.Hours() / 24)
	hours := int64(math.Mod(duration.Hours(), 24))
	minutes := int64(math.Mod(duration.Minutes(), 60))
	seconds := int64(math.Mod(duration.Seconds(), 60))

	chunks := []struct {
		singularName string
		amount       int64
	}{
		{"d", days},
		{"h", hours},
		{"m", minutes},
		{"s", seconds},
	}

	parts := []string{}

	for _, chunk := range chunks {
		switch chunk.amount {
		case 0:
			continue
		default:
			parts = append(parts, fmt.Sprintf("%d%s", chunk.amount, chunk.singularName))
		}
	}

	if len(parts) == 0 {
		return "0s"
	}
	if len(parts) > 2 {
		parts = parts[:2]
	}
	return strings.Join(parts, " ")
}

func getMaxLineCount(ss ...string) int {
	max := 0
	for _, s := range ss {
		c := strings.Count(s, "\n")
		if c > max {
			max = c
		}
	}
	return max
}
