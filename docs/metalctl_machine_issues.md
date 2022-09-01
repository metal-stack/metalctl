## metalctl machine issues

display machines which are in a potential bad state

### Synopsis

display machines which are in a potential bad state

Meaning of the emojis:

🚧 Machine is reserved. Reserved machines are not considered for random allocation until the reservation flag is removed.
🔒 Machine is locked. Locked machines can not be deleted until the lock is removed.
💀 Machine is dead. The metal-api does not receive any events from this machine.
❗ Machine has a last event error. The machine has recently encountered an error during the provisioning lifecycle.
❓ Machine is in unknown condition. The metal-api does not receive phoned home events anymore or has never booted successfully.
⭕ Machine is in a provisioning crash loop. Flag can be reset through an API-triggered reboot or when the machine reaches the phoned home state.
🚑 Machine reclaim has failed. The machine was deleted but it is not going back into the available machine pool.


```
metalctl machine issues [<machine ID>] [flags]
```

### Options

```
      --failed-event-threshold duration   issue severity to include [optional] (default 168h0m0s)
  -h, --help                              help for issues
      --hostname string                   allocation hostname to filter [optional]
      --id string                         ID to filter [optional]
      --image string                      allocation image to filter [optional]
      --mac string                        mac to filter [optional]
      --name string                       allocation name to filter [optional]
      --omit strings                      issue types to omit [optional]
      --only strings                      issue types to include [optional]
      --partition string                  partition to filter [optional]
      --project string                    allocation project to filter [optional]
      --severity string                   issue severity to include [optional]
      --size string                       size to filter [optional]
      --tags strings                      tags to filter, use it like: --tags "tag1,tag2" or --tags "tag3".
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

* [metalctl machine](metalctl_machine.md)	 - manage machine entities

