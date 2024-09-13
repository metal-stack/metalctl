## metalctl network create

creates the network

```
metalctl network create [flags]
```

### Options

```
      --additional-announcable-cidrs strings   list of cidrs which are added to the route maps per tenant private network, these are typically pod- and service cidrs, can only be set in a supernetwork
      --bulk-output                            when used with --file (bulk operation): prints results at the end as a list. default is printing results intermediately during the operation, which causes single entities to be printed in a row.
  -d, --description string                     description of the network to create. [optional]
      --destination-prefixes strings           destination prefixes in this network.
  -f, --file string                            filename of the create or update request in yaml format, or - for stdin.
                                               
                                               Example:
                                               $ metalctl network describe network-1 -o yaml > network.yaml
                                               $ vi network.yaml
                                               $ # either via stdin
                                               $ cat network.yaml | metalctl network create -f -
                                               $ # or via file
                                               $ metalctl network create -f network.yaml
                                               
                                               the file can also contain multiple documents and perform a bulk operation.
                                               	
  -h, --help                                   help for create
      --id string                              id of the network to create. [optional]
      --labels strings                         add initial labels, must be in the form of key=value, use it like: --labels "key1=value1,key2=value2".
  -n, --name string                            name of the network to create. [optional]
      --nat                                    set nat flag of network, if set to true, traffic from this network will be natted.
  -p, --partition string                       partition where this network should exist.
      --prefixes strings                       prefixes in this network.
      --privatesuper                           set private super flag of network, if set to true, this network is used to start machines there.
      --project string                         project of the network to create. [optional]
      --skip-security-prompts                  skips security prompt for bulk operations
      --timestamps                             when used with --file (bulk operation): prints timestamps in-between the operations
      --underlay                               set underlay flag of network, if set to true, this is used to transport underlay network traffic
      --vrf int                                vrf of this network
      --vrfshared                              vrf shared allows multiple networks to share a vrf
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

* [metalctl network](metalctl_network.md)	 - manage network entities

