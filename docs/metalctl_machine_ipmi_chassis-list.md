## metalctl machine ipmi chassis-list

display ipmi machines grouped by chassis serial

### Synopsis

display ipmi machines grouped by chassis serial

Meaning of the emojis:

🚧 Machine is reserved. Reserved machines are not considered for random allocation until the reservation flag is removed.
🔒 Machine is locked. Locked machines can not be deleted until the lock is removed.
💀 Machine is dead. The metal-api does not receive any events from this machine.
❗ Machine has a last event error. The machine has recently encountered an error during the provisioning lifecycle.
❓ Machine is in unknown condition. The metal-api does not receive phoned home events anymore or has never booted successfully.
⭕ Machine is in a provisioning crash loop. Flag can be reset through an API-triggered reboot or when the machine reaches the phoned home state.
🚑 Machine reclaim has failed. The machine was deleted but it is not going back into the available machine pool.
🛡 Machine is connected to our VPN, ssh access only possible via this VPN.


```
metalctl machine ipmi chassis-list [flags]
```

### Options

```
      --bmc-address string                    bmc ipmi address (needs to include port) to filter [optional]
      --bmc-mac string                        bmc mac address to filter [optional]
      --board-part-number string              fru board part number to filter [optional]
  -h, --help                                  help for chassis-list
      --hostname string                       allocation hostname to filter [optional]
      --id string                             ID to filter [optional]
      --image string                          allocation image to filter [optional]
      --last-event-error-threshold duration   the duration up to how long in the past a machine last event error will be counted as an issue [optional] (default 1h0m0s)
      --mac string                            mac to filter [optional]
      --manufacturer string                   fru manufacturer to filter [optional]
      --name string                           allocation name to filter [optional]
      --network-destination-prefixes string   network destination prefixes to filter [optional]
      --network-ids string                    network ids to filter [optional]
      --network-ips string                    network ips to filter [optional]
      --partition string                      partition to filter [optional]
      --product-part-number string            fru product part number to filter [optional]
      --product-serial string                 fru product serial to filter [optional]
      --project string                        allocation project to filter [optional]
      --rack string                           rack to filter [optional]
      --role string                           allocation role to filter [optional]
      --size string                           size to filter [optional]
      --sort-by strings                       sort by (comma separated) column(s), sort direction can be changed by appending :asc or :desc behind the column identifier. possible values: age|bios|bmc|event|id|liveliness|partition|project|rack|size|when
      --state string                          state to filter [optional]
      --tags strings                          tags to filter, use it like: --tags "tag1,tag2" or --tags "tag3".
```

### Options inherited from parent commands

```
      --api-token string       api token to authenticate. Can be specified with METALCTL_API_TOKEN environment variable.
      --api-url string         api server address. Can be specified with METALCTL_API_URL environment variable.
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

* [metalctl machine ipmi](metalctl_machine_ipmi.md)	 - display ipmi details of the machine, if no machine ID is given all ipmi addresses are returned.

