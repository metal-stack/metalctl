## metalctl size imageconstraint

manage imageconstraint entities

### Synopsis

If a size has specific requirements regarding the images which must fullfil certain constraints, this can be configured here.

### Options

```
  -h, --help   help for imageconstraint
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

* [metalctl size](metalctl_size.md)	 - manage size entities
* [metalctl size imageconstraint apply](metalctl_size_imageconstraint_apply.md)	 - applies one or more imageconstraints from a given file
* [metalctl size imageconstraint create](metalctl_size_imageconstraint_create.md)	 - creates the imageconstraint
* [metalctl size imageconstraint delete](metalctl_size_imageconstraint_delete.md)	 - deletes the imageconstraint
* [metalctl size imageconstraint describe](metalctl_size_imageconstraint_describe.md)	 - describes the imageconstraint
* [metalctl size imageconstraint edit](metalctl_size_imageconstraint_edit.md)	 - edit the imageconstraint through an editor and update
* [metalctl size imageconstraint list](metalctl_size_imageconstraint_list.md)	 - list all imageconstraints
* [metalctl size imageconstraint try](metalctl_size_imageconstraint_try.md)	 - try if size and image can be allocated
* [metalctl size imageconstraint update](metalctl_size_imageconstraint_update.md)	 - updates the imageconstraint

