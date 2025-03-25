## metalctl image

manage image entities

### Synopsis

os images available to be installed on machines.

### Options

```
  -h, --help   help for image
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
* [metalctl image apply](metalctl_image_apply.md)	 - applies one or more images from a given file
* [metalctl image create](metalctl_image_create.md)	 - creates the image
* [metalctl image delete](metalctl_image_delete.md)	 - deletes the image
* [metalctl image describe](metalctl_image_describe.md)	 - describes the image
* [metalctl image edit](metalctl_image_edit.md)	 - edit the image through an editor and update
* [metalctl image list](metalctl_image_list.md)	 - list all images
* [metalctl image update](metalctl_image_update.md)	 - updates the image

