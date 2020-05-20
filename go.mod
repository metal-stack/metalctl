module github.com/metal-stack/metalctl

go 1.13

require (
	github.com/dustin/go-humanize v1.0.0
	github.com/fatih/color v1.9.0
	github.com/go-openapi/loads v0.19.5 // indirect
	github.com/metal-stack/metal-go v0.7.3
	github.com/metal-stack/metal-lib v0.4.0
	github.com/metal-stack/updater v1.0.1
	github.com/metal-stack/v v1.0.2
	github.com/olekukonko/tablewriter v0.0.4
	github.com/pelletier/go-toml v1.7.0 // indirect
	github.com/pkg/errors v0.9.1
	github.com/spf13/afero v1.2.2 // indirect
	github.com/spf13/cast v1.3.1 // indirect
	github.com/spf13/cobra v1.0.0
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/spf13/viper v1.6.3
	github.com/stretchr/testify v1.5.1
	go.mongodb.org/mongo-driver v1.3.3 // indirect
	golang.org/x/crypto v0.0.0-20200429183012-4b2356b1ed79
	gopkg.in/ini.v1 v1.55.0 // indirect
	gopkg.in/yaml.v3 v3.0.0-20200506231410-2ff61e1afc86
)

replace github.com/coreos/etcd => github.com/coreos/etcd v3.3.18+incompatible
