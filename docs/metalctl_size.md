## metalctl size

manage size entities

### Synopsis

a size matches a machine in terms of cpu cores, ram and storage.

### Options

```
  -h, --help   help for size
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
* [metalctl size apply](metalctl_size_apply.md)	 - applies one or more sizes from a given file
* [metalctl size create](metalctl_size_create.md)	 - creates the size
* [metalctl size delete](metalctl_size_delete.md)	 - deletes the size
* [metalctl size describe](metalctl_size_describe.md)	 - describes the size
* [metalctl size edit](metalctl_size_edit.md)	 - edit the size through an editor and update
* [metalctl size imageconstraint](metalctl_size_imageconstraint.md)	 - manage imageconstraint entities
* [metalctl size list](metalctl_size_list.md)	 - list all sizes
* [metalctl size reservations](metalctl_size_reservations.md)	 - manage size reservations
* [metalctl size suggest](metalctl_size_suggest.md)	 - suggest size from a given machine id
* [metalctl size update](metalctl_size_update.md)	 - updates the size

