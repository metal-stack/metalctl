## metalctl audit list

list all audit traces

```
metalctl audit list [flags]
```

### Options

````
      --component string       component of the audit trace.
      --detail string          detail of the audit trace. An HTTP method, unary or stream
      --error string           error of the audit trace.
      --forwarded-for string   forwarded for of the audit trace.
      --from string            start of range of the audit traces. e.g. 1h, 10m, 2006-01-02 15:04:05
  -h, --help                   help for list
      --limit int              limit the number of audit traces. (default 100)
      --path string            api path of the audit trace.
      --phase string           phase of the audit trace. One of [request, response, single, error, opened, closed]
  -q, --query string           filters audit trace body payloads for the given text.
      --remote-addr string     remote address of the audit trace.
      --request-id string      request id of the audit trace.
      --sort-by strings        sort by (comma separated) column(s), sort direction can be changed by appending :asc or :desc behind the column identifier. possible values: path|tenant|timestamp|user
      --status-code int32      HTTP status code of the audit trace.
      --tenant string          tenant of the audit trace.
      --to string              end of range of the audit traces. e.g. 1h, 10m, 2006-01-02 15:04:05
      --type string            type of the audit trace. One of [http, grpc, event].
      --user string            user of the audit trace.```
````

### Options inherited from parent commands

```

      --api-token string       api token to authenticate. Can be specified with METALCTL_API_TOKEN environment variable.
      --api-url string         api server address. Can be specified with METALCTL_API_URL environment variable.

-c, --config string alternative config file path, (default is ~/.metalctl/config.yaml).
Example config.yaml:

                               ---
                               apitoken: "alongtoken"
                               ...


      --debug                  debug output
      --force-color            force colored output even without tty
      --kubeconfig string      Path to the kube-config to use for authentication and authorization. Is updated by login. Uses default path if not specified.
      --no-headers             do not print headers of table output format (default print headers)

-o, --output-format string output format (table|wide|markdown|json|yaml|template), wide is a table with more columns. (default "table")
--template string output template for template output-format, go template format.
For property names inspect the output of -o json or -o yaml for reference.
Example for machines:

                               metalctl machine list -o template --template "{{ .id }}:{{ .size.id  }}"


      --yes-i-really-mean-it   skips security prompts (which can be dangerous to set blindly because actions can lead to data loss or additional costs)

```

### SEE ALSO

- [metalctl](metalctl.md) - a cli to manage entities in the metal-stack api
- [metalctl audit list](metalctl_audit_list.md) - list all audit traces
- [metalctl audit describe](metalctl_audit_describe.md) - describe an audit trace
