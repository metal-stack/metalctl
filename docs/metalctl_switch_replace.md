## metalctl switch replace

put a leaf switch into replace mode in preparation for physical replacement. For a description of the steps involved see the long help.

### Synopsis

Put a leaf switch into replace mode in preparation for physical replacement

Operational steps to replace a switch:

- Put the switch that needs to be replaced in replace mode with this command
- Replace the switch MAC address in the metal-stack deployment configuration
- Make sure that interfaces on the new switch do not get connected to the PXE-bridge immediately by setting the interfaces list of the respective leaf switch to [] in the metal-stack deployment configuration
- Deploy the management servers so that the dhcp servers will serve the right address and DHCP options to the new switch
- Replace the switch physically. Be careful to ensure that the cabling mirrors the remaining leaf exactly because the new switch information will be cloned from the remaining switch! Also make sure to have console access to the switch so you can start and monitor the install process
- If the switch is not in onie install mode but already has an operating system installed, put it into install mode with "sudo onie-select -i -f -v" and reboot it. Now the switch should be provisioned with a management IP from a management server, install itself with the right software image and receive license and ssh keys through ZTP. You can check whether that process has completed successfully with the command "sudo ztp -s". The ZTP state should be disabled and the result should be success.
- Deploy the switch plane and metal-core through metal-stack deployment CI job
- The switch will now register with its metal-api, and the metal-core service will receive the cloned interface and routing information. You can verify successful switch replacement by checking the interface and BGP configuration, and checking the switch status with "metalctl switch ls -o wide"; it should now be operational again

```
metalctl switch replace <switchID> [flags]
```

### Options

```
  -h, --help   help for replace
```

### Options inherited from parent commands

```
      --api-token string       api token to authenticate. Can be specified with METALCTL_API_TOKEN environment variable.
      --api-url string         api server address. Can be specified with METALCTL_API_URL environment variable.
      --api-v2-token string    api v2 token to authenticate. Can be specified with METALCTL_API_V2_TOKEN environment variable.
      --api-v2-url string      api server v2 address. Can be specified with METALCTL_API_V2_URL environment variable.
  -c, --config string          alternative config file path, (default is ~/.metalctl/config.yaml).
                               Example config.yaml:
                               
                               ---
                               apitoken: "alongtoken"
                               ...
                               
                               
      --debug                  debug output
      --force-color            force colored output even without tty
      --kubeconfig string      Path to the kube-config to use for authentication and authorization. Is updated by login. Uses default path if not specified.
      --no-headers             do not print headers of table output format (default print headers)
  -o, --output-format string   output format (table|wide|markdown|json|yaml|template), wide is a table with more columns. (default "table")
      --template string        output template for template output-format, go template format.
                               For property names inspect the output of -o json or -o yaml for reference.
                               Example for machines:
                               
                               metalctl machine list -o template --template "{{ .id }}:{{ .size.id  }}"
                               
                               
      --yes-i-really-mean-it   skips security prompts (which can be dangerous to set blindly because actions can lead to data loss or additional costs)
```

### SEE ALSO

* [metalctl switch](metalctl_switch.md)	 - manage switch entities

