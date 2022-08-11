## metalctl partition

manage partition entities

### Synopsis

a partition is a group of machines and network which is logically separated from other partitions. Machines have no direct network connections between partitions.

### Options

```
  -h, --help   help for partition
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

* [metalctl](metalctl.md)	 - a cli to manage entities in the metal-stack api
* [metalctl partition apply](metalctl_partition_apply.md)	 - applies one or more partitions from a given file
* [metalctl partition capacity](metalctl_partition_capacity.md)	 - show partition capacity
* [metalctl partition create](metalctl_partition_create.md)	 - creates the partition
* [metalctl partition delete](metalctl_partition_delete.md)	 - deletes the partition
* [metalctl partition describe](metalctl_partition_describe.md)	 - describes the partition
* [metalctl partition edit](metalctl_partition_edit.md)	 - edit the partition through an editor and update
* [metalctl partition list](metalctl_partition_list.md)	 - list all partitions
* [metalctl partition update](metalctl_partition_update.md)	 - updates the partition

