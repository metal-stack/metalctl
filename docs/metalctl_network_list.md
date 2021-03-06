## metalctl network list

list all networks

### Synopsis

list all networks

```
metalctl network list [flags]
```

### Options

```
      --destination-prefixes strings   destination prefixes to filter, use it like: --destination-prefixes prefix1,prefix2.
  -h, --help                           help for list
      --id string                      ID to filter [optional]
      --name string                    name to filter [optional]
      --nat                            nat to filter [optional]
      --parent string                  parent network to filter [optional]
      --partition string               partition to filter [optional]
      --prefixes strings               prefixes to filter, use it like: --prefixes prefix1,prefix2.
      --privatesuper                   privatesuper to filter [optional]
      --project string                 project to filter [optional]
      --underlay                       underlay to filter [optional]
      --vrf int                        vrf to filter [optional]
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
  -f, --file string            filename of the create or update request in yaml format, or - for stdin.
                               Example image update:
                               
                               # metalctl image describe ubuntu-19.04 > ubuntu.yaml
                               # vi ubuntu.yaml
                               ## either via stdin
                               # cat ubuntu.yaml | metalctl image update -f -
                               ## or via file
                               # metalctl image update -f ubuntu.yaml
                               
      --kubeconfig string      Path to the kube-config to use for authentication and authorization. Is updated by login.
      --no-headers             do not print headers of table output format (default print headers)
      --order string           order by (comma separated) column(s), possible values: size|id|status|event|when|partition|project
  -o, --output-format string   output format (table|wide|markdown|json|yaml|template), wide is a table with more columns. (default "table")
      --template string        output template for template output-format, go template format.
                               For property names inspect the output of -o json or -o yaml for reference.
                               Example for machines:
                               
                               metalctl machine list -o template --template "{{ .id }}:{{ .size.id  }}"
                               
                               
  -u, --url string             api server address. Can be specified with METALCTL_URL environment variable.
```

### SEE ALSO

* [metalctl network](metalctl_network.md)	 - manage networks

###### Auto generated by spf13/cobra on 14-Aug-2020
