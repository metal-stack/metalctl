package completion

import (
	"github.com/metal-stack/metal-go/api/client/switch_operations"
	"github.com/metal-stack/metal-go/api/models"
	"github.com/spf13/cobra"
)

func (c *Completion) SwitchListCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	resp, err := c.client.SwitchOperations().ListSwitches(switch_operations.NewListSwitchesParams(), nil)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	var names []string
	for _, s := range resp.Payload {
		if s.ID == nil {
			continue
		}
		names = append(names, *s.ID)
	}
	return names, cobra.ShellCompDirectiveNoFileComp
}

func (c *Completion) SwitchNameListCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	resp, err := c.client.SwitchOperations().ListSwitches(switch_operations.NewListSwitchesParams(), nil)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	var names []string
	for _, p := range resp.Payload {
		names = append(names, p.Name)
	}
	return names, cobra.ShellCompDirectiveNoFileComp
}

func (c *Completion) SwitchRackListCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	resp, err := c.client.SwitchOperations().ListSwitches(switch_operations.NewListSwitchesParams(), nil)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	var names []string
	for _, p := range resp.Payload {
		if p.RackID == nil {
			continue
		}
		names = append(names, *p.RackID)
	}
	return names, cobra.ShellCompDirectiveNoFileComp
}

func (c *Completion) SwitchOSVendorListCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	resp, err := c.client.SwitchOperations().ListSwitches(switch_operations.NewListSwitchesParams(), nil)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	var names []string
	for _, p := range resp.Payload {
		if p.Os == nil {
			continue
		}
		names = append(names, p.Os.Vendor)
	}
	return names, cobra.ShellCompDirectiveNoFileComp
}

func (c *Completion) SwitchOSVersionListCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	resp, err := c.client.SwitchOperations().ListSwitches(switch_operations.NewListSwitchesParams(), nil)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	var names []string
	for _, p := range resp.Payload {
		if p.Os == nil {
			continue
		}
		names = append(names, p.Os.Version)
	}
	return names, cobra.ShellCompDirectiveNoFileComp
}

func (c *Completion) SwitchListPorts(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) == 0 {
		// there is no switch selected so we cannot get the list of ports
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	resp, err := c.client.SwitchOperations().FindSwitch(switch_operations.NewFindSwitchParams().WithID(args[0]), nil)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	var names []string
	for _, n := range resp.Payload.Nics {
		if n != nil {
			names = append(names, *n.Name)
		}
	}
	return names, cobra.ShellCompDirectiveNoFileComp
}

func (c *Completion) xSwitchListCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	var names []string
	if len(args) == 0 {
		resp, err := c.client.SwitchOperations().ListSwitches(switch_operations.NewListSwitchesParams(), nil)
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}

		for _, s := range resp.Payload {
			if s.ID == nil {
				continue
			}
			names = append(names, *s.ID)
		}
		return names, cobra.ShellCompDirectiveNoFileComp
	}
	if len(args) < 3 {
		resp, err := c.client.SwitchOperations().FindSwitch(switch_operations.NewFindSwitchParams().WithID(args[0]), nil)
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}
		if len(args) == 1 {
			for _, n := range resp.Payload.Nics {
				if n != nil {
					names = append(names, *n.Name)
				}
			}
			return names, cobra.ShellCompDirectiveNoFileComp
		}

		for _, n := range resp.Payload.Nics {
			if n != nil && *n.Name == args[1] {
				if n.Actual == nil {
					names = append(names, models.V1SwitchNicActualUP, models.V1SwitchNicActualDOWN)
				} else {
					if *n.Actual == models.V1SwitchNicActualUP {
						names = append(names, models.V1SwitchNicActualDOWN)
					}
					if *n.Actual == models.V1SwitchNicActualDOWN {
						names = append(names, models.V1SwitchNicActualUP)
					}
				}
				break
			}
		}
	}
	return names, cobra.ShellCompDirectiveNoFileComp
}
