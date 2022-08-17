## metalctl project

manage project entities

### Synopsis

a project belongs to a tenant and groups together entities in metal-stack.

### Options

```
  -h, --help   help for project
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
* [metalctl project apply](metalctl_project_apply.md)	 - applies one or more projects from a given file
* [metalctl project create](metalctl_project_create.md)	 - creates the project
* [metalctl project delete](metalctl_project_delete.md)	 - deletes the project
* [metalctl project describe](metalctl_project_describe.md)	 - describes the project
* [metalctl project edit](metalctl_project_edit.md)	 - edit the project through an editor and update
* [metalctl project list](metalctl_project_list.md)	 - list all projects
* [metalctl project update](metalctl_project_update.md)	 - updates the project

