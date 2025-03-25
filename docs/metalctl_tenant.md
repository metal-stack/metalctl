## metalctl tenant

manage tenant entities

### Synopsis

a tenant belongs to a tenant and groups together entities in metal-stack.

### Options

```
  -h, --help   help for tenant
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

* [metalctl](metalctl.md)	 - a cli to manage entities in the metal-stack api
* [metalctl tenant apply](metalctl_tenant_apply.md)	 - applies one or more tenants from a given file
* [metalctl tenant create](metalctl_tenant_create.md)	 - creates the tenant
* [metalctl tenant delete](metalctl_tenant_delete.md)	 - deletes the tenant
* [metalctl tenant describe](metalctl_tenant_describe.md)	 - describes the tenant
* [metalctl tenant edit](metalctl_tenant_edit.md)	 - edit the tenant through an editor and update
* [metalctl tenant list](metalctl_tenant_list.md)	 - list all tenants
* [metalctl tenant update](metalctl_tenant_update.md)	 - updates the tenant

