module github.com/metal-stack/metalctl

go 1.13

require (
	github.com/dustin/go-humanize v1.0.0
	github.com/fatih/color v1.9.0
	github.com/go-openapi/errors v0.19.3
	github.com/go-openapi/loads v0.19.5 // indirect
	github.com/go-openapi/strfmt v0.19.4
	github.com/go-openapi/swag v0.19.7
	github.com/go-openapi/validate v0.19.6
	github.com/lestrrat-go/jwx v0.9.1 // indirect
	github.com/metal-stack/metal-go v0.3.3
	github.com/metal-stack/metal-lib v0.3.4
	github.com/metal-stack/updater v1.0.1
	github.com/metal-stack/v v1.0.2
	github.com/olekukonko/tablewriter v0.0.4
	github.com/pelletier/go-toml v1.6.0 // indirect
	github.com/pkg/errors v0.9.1
	github.com/spf13/afero v1.2.2 // indirect
	github.com/spf13/cast v1.3.1 // indirect
	github.com/spf13/cobra v0.0.6
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/spf13/viper v1.6.2
	github.com/stretchr/testify v1.5.1
	go.mongodb.org/mongo-driver v1.3.1 // indirect
	golang.org/x/crypto v0.0.0-20200302210943-78000ba7a073
	golang.org/x/net v0.0.0-20200301022130-244492dfa37a // indirect
	golang.org/x/sys v0.0.0-20200302150141-5c8b2ff67527 // indirect
	gopkg.in/ini.v1 v1.54.0 // indirect
	gopkg.in/yaml.v3 v3.0.0-20200121175148-a6ecf24a6d71
)

replace github.com/coreos/etcd => github.com/coreos/etcd v3.3.18+incompatible
