## metalctl machine create

creates the machine

```
metalctl machine create [flags]
```

### Examples

```
machine create can be done in two different ways:

- default with automatic allocation:

	metalctl machine create \
		--hostname worker01 \
		--name worker \
		--image ubuntu-18.04 \ # query available with: metalctl image list
		--size t1-small-x86 \  # query available with: metalctl size list
		--partition test \     # query available with: metalctl partition list
		--project cluster01 \
		--sshpublickey "@~/.ssh/id_rsa.pub"

- for metal administration with reserved machines:

	reserve a machine you want to allocate:

	metalctl machine reserve 00000000-0000-0000-0000-0cc47ae54694 --description "blocked for maintenance"

	allocate this machine:

	metalctl machine create \
		--hostname worker01 \
		--name worker \
		--image ubuntu-18.04 \ # query available with: metalctl image list
		--project cluster01 \
		--sshpublickey "@~/.ssh/id_rsa.pub" \
		--id 00000000-0000-0000-0000-0cc47ae54694

after you do not want to use this machine exclusive, remove the reservation:

metalctl machine reserve 00000000-0000-0000-0000-0cc47ae54694 --remove

Once created the machine installation can not be modified anymore.

```

### Options

```
      --bulk-output               when used with --file (bulk operation): prints results at the end as a list. default is printing results intermediately during the operation, which causes single entities to be printed in a row.
  -d, --description string        Description of the machine to create. [optional]
      --dnsservers strings        dns servers to add to the machine or firewall. [optional]
  -f, --file string               filename of the create or update request in yaml format, or - for stdin.
                                  
                                  Example:
                                  $ metalctl machine describe machine-1 -o yaml > machine.yaml
                                  $ vi machine.yaml
                                  $ # either via stdin
                                  $ cat machine.yaml | metalctl machine create -f -
                                  $ # or via file
                                  $ metalctl machine create -f machine.yaml
                                  
                                  the file can also contain multiple documents and perform a bulk operation.
                                  	
      --filesystemlayout string   Filesystemlayout to use during machine installation. [optional]
  -h, --help                      help for create
  -H, --hostname string           Hostname of the machine. [required]
  -I, --id string                 ID of a specific machine to allocate, if given, size and partition are ignored. Need to be set to reserved (--reserve) state before.
  -i, --image string              OS Image to install. [required]
      --ips strings               Sets the machine's IP address. Usage: [--ips[=IPV4-ADDRESS[,IPV4-ADDRESS]...]]...
                                  IPV4-ADDRESS specifies the IPv4 address to add.
                                  It can only be used in conjunction with --networks.
  -n, --name string               Name of the machine. [optional]
      --networks strings          Adds a network. Usage: [--networks NETWORK[:MODE][,NETWORK[:MODE]]...]...
                                  NETWORK specifies the name or id of an existing network.
                                  MODE cane be omitted or one of:
                                  	auto	IP address is automatically acquired from the given network
                                  	noauto	IP address for the given network must be provided via --ips
      --ntpservers strings        ntp servers to add to the machine or firewall. [optional]
  -S, --partition string          partition/datacenter where the machine is created. [required, except for reserved machines]
  -P, --project string            Project where the machine should belong to. [required]
  -s, --size string               Size of the machine. [required, except for reserved machines]
      --skip-security-prompts     skips security prompt for bulk operations
  -p, --sshpublickey string       SSH public key for access via ssh and console. [optional]
                                  Can be either the public key as string, or pointing to the public key file to use e.g.: "@~/.ssh/id_rsa.pub".
                                  If ~/.ssh/[id_ed25519.pub | id_rsa.pub | id_dsa.pub] is present it will be picked as default, matching the first one in this order.
      --tags strings              tags to add to the machine, use it like: --tags "tag1,tag2" or --tags "tag3".
      --timestamps                when used with --file (bulk operation): prints timestamps in-between the operations
      --userdata string           cloud-init.io compatible userdata. [optional]
                                  Can be either the userdata as string, or pointing to the userdata file to use e.g.: "@/tmp/userdata.cfg".
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

