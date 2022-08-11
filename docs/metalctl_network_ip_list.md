## metalctl network ip list

list all ips

```
metalctl network ip list [flags]
```

### Options

```
  -h, --help               help for list
      --ipaddress string   ipaddress to filter [optional]
      --machineid string   machineid to filter [optional]
      --name string        name to filter [optional]
      --network string     network to filter [optional]
      --order strings      order by (comma separated) column(s), sort direction can be changed by appending :asc or :desc behind the column identifier. possible values: description|id|ipaddress|name|network|type
      --prefix string      prefix to filter [optional]
      --project string     project to filter [optional]
      --tags strings       tags to filter [optional]
      --type string        type to filter [optional]
```

### Options inherited from parent commands

```
      --api-token string       api token to authenticate. Can be specified with METALCTL_APITOKEN environment variable.
      --api-url string         api server address. Can be specified with METALCTL_URL environment variable.
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

* [metalctl network ip](metalctl_network_ip.md)	 - manage ip entities

