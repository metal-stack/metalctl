## metalctl network allocate

allocate a network

```
metalctl network allocate [flags]
```

### Options

```
      --addressfamily string     addressfamily of the network to acquire, if not specified the network inherits the address families from the parent [optional]
  -d, --description string       description of the network to create. [optional]
      --dmz                      use this private network as dmz. [optional]
  -h, --help                     help for allocate
      --ipv4-prefix-length int   ipv4 prefix bit length of the network to create, defaults to default child prefix length of the parent network. [optional]
      --ipv6-prefix-length int   ipv6 prefix bit length of the network to create, defaults to default child prefix length of the parent network. [optional]
      --labels strings           labels for this network. [optional]
  -n, --name string              name of the network to create. [required]
      --partition string         partition where this network should exist. [required]
      --project string           partition where this network should exist. [required]
      --shared                   shared allows usage of this private network from other networks
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

* [metalctl network](metalctl_network.md)	 - manage network entities

