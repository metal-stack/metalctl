## metalctl machine apply

applies one or more machines from a given file

```
metalctl machine apply [flags]
```

### Options

```
  -f, --file string   filename of the create or update request in yaml format, or - for stdin.
                      
                      Example:
                      $ metalctl machine describe machine-1 -o yaml > machine.yaml
                      $ vi machine.yaml
                      $ # either via stdin
                      $ cat machine.yaml | metalctl machine apply -f -
                      $ # or via file
                      $ metalctl machine apply -f machine.yaml
                      	
  -h, --help          help for apply
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

