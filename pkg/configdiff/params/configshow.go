package params

import (
	"context"
	"fmt"

	"github.com/sdcio/sdc-lite/cmd/interfaces"
	"github.com/sdcio/sdc-lite/pkg/configdiff/output"
	"github.com/sdcio/sdc-lite/pkg/types"
	sdcpb "github.com/sdcio/sdc-protos/sdcpb"
)

type ConfigShowConfig struct {
	all          bool
	path         *sdcpb.Path
	outputFormat types.ConfigFormat
}

func NewConfigShowConfig() *ConfigShowConfig {
	return &ConfigShowConfig{}
}

func (c *ConfigShowConfig) SetAll(b bool) *ConfigShowConfig {
	c.all = b
	return c
}

func (c *ConfigShowConfig) SetPath(p *sdcpb.Path) *ConfigShowConfig {
	c.path = p
	return c
}

func (c *ConfigShowConfig) SetOutputFormat(f types.ConfigFormat) *ConfigShowConfig {
	c.outputFormat = f
	return c
}

func (c *ConfigShowConfig) GetAll() bool {
	return c.all
}

func (c *ConfigShowConfig) GetPath() *sdcpb.Path {
	return c.path
}

func (c *ConfigShowConfig) GetOutputFormat() types.ConfigFormat {
	return c.outputFormat
}

func (c *ConfigShowConfig) String() string {
	return "show-config"
}

func (c *ConfigShowConfig) Run(ctx context.Context, cde Executor) (interfaces.Output, error) {
	entry, err := cde.TreeShow(ctx, c)
	if err != nil {
		return nil, err
	}
	switch c.GetOutputFormat() {
	case types.ConfigFormatXml:
		return output.NewConfigShowXmlOutput(entry), nil
	case types.ConfigFormatJson:
		return output.NewConfigShowJsonOutput(entry), nil
	case types.ConfigFormatJsonIetf:
		return output.NewConfigShowJsonIetfOutput(entry), nil
	case types.ConfigFormatYaml:
		return output.NewConfigShowYamlOutput(entry), nil
	case types.ConfigFormatXPath:
		// TODO
		// xpv := visitors.NewXPathVisitor()
		// err := entry.Walk(ctx, xpv)
		// if err != nil {
		// 	return nil, err
		// }
		// return xpv.GetResult(), nil
		return &output.NullOutput{}, nil
	case types.ConfigFormatSdc:
		fallthrough
	default:
		return nil, fmt.Errorf("output in %s format not supported", c.GetOutputFormat())
	}
}
