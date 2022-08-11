package cmd

import (
	"log"

	"github.com/fatih/color"
	"github.com/metal-stack/metal-lib/pkg/genericcli/printers"
	"github.com/metal-stack/metalctl/cmd/tableprinters"
	"github.com/spf13/viper"
)

func newPrinterFromCLI() printers.Printer {
	var printer printers.Printer

	switch format := viper.GetString("output-format"); format {
	case "yaml":
		printer = printers.NewYAMLPrinter()
	case "json":
		printer = printers.NewJSONPrinter()
	case "table", "wide", "markdown":
		tp := tableprinters.New()
		cfg := &printers.TablePrinterConfig{
			ToHeaderAndRows: tp.ToHeaderAndRows,
			Wide:            format == "wide",
			Markdown:        format == "markdown",
			NoHeaders:       viper.GetBool("no-headers"),
		}
		tablePrinter := printers.NewTablePrinter(cfg)
		tp.SetPrinter(tablePrinter)
		printer = tablePrinter
	case "template":
		printer = printers.NewTemplatePrinter(viper.GetString("template"))
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

func defaultToYAMLPrinter() printers.Printer {
	if viper.IsSet("output-format") {
		return newPrinterFromCLI()
	}
	return printers.NewYAMLPrinter()
}
