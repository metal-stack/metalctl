## metalctl size reservation

manage reservation entities

### Synopsis

manage size reservations

### Options

```
  -h, --help   help for reservation
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

* [metalctl size](metalctl_size.md)	 - manage size entities
* [metalctl size reservation apply](metalctl_size_reservation_apply.md)	 - applies one or more reservations from a given file
* [metalctl size reservation create](metalctl_size_reservation_create.md)	 - creates the reservation
* [metalctl size reservation delete](metalctl_size_reservation_delete.md)	 - deletes the reservation
* [metalctl size reservation describe](metalctl_size_reservation_describe.md)	 - describes the reservation
* [metalctl size reservation edit](metalctl_size_reservation_edit.md)	 - edit the reservation through an editor and update
* [metalctl size reservation list](metalctl_size_reservation_list.md)	 - list all reservations
* [metalctl size reservation update](metalctl_size_reservation_update.md)	 - updates the reservation
* [metalctl size reservation usage](metalctl_size_reservation_usage.md)	 - see current usage of size reservations

