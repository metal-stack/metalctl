## metalctl machine create

create a machine

### Synopsis

create a new machine with the given operating system, the size and a project.

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
  -d, --description string    Description of the machine to create. [optional]
  -h, --help                  help for create
  -H, --hostname string       Hostname of the machine. [required]
  -I, --id string             ID of a specific machine to allocate, if given, size and partition are ignored. Need to be set to reserved (--reserve) state before.
  -i, --image string          OS Image to install. [required]
      --ips strings           Sets the machine's IP address. Usage: [--ips[=IPV4-ADDRESS[,IPV4-ADDRESS]...]]...
                              IPV4-ADDRESS specifies the IPv4 address to add.
                              It can only be used in conjunction with --networks.
  -n, --name string           Name of the machine. [optional]
      --networks strings      Adds a network. Usage: [--networks NETWORK[:MODE][,NETWORK[:MODE]]...]...
                              NETWORK specifies the name or id of an existing network.
                              MODE cane be omitted or one of:
                              	auto	IP address is automatically acquired from the given network
                              	noauto	IP address for the given network must be provided via --ips
  -S, --partition string      partition/datacenter where the machine is created. [required, except for reserved machines]
  -P, --project string        Project where the machine should belong to. [required]
  -s, --size string           Size of the machine. [required, except for reserved machines]
  -p, --sshpublickey string   SSH public key for access via ssh and console. [optional]
                              Can be either the public key as string, or pointing to the public key file to use e.g.: "@~/.ssh/id_rsa.pub".
                              If ~/.ssh/id_rsa.pub is present it will be picked as default.
      --tags strings          tags to add to the machine, use it like: --tags "tag1,tag2" or --tags "tag3".
      --userdata string       cloud-init.io compatible userdata. [optional]
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

* [metalctl machine](metalctl_machine.md)	 - manage machines

###### Auto generated by spf13/cobra on 21-Apr-2020
