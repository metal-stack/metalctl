## metalctl tenant create

creates the tenant

```
metalctl tenant create [flags]
```

### Options

```
      --annotations strings     add initial annotations, must be in the form of key=value, can be given multiple times to add multiple annotations, e.g. --annotation key=value --annotation foo=bar
      --bulk-output             when used with --file (bulk operation): prints results at the end as a list. default is printing results intermediately during the operation, which causes single entities to be printed in a row.
      --cluster-quota int32     cluster quota
      --description string      description of the tenant.
  -f, --file string             filename of the create or update request in yaml format, or - for stdin.
                                
                                Example:
                                $ metalctl tenant describe tenant-1 -o yaml > tenant.yaml
                                $ vi tenant.yaml
                                $ # either via stdin
                                $ cat tenant.yaml | metalctl tenant create -f -
                                $ # or via file
                                $ metalctl tenant create -f tenant.yaml
                                
                                the file can also contain multiple documents and perform a bulk operation.
                                	
  -h, --help                    help for create
      --id string               id of the tenant, max 10 characters.
      --ip-quota int32          ip quota
      --labels strings          add initial label, can be given multiple times to add multiple labels, e.g. --label=foo --label=bar
      --machine-quota int32     machine quota
      --name string             name of the tenant, max 10 characters.
      --skip-security-prompts   skips security prompt for bulk operations
      --timestamps              when used with --file (bulk operation): prints timestamps in-between the operations
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

* [metalctl tenant](metalctl_tenant.md)	 - manage tenant entities

