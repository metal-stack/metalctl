package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/metal-stack/metal-lib/auth"
	"io"
	"math"
	"net"
	"os"
	"sort"
	"strings"
	"time"

	metalgo "github.com/metal-stack/metal-go"
	"gopkg.in/yaml.v3"

	"github.com/metal-stack/metal-go/api/models"

	"github.com/spf13/viper"
)

func atLeastOneViperStringFlagGiven(flags ...string) bool {
	for _, flag := range flags {
		if viper.GetString(flag) != "" {
			return true
		}
	}
	return false
}

func atLeastOneViperStringSliceFlagGiven(flags ...string) bool {
	for _, flag := range flags {
		if len(viper.GetStringSlice(flag)) > 0 {
			return true
		}
	}
	return false
}

func atLeastOneViperBoolFlagGiven(flags ...string) bool {
	for _, flag := range flags {
		if viper.GetBool(flag) {
			return true
		}
	}
	return false
}

func atLeastOneViperInt64FlagGiven(flags ...string) bool {
	for _, flag := range flags {
		if viper.GetInt64(flag) != 0 {
			return true
		}
	}
	return false
}

func viperString(flag string) *string {
	if viper.GetString(flag) == "" {
		return nil
	}
	value := viper.GetString(flag)
	return &value
}

func viperStringSlice(flag string) []string {
	value := viper.GetStringSlice(flag)
	if len(value) == 0 {
		return nil
	}
	return value
}

func viperBool(flag string) *bool {
	if !viper.GetBool(flag) {
		return nil
	}
	value := viper.GetBool(flag)
	return &value
}

func viperInt64(flag string) *int64 {
	if viper.GetInt64(flag) == 0 {
		return nil
	}
	value := viper.GetInt64(flag)
	return &value
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

// FIXME write a test
func truncate(input, elipsis string, maxlength int) string {
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

func parseNetworks(values []string) ([]metalgo.MachineAllocationNetwork, error) {
	nets := []metalgo.MachineAllocationNetwork{}
	for _, netWithFlag := range values {
		id, autoAcquire, err := splitNetwork(netWithFlag)
		if err != nil {
			return nil, err
		}

		net := metalgo.MachineAllocationNetwork{
			Autoacquire: autoAcquire,
			NetworkID:   id,
		}
		nets = append(nets, net)
	}
	return nets, nil
}

func splitNetwork(value string) (string, bool, error) {
	splitNets := strings.SplitN(value, ":", 2)
	id := splitNets[0]
	if len(splitNets) > 1 {
		mode := strings.ToLower(splitNets[1])
		switch mode {
		// case NETWORK:auto
		case "auto":
			return id, true, nil
		// case NETWORK:noauto
		case "noauto":
			return id, false, nil
		// case NETWORK:<illegal value>
		default:
			return "", false, fmt.Errorf("illegal mode: %s", mode)
		}
	}
	// case: NETWORK, defaults to NETWORK:auto
	return id, true, nil
}

// func shortID(machineID string) string {
// 	result := strings.ReplaceAll(machineID, "00000000-", "")
// 	result = strings.ReplaceAll(result, "0000-", "")
// 	return result
// }

// func longID(shortID string) string {
// 	machineIDPattern := []byte("00000000-0000-0000-0000-000000000000")
// 	longIDLength := len(machineIDPattern)
// 	result := machineIDPattern
// 	shortIDSlice := []byte(strings.TrimSpace(shortID))
// 	for i := len(shortID) - 1; i >= 0; i-- {
// 		pos := longIDLength - i - 1
// 		shortPos := len(shortIDSlice) - i - 1
// 		result[pos] = shortIDSlice[shortPos]
// 	}
// 	return string(result)
// }

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

// strValue returns the value of a string pointer of not nil, otherwise empty string
func strValue(strPtr *string) string {
	if strPtr != nil {
		return *strPtr
	}
	return ""
}

// genericObject transforms the input to a struct which has fields with the same name as in the json struct.
// this is handy for template rendering as the output of -o json|yaml can be used as the input for the template
func genericObject(input interface{}) map[string]interface{} {
	b, err := json.Marshal(input)
	if err != nil {
		fmt.Printf("unable to marshall input:%v", err)
		os.Exit(1)
	}
	var result interface{}
	err = json.Unmarshal(b, &result)
	if err != nil {
		fmt.Printf("unable to unmarshal input:%v", err)
		os.Exit(1)
	}
	return result.(map[string]interface{})
}

func labelsFromTags(tags []string) map[string]string {
	labels := make(map[string]string)
	for _, tag := range tags {
		parts := strings.Split(tag, "=")
		partslen := len(parts)
		switch partslen {
		case 1:
			labels[tag] = "true"
		case 2:
			labels[parts[0]] = parts[1]
		default:
			values := strings.Join(parts[1:], "")
			labels[parts[0]] = values
		}
	}
	return labels
}

// readFrom will either read from stdin (-) or a file path an marshall from yaml to data
func readFrom(from string, data interface{}, f func(target interface{})) error {
	var reader io.Reader
	var err error
	switch from {
	case "-":
		reader = os.Stdin
	default:
		reader, err = os.Open(from)
		if err != nil {
			return fmt.Errorf("unable to open %s %v", from, err)
		}
	}
	dec := yaml.NewDecoder(reader)
	for {
		err := dec.Decode(data)
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("decode error: %v", err)
		}
		f(data)
	}
	return nil
}

const cloudContext = "cloudctl"

// getAuthContext reads AuthContext from given kubeconfig
func getAuthContext(kubeconfig string) (*auth.AuthContext, error) {
	cs, err := getContexts()
	if err != nil {
		return nil, err
	}
	authContext, err := auth.GetAuthContext(kubeconfig, formatContextName(cloudContext, cs.CurrentContext))
	if err != nil {
		return nil, err
	}

	if !authContext.AuthProviderOidc {
		return nil, fmt.Errorf("active user %s has no oidc authProvider, check config", authContext.User)
	}

	return &authContext, nil
}

// formatContextName returns the contextName for the given suffix. suffix can be empty.
func formatContextName(prefix string, suffix string) string {
	contextName := prefix
	if suffix != "" {
		contextName = fmt.Sprintf("%s-%s", cloudContext, suffix)
	}
	return contextName
}
