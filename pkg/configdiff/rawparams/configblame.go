package rawparams

import (
	"github.com/sdcio/sdc-lite/pkg/configdiff/commandregistry"
	"github.com/sdcio/sdc-lite/pkg/configdiff/executor"
	"github.com/sdcio/sdc-lite/pkg/configdiff/params"
	"github.com/sdcio/sdc-lite/pkg/configdiff/rpc"
	"github.com/sdcio/sdc-lite/pkg/types"
	sdcpb "github.com/sdcio/sdc-protos/sdcpb"
)

type ConfigBlameParamsRaw struct {
	IncludeDefaults bool   `json:"include-defaults" yaml:"include-defaults"`
	Path            string `json:"path" yaml:"path"`
}

func NewConfigBlameParamsRaw() *ConfigBlameParamsRaw {
	return &ConfigBlameParamsRaw{}
}

func (c *ConfigBlameParamsRaw) GetMethod() types.CommandType {
	return types.CommandTypeConfigBlame
}

func (c *ConfigBlameParamsRaw) SetPath(p string) *ConfigBlameParamsRaw {
	c.Path = p
	return c
}

func (c *ConfigBlameParamsRaw) SetIncludeDefaults(includeDefaults bool) *ConfigBlameParamsRaw {
	c.IncludeDefaults = includeDefaults
	return c
}

func (c *ConfigBlameParamsRaw) UnRaw() (executor.RunCommand, error) {
	p, err := sdcpb.ParsePath(c.Path)
	if err != nil {
		return nil, err
	}

	result := params.NewConfigBlameParams().SetIncludeDefaults(c.IncludeDefaults).SetPath(p)
	return result, nil
}

func init() {
	commandregistry.GetCommandRegistry().Register(types.CommandTypeConfigBlame, func() rpc.RpcRawParams { return NewConfigBlameParamsRaw() })
}
