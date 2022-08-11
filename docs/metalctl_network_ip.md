## metalctl network ip

manage ip entities

### Synopsis

an ip address can be attached to a machine or firewall such that network traffic can be routed to these servers.

### Options

```
  -h, --help   help for ip
```

### Options inherited from parent commands

```
      --apitoken string        api token to authenticate. Can be specified with METALCTL_APITOKEN environment variable.
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
                               
                               
  -u, --url string             api server address. Can be specified with METALCTL_URL environment variable.
      --yes-i-really-mean-it   skips security prompts (which can be dangerous to set blindly because actions can lead to data loss or additional costs)
```

### SEE ALSO

* [metalctl network](metalctl_network.md)	 - manage network entities
* [metalctl network ip apply](metalctl_network_ip_apply.md)	 - applies one or more ips from a given file
* [metalctl network ip create](metalctl_network_ip_create.md)	 - creates the ip
* [metalctl network ip delete](metalctl_network_ip_delete.md)	 - deletes the ip
* [metalctl network ip describe](metalctl_network_ip_describe.md)	 - describes the ip
* [metalctl network ip edit](metalctl_network_ip_edit.md)	 - edit the ip through an editor and update
* [metalctl network ip issues](metalctl_network_ip_issues.md)	 - display ips which are in a potential bad state
* [metalctl network ip list](metalctl_network_ip_list.md)	 - list all ips
* [metalctl network ip update](metalctl_network_ip_update.md)	 - updates the ip

