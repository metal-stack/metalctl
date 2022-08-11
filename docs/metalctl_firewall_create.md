## metalctl firewall create

creates the firewall

```
metalctl firewall create [flags]
```

### Options

```
  -d, --description string        Description of the firewall to create. [optional]
  -f, --file string               filename of the create or update request in yaml format, or - for stdin.
                                  
                                  Example:
                                  $ metalctl firewall describe firewall-1 -o yaml > firewall.yaml
                                  $ vi firewall.yaml
                                  $ # either via stdin
                                  $ cat firewall.yaml | metalctl firewall create -f -
                                  $ # or via file
                                  $ metalctl firewall create -f firewall.yaml
                                  	
      --filesystemlayout string   Filesystemlayout to use during machine installation. [optional]
  -h, --help                      help for create
  -H, --hostname string           Hostname of the firewall. [required]
  -I, --id string                 ID of a specific firewall to allocate, if given, size and partition are ignored. Need to be set to reserved (--reserve) state before.
  -i, --image string              OS Image to install. [required]
      --ips strings               Sets the firewall's IP address. Usage: [--ips[=IPV4-ADDRESS[,IPV4-ADDRESS]...]]...
                                  IPV4-ADDRESS specifies the IPv4 address to add.
                                  It can only be used in conjunction with --networks.
  -n, --name string               Name of the firewall. [optional]
      --networks strings          Adds network(s). Usage: --networks NETWORK[:MODE][,NETWORK[:MODE]]... [--networks NETWORK[:MODE][,
                                  NETWORK[:MODE]]...]...
                                  NETWORK specifies the id of an existing network.
                                  MODE can be omitted or one of:
                                  	auto	IP address is automatically acquired from the given network
                                  	noauto	No automatic IP address acquisition
  -S, --partition string          partition/datacenter where the firewall is created. [required, except for reserved machines]
  -P, --project string            Project where the firewall should belong to. [required]
  -s, --size string               Size of the firewall. [required, except for reserved machines]
  -p, --sshpublickey string       SSH public key for access via ssh and console. [optional]
                                  Can be either the public key as string, or pointing to the public key file to use e.g.: "@~/.ssh/id_rsa.pub".
                                  If ~/.ssh/[id_ed25519.pub | id_rsa.pub | id_dsa.pub] is present it will be picked as default, matching the first one in this order.
      --tags strings              tags to add to the firewall, use it like: --tags "tag1,tag2" or --tags "tag3".
      --userdata string           cloud-init.io compatible userdata. [optional]
                                  Can be either the userdata as string, or pointing to the userdata file to use e.g.: "@/tmp/userdata.cfg".
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

* [metalctl firewall](metalctl_firewall.md)	 - manage firewall entities

###### Auto generated by spf13/cobra on 11-Aug-2022
