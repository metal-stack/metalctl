## metalctl partition create

creates the partition

```
metalctl partition create [flags]
```

### Options

```
      --bulk-output             when used with --file (bulk operation): prints results at the end as a list. default is printing results intermediately during the operation, which causes single entities to be printed in a row.
      --cmdline string          kernel commandline for the metal-hammer in the partition. [required]
  -d, --description string      Description of the partition. [required]
      --dnsservers string       dns servers for the machines and firewalls in the partition. [optional]
  -f, --file string             filename of the create or update request in yaml format, or - for stdin.
                                
                                Example:
                                $ metalctl partition describe partition-1 -o yaml > partition.yaml
                                $ vi partition.yaml
                                $ # either via stdin
                                $ cat partition.yaml | metalctl partition create -f -
                                $ # or via file
                                $ metalctl partition create -f partition.yaml
                                
                                the file can also contain multiple documents and perform a bulk operation.
                                	
  -h, --help                    help for create
      --id string               ID of the partition. [required]
      --imageurl string         initrd for the metal-hammer in the partition. [required]
      --kernelurl string        kernel url for the metal-hammer in the partition. [required]
      --mgmtserver string       management server address in the partition. [required]
  -n, --name string             Name of the partition. [optional]
      --ntpservers string       ntp servers for the machines and firewalls in the partition. [optional]
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

* [metalctl partition](metalctl_partition.md)	 - manage partition entities

