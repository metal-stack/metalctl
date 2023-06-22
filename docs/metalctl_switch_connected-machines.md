## metalctl switch connected-machines

shows switches with their connected machines

```
metalctl switch connected-machines [flags]
```

### Examples

```
The command will show the machines connected to the switch ports.

Can also be used with -o template in order to generate CSV-style output:

$ metalctl switch connected-machines -o template --template '{{ $machines := .machines }}{{ range .switches }}{{ $switch := . }}{{ range .connections }}{{ $switch.id }},{{ $switch.rack_id }},{{ .nic.name }},{{ .machine_id }},{{ (index $machines .machine_id).ipmi.fru.product_serial }}{{ printf "\n" }}{{ end }}{{ end }}'
r01leaf01,swp1,f78cc340-e5e8-48ed-8fe7-2336c1e2ded2,<a-serial>
r01leaf01,swp2,44e3a522-5f48-4f3c-9188-41025f9e401e,<b-serial>
...

```

### Options

```
  -h, --help                help for connected-machines
      --id string           ID of the switch.
      --name string         Name of the switch.
      --os-vendor string    OS vendor of this switch.
      --os-version string   OS version of this switch.
      --partition string    Partition of this switch.
      --rack string         Rack of this switch.
      --size string         Size of the connectedmachines.
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

* [metalctl switch](metalctl_switch.md)	 - manage switch entities

