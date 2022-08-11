## metalctl filesystemlayout

manage filesystemlayout entities

### Synopsis

a filesystemlayout is a specification how the disks in a machine are partitioned, formatted and mounted.

### Options

```
  -h, --help   help for filesystemlayout
```

### Options inherited from parent commands

```
      --api-token string       api token to authenticate. Can be specified with METALCTL_APITOKEN environment variable.
      --api-url string         api server address. Can be specified with METALCTL_URL environment variable.
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
* [metalctl filesystemlayout apply](metalctl_filesystemlayout_apply.md)	 - applies one or more filesystemlayouts from a given file
* [metalctl filesystemlayout create](metalctl_filesystemlayout_create.md)	 - creates the filesystemlayout
* [metalctl filesystemlayout delete](metalctl_filesystemlayout_delete.md)	 - deletes the filesystemlayout
* [metalctl filesystemlayout describe](metalctl_filesystemlayout_describe.md)	 - describes the filesystemlayout
* [metalctl filesystemlayout edit](metalctl_filesystemlayout_edit.md)	 - edit the filesystemlayout through an editor and update
* [metalctl filesystemlayout list](metalctl_filesystemlayout_list.md)	 - list all filesystemlayouts
* [metalctl filesystemlayout match](metalctl_filesystemlayout_match.md)	 - check if a machine satisfies all disk requirements of a given filesystemlayout
* [metalctl filesystemlayout try](metalctl_filesystemlayout_try.md)	 - try to detect a filesystem by given size and image
* [metalctl filesystemlayout update](metalctl_filesystemlayout_update.md)	 - updates the filesystemlayout

