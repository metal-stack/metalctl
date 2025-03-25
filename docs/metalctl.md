## metalctl

a cli to manage entities in the metal-stack api

### Options

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
  -h, --help                   help for metalctl
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

* [metalctl audit](metalctl_audit.md)	 - manage audit trace entities
* [metalctl completion](metalctl_completion.md)	 - Generate the autocompletion script for the specified shell
* [metalctl context](metalctl_context.md)	 - manage metalctl context
* [metalctl filesystemlayout](metalctl_filesystemlayout.md)	 - manage filesystemlayout entities
* [metalctl firewall](metalctl_firewall.md)	 - manage firewall entities
* [metalctl firmware](metalctl_firmware.md)	 - manage firmwares
* [metalctl health](metalctl_health.md)	 - shows the server health
* [metalctl image](metalctl_image.md)	 - manage image entities
* [metalctl login](metalctl_login.md)	 - login user and receive token
* [metalctl logout](metalctl_logout.md)	 - logout user from OIDC SSO session
* [metalctl machine](metalctl_machine.md)	 - manage machine entities
* [metalctl markdown](metalctl_markdown.md)	 - create markdown documentation
* [metalctl network](metalctl_network.md)	 - manage network entities
* [metalctl partition](metalctl_partition.md)	 - manage partition entities
* [metalctl project](metalctl_project.md)	 - manage project entities
* [metalctl size](metalctl_size.md)	 - manage size entities
* [metalctl switch](metalctl_switch.md)	 - manage switch entities
* [metalctl tenant](metalctl_tenant.md)	 - manage tenant entities
* [metalctl update](metalctl_update.md)	 - update the program
* [metalctl version](metalctl_version.md)	 - print the client and server version information
* [metalctl vpn](metalctl_vpn.md)	 - access VPN
* [metalctl whoami](metalctl_whoami.md)	 - shows current user

