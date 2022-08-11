## metalctl machine

manage machine entities

### Synopsis

a machine is a bare metal server provisioned through metal-stack that is intended to run user workload.

### Options

```
  -h, --help   help for machine
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

* [metalctl](metalctl.md)	 - a cli to manage entities in the metal-stack api
* [metalctl machine apply](metalctl_machine_apply.md)	 - applies one or more machines from a given file
* [metalctl machine console](metalctl_machine_console.md)	 - console access to a machine
* [metalctl machine consolepassword](metalctl_machine_consolepassword.md)	 - fetch the consolepassword for a machine
* [metalctl machine create](metalctl_machine_create.md)	 - creates the machine
* [metalctl machine delete](metalctl_machine_delete.md)	 - deletes the machine
* [metalctl machine describe](metalctl_machine_describe.md)	 - describes the machine
* [metalctl machine edit](metalctl_machine_edit.md)	 - edit the machine through an editor and update
* [metalctl machine identify](metalctl_machine_identify.md)	 - manage machine chassis identify LED power
* [metalctl machine ipmi](metalctl_machine_ipmi.md)	 - display ipmi details of the machine, if no machine ID is given all ipmi addresses are returned.
* [metalctl machine issues](metalctl_machine_issues.md)	 - display machines which are in a potential bad state
* [metalctl machine list](metalctl_machine_list.md)	 - list all machines
* [metalctl machine lock](metalctl_machine_lock.md)	 - lock a machine
* [metalctl machine logs](metalctl_machine_logs.md)	 - display machine provisioning logs
* [metalctl machine power](metalctl_machine_power.md)	 - manage machine power
* [metalctl machine reinstall](metalctl_machine_reinstall.md)	 - reinstalls an already allocated machine
* [metalctl machine reserve](metalctl_machine_reserve.md)	 - reserve a machine
* [metalctl machine update](metalctl_machine_update.md)	 - updates the machine
* [metalctl machine update-firmware](metalctl_machine_update-firmware.md)	 - update a machine firmware

