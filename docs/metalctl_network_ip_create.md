## metalctl network ip create

creates the ip

```
metalctl network ip create [flags]
```

### Options

```
      --addressfamily string    addressfamily of the ip to acquire, defaults to IPv4 [optional]
      --bulk-output             when used with --file (bulk operation): prints results at the end as a list. default is printing results intermediately during the operation, which causes single entities to be printed in a row.
  -d, --description string      description of the IP to allocate. [optional]
  -f, --file string             filename of the create or update request in yaml format, or - for stdin.
                                
                                Example:
                                $ metalctl ip describe ip-1 -o yaml > ip.yaml
                                $ vi ip.yaml
                                $ # either via stdin
                                $ cat ip.yaml | metalctl ip create -f -
                                $ # or via file
                                $ metalctl ip create -f ip.yaml
                                
                                the file can also contain multiple documents and perform a bulk operation.
                                	
  -h, --help                    help for create
      --ipaddress string        a specific ip address to allocate. [optional]
  -n, --name string             name of the IP to allocate. [optional]
      --network string          network from where the IP should be allocated.
      --project string          project for which the IP should be allocated.
      --skip-security-prompts   skips security prompt for bulk operations
      --tags strings            tags to attach to the IP.
      --timestamps              when used with --file (bulk operation): prints timestamps in-between the operations
      --type string             type of the IP to allocate: ephemeral|static [optional] (default "ephemeral")
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

* [metalctl network ip](metalctl_network_ip.md)	 - manage ip entities

