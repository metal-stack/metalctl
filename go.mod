module github.com/metal-stack/metalctl

go 1.15

require (
	github.com/dustin/go-humanize v1.0.0
	github.com/fatih/color v1.10.0
	github.com/metal-stack/masterdata-api v0.8.4
	github.com/metal-stack/metal-go v0.13.0
	github.com/metal-stack/metal-lib v0.7.0
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
	github.com/stretchr/testify v1.7.0
	golang.org/x/crypto v0.0.0-20201221181555-eec23a3978ad
	golang.org/x/term v0.0.0-20201117132131-f5c789dd3221
	gopkg.in/ini.v1 v1.57.0 // indirect
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
)

replace github.com/coreos/etcd => github.com/coreos/etcd v3.3.18+incompatible
