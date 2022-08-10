package printers

import (
	"log"

	"github.com/fatih/color"
	"github.com/metal-stack/metal-lib/pkg/genericcli"
	"github.com/metal-stack/metalctl/cmd/printers/tableprinters"
	"github.com/spf13/viper"
)

func NewPrinterFromCLI() genericcli.Printer {
	var printer genericcli.Printer
	var err error

	switch format := viper.GetString("output-format"); format {
	case "yaml":
		printer = genericcli.NewYAMLPrinter()
	case "json":
		printer = genericcli.NewJSONPrinter()
	case "table", "wide", "markdown":
		tp := tableprinters.New()
		cfg := &genericcli.TablePrinterConfig{
			ToHeaderAndRows: tp.ToHeaderAndRows,
			Wide:            format == "wide",
			Markdown:        format == "markdown",
			NoHeaders:       viper.GetBool("no-headers"),
		}
		tablePrinter, err := genericcli.NewTablePrinter(cfg)
		if err != nil {
			log.Fatalf("unable to initialize printer: %v", err)
		}
		tp.SetPrinter(tablePrinter)
		printer = tablePrinter
	case "template":
		printer, err = genericcli.NewTemplatePrinter(viper.GetString("template"))
		if err != nil {
			log.Fatalf("unable to initialize printer: %v", err)
		}
	default:
		log.Fatalf("unknown output format: %q", format)
	}

	if viper.IsSet("force-color") {
		enabled := viper.GetBool("force-color")
		if enabled {
			color.NoColor = false
		} else {
			color.NoColor = true
		}
	}

	return printer
}

func DefaultToYAMLPrinter() genericcli.Printer {
	if viper.IsSet("output-format") {
		return NewPrinterFromCLI()
	}
	return genericcli.NewYAMLPrinter()
}
