package params

import (
	"context"
	"fmt"

	"github.com/sdcio/sdc-lite/cmd/interfaces"
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

	// TODO: dirty hack for now ... this must return a proper interface.Output
	out, err := cde.TreeGetString(ctx, c)
	if err != nil {
		return nil, err
	}
	fmt.Println(out)
	return nil, nil
}
