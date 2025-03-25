package v2

import (
	"fmt"
	"strings"

	"connectrpc.com/connect"
	apiv2 "github.com/metal-stack/api/go/metalstack/api/v2"
	"github.com/metal-stack/metal-lib/pkg/genericcli"
	"github.com/metal-stack/metal-lib/pkg/genericcli/printers"
	"github.com/metal-stack/metal-lib/pkg/pointer"
	"github.com/metal-stack/metalctl/pkg/api"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type image struct {
	c *api.Config
}

func newImageCmd(c *api.Config) *cobra.Command {
	w := &image{
		c: c,
	}

	gcli := genericcli.NewGenericCLI(w).WithFS(c.FS)

	cmdsConfig := &genericcli.CmdsConfig[any, any, *apiv2.Image]{
		BinaryName:      "metalctl", // FIXME
		GenericCLI:      gcli,
		Singular:        "image",
		Plural:          "images",
		Description:     "manage images which are used to be installed on machines and firewalls",
		DescribePrinter: func() printers.Printer { return c.DescribePrinter },
		ListPrinter:     func() printers.Printer { return c.ListPrinter },
		OnlyCmds:        genericcli.OnlyCmds(genericcli.DescribeCmd, genericcli.ListCmd),
		DescribeCmdMutateFn: func(cmd *cobra.Command) {
			cmd.RunE = func(cmd *cobra.Command, args []string) error {
				return gcli.DescribeAndPrint("", w.c.DescribePrinter)
			}
		},
		ListCmdMutateFn: func(cmd *cobra.Command) {
			cmd.Flags().StringP("id", "", "", "image id to filter for")
			cmd.Flags().StringP("os", "", "", "image os to filter for")
			cmd.Flags().StringP("version", "", "", "image version to filter for")
			cmd.Flags().StringP("name", "", "", "image name to filter for")
			cmd.Flags().StringP("description", "", "", "image description to filter for")
			cmd.Flags().StringP("feature", "", "", "image feature to filter for, can be either machine|firewall")
		},
	}

	latestCmd := &cobra.Command{
		Use:   "latest",
		Short: "find latest image of one kind",
		RunE: func(cmd *cobra.Command, args []string) error {
			return w.latest()
		},
	}

	latestCmd.Flags().StringP("os", "", "", "find latest image for this os")

	return genericcli.NewCmds(cmdsConfig, latestCmd)
}

func (c *image) Get(id string) (*apiv2.Image, error) {
	ctx, cancel := c.c.NewRequestContext()
	defer cancel()

	req := &apiv2.ImageServiceGetRequest{Id: id}

	resp, err := c.c.V2Client.Apiv2().Image().Get(ctx, connect.NewRequest(req))
	if err != nil {
		return nil, fmt.Errorf("failed to get image: %w", err)
	}

	return resp.Msg.Image, nil
}

func (c *image) List() ([]*apiv2.Image, error) {
	ctx, cancel := c.c.NewRequestContext()
	defer cancel()

	req := &apiv2.ImageServiceListRequest{Query: &apiv2.ImageQuery{
		Id:          pointer.PointerOrNil(viper.GetString("id")),
		Os:          pointer.PointerOrNil(viper.GetString("os")),
		Version:     pointer.PointerOrNil(viper.GetString("version")),
		Name:        pointer.PointerOrNil(viper.GetString("name")),
		Description: pointer.PointerOrNil(viper.GetString("description")),
		Feature:     imageFeatureFromString(viper.GetString("feature")),
	}}

	resp, err := c.c.V2Client.Apiv2().Image().List(ctx, connect.NewRequest(req))
	if err != nil {
		return nil, fmt.Errorf("failed to get images: %w", err)
	}

	return resp.Msg.Images, nil
}

func (c *image) latest() error {
	ctx, cancel := c.c.NewRequestContext()
	defer cancel()

	req := &apiv2.ImageServiceLatestRequest{Os: viper.GetString("os")}

	resp, err := c.c.V2Client.Apiv2().Image().Latest(ctx, connect.NewRequest(req))
	if err != nil {
		return fmt.Errorf("failed to get images: %w", err)
	}

	return c.c.ListPrinter.Print(resp.Msg.Image)
}

func imageFeatureFromString(feature string) *apiv2.ImageFeature {
	if feature == "" {
		return nil
	}

	switch strings.ToLower(feature) {
	case "machine":
		return apiv2.ImageFeature_IMAGE_FEATURE_MACHINE.Enum()
	case "firewall":
		return apiv2.ImageFeature_IMAGE_FEATURE_FIREWALL.Enum()
	}
	return apiv2.ImageFeature_IMAGE_FEATURE_UNSPECIFIED.Enum()
}

func (c *image) Create(rq any) (*apiv2.Image, error) {
	panic("unimplemented")
}

func (c *image) Delete(id string) (*apiv2.Image, error) {
	panic("unimplemented")
}

func (t *image) Convert(r *apiv2.Image) (string, any, any, error) {
	panic("unimplemented")
}

func (t *image) Update(rq any) (*apiv2.Image, error) {
	panic("unimplemented")
}
