## metalctl network

manage network entities

### Synopsis

networks can be attached to a machine or firewall such that they can communicate with each other.

### Options

```
  -h, --help   help for network
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

* [metalctl](metalctl.md)	 - a cli to manage entities in the metal-stack api
* [metalctl network allocate](metalctl_network_allocate.md)	 - allocate a network
* [metalctl network apply](metalctl_network_apply.md)	 - applies one or more networks from a given file
* [metalctl network create](metalctl_network_create.md)	 - creates the network
* [metalctl network delete](metalctl_network_delete.md)	 - deletes the network
* [metalctl network describe](metalctl_network_describe.md)	 - describes the network
* [metalctl network edit](metalctl_network_edit.md)	 - edit the network through an editor and update
* [metalctl network free](metalctl_network_free.md)	 - free a network
* [metalctl network ip](metalctl_network_ip.md)	 - manage ip entities
* [metalctl network list](metalctl_network_list.md)	 - list all networks
* [metalctl network update](metalctl_network_update.md)	 - updates the network

