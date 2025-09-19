package params

import (
	"context"

	"github.com/sdcio/sdc-lite/cmd/interfaces"
	"github.com/sdcio/sdc-lite/pkg/configdiff/output"
	"github.com/sdcio/sdc-lite/pkg/types"
	sdcpb "github.com/sdcio/sdc-protos/sdcpb"
)

type ConfigBlameParams struct {
	includeDefaults bool
	path            *sdcpb.Path
}

func NewConfigBlameParams() *ConfigBlameParams {
	return &ConfigBlameParams{}
}

func (c *ConfigBlameParams) SetPath(p *sdcpb.Path) *ConfigBlameParams {
	c.path = p
	return c
}

func (c *ConfigBlameParams) SetIncludeDefaults(includeDefaults bool) *ConfigBlameParams {
	c.includeDefaults = includeDefaults
	return c
}

func (c *ConfigBlameParams) GetPath() *sdcpb.Path {
	return c.path
}

func (c *ConfigBlameParams) GetIncludeDefaults() bool {
	return c.includeDefaults
}

func (c *ConfigBlameParams) String() string {
	return types.CommandTypeConfigBlame
}

func (c *ConfigBlameParams) Run(ctx context.Context, cde Executor) (interfaces.Output, error) {
	tbe, err := cde.TreeBlame(ctx, c)
	if err != nil {
		return nil, err
	}
	return output.NewBlameResultOutput(tbe), nil
}
