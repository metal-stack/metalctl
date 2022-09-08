## metalctl switch

manage switch entities

### Synopsis

switch are the leaf switches in the data center that are controlled by metal-stack.

### Options

```
  -h, --help   help for switch
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

* [metalctl](metalctl.md)	 - a cli to manage entities in the metal-stack api
* [metalctl switch delete](metalctl_switch_delete.md)	 - deletes the switch
* [metalctl switch describe](metalctl_switch_describe.md)	 - describes the switch
* [metalctl switch detail](metalctl_switch_detail.md)	 - switch details
* [metalctl switch edit](metalctl_switch_edit.md)	 - edit the switch through an editor and update
* [metalctl switch list](metalctl_switch_list.md)	 - list all switches
* [metalctl switch replace](metalctl_switch_replace.md)	 - put a leaf switch into replace mode in preparation for physical replacement. For a description of the steps involved see the long help.
* [metalctl switch update](metalctl_switch_update.md)	 - updates the switch

