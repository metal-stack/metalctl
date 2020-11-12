module github.com/metal-stack/metalctl

go 1.15

require (
	github.com/dustin/go-humanize v1.0.0
	github.com/fatih/color v1.9.0
	github.com/go-openapi/errors v0.19.8
	github.com/go-openapi/strfmt v0.19.8
	github.com/go-openapi/swag v0.19.11
	github.com/go-openapi/validate v0.19.12
	github.com/metal-stack/masterdata-api v0.8.3
	github.com/metal-stack/metal-go v0.11.1-0.20201112083418-6abd86b6d538
	github.com/metal-stack/metal-lib v0.6.4
	github.com/metal-stack/updater v1.1.1
	github.com/metal-stack/v v1.0.2
	github.com/mitchellh/go-homedir v1.1.0
	github.com/olekukonko/tablewriter v0.0.4
	github.com/pelletier/go-toml v1.8.0 // indirect
	github.com/pkg/errors v0.9.1
	github.com/spf13/afero v1.3.3 // indirect
	github.com/spf13/cast v1.3.1 // indirect
	github.com/spf13/cobra v1.1.1
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/viper v1.7.1
	github.com/stretchr/testify v1.6.1
	golang.org/x/crypto v0.0.0-20201012173705-84dcc777aaee
	gopkg.in/ini.v1 v1.57.0 // indirect
	gopkg.in/yaml.v3 v3.0.0-20200615113413-eeeca48fe776
)

replace github.com/coreos/etcd => github.com/coreos/etcd v3.3.18+incompatible
