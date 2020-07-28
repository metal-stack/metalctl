module github.com/metal-stack/metalctl

go 1.13

require (
	github.com/dustin/go-humanize v1.0.0
	github.com/fatih/color v1.9.0
	github.com/metal-stack/metal-go v0.8.2-0.20200728041954-234e0bc8dc52
	github.com/metal-stack/metal-lib v0.5.0
	github.com/metal-stack/updater v1.1.0
	github.com/metal-stack/v v1.0.2
	github.com/olekukonko/tablewriter v0.0.4
	github.com/pelletier/go-toml v1.8.0 // indirect
	github.com/pkg/errors v0.9.1
	github.com/spf13/afero v1.2.2 // indirect
	github.com/spf13/cast v1.3.1 // indirect
	github.com/spf13/cobra v1.0.0
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/spf13/viper v1.7.0
	github.com/stretchr/testify v1.6.1
	go.mongodb.org/mongo-driver v1.3.4 // indirect
	golang.org/x/crypto v0.0.0-20200622213623-75b288015ac9
	gopkg.in/ini.v1 v1.57.0 // indirect
	gopkg.in/yaml.v3 v3.0.0-20200615113413-eeeca48fe776
)

replace github.com/coreos/etcd => github.com/coreos/etcd v3.3.18+incompatible
