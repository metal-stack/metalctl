package tableprinters

import (
	"bytes"
	"fmt"
	"math"
	"net"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/metal-stack/metal-go/api/models"
)

const (
	bark            = "🚧"
	circle          = "↻"
	dot             = "●"
	exclamationMark = "❗"
	lock            = "🔒"
	nbr             = " "
	question        = "❓"
	skull           = "💀"
)

func depth(path string) uint {
	var count uint = 0
	for p := filepath.Clean(path); p != "/"; count++ {
		p = filepath.Dir(p)
	}
	return count
}

//nolint:unparam
func truncate(input string, maxlength int) string {
	elipsis := "..."
	il := len(input)
	el := len(elipsis)
	if il <= maxlength {
		return input
	}
	if maxlength <= el {
		return input[:maxlength]
	}
	startlength := ((maxlength - el) / 2) - el/2

	output := input[:startlength] + elipsis
	missing := maxlength - len(output)
	output = output + input[il-missing:]
	return output
}

func truncateEnd(input string, maxlength int) string {
	elipsis := "..."
	length := len(input) + len(elipsis)
	if length <= maxlength {
		return input
	}
	return input[:maxlength] + elipsis
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

func sortIPs(v1ips []*models.V1IPResponse) []*models.V1IPResponse {

	v1ipmap := make(map[string]*models.V1IPResponse)
	var ips []string
	for _, v1ip := range v1ips {
		v1ipmap[*v1ip.Ipaddress] = v1ip
		ips = append(ips, *v1ip.Ipaddress)
	}

	realIPs := make([]net.IP, 0, len(ips))

	for _, ip := range ips {
		realIPs = append(realIPs, net.ParseIP(ip))
	}

	sort.Slice(realIPs, func(i, j int) bool {
		return bytes.Compare(realIPs[i], realIPs[j]) < 0
	})

	var result []*models.V1IPResponse
	for _, ip := range realIPs {
		result = append(result, v1ipmap[ip.String()])
	}
	return result
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
