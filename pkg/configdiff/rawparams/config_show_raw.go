package rawparams

import (
	"github.com/sdcio/sdc-lite/pkg/configdiff/command_registry"
	"github.com/sdcio/sdc-lite/pkg/configdiff/executor"
	"github.com/sdcio/sdc-lite/pkg/configdiff/params"
	"github.com/sdcio/sdc-lite/pkg/configdiff/rpc"
	"github.com/sdcio/sdc-lite/pkg/types"
	sdcpb "github.com/sdcio/sdc-protos/sdcpb"
)

type ConfigShowConfigRaw struct {
	All          bool   `json:"all" yaml:"all"`
	Path         string `json:"path" yaml:"path"`
	OutputFormat string `json:"format" yaml:"format"`
}

func NewConfigShowConfigRaw() *ConfigShowConfigRaw {
	result := &ConfigShowConfigRaw{}
	return result
}

func (c *ConfigShowConfigRaw) GetMethod() types.CommandType {
	return types.CommandTypeConfigShow
}

func (c *ConfigShowConfigRaw) SetAll(b bool) *ConfigShowConfigRaw {
	c.All = b
	return c
}

func (c *ConfigShowConfigRaw) SetPath(p string) *ConfigShowConfigRaw {
	c.Path = p
	return c
}

func (c *ConfigShowConfigRaw) SetOutputFormat(f string) *ConfigShowConfigRaw {
	c.OutputFormat = f
	return c
}

func (c *ConfigShowConfigRaw) UnRaw() (executor.RunCommand, error) {
	p, err := sdcpb.ParsePath(c.Path)
	if err != nil {
		return nil, err
	}

	f, err := types.ParseConfigFormat(c.OutputFormat)
	if err != nil {
		return nil, err
	}

	result := params.NewConfigShowConfig().SetPath(p).SetAll(c.All).SetOutputFormat(f)

	return result, nil
}

func init() {
	command_registry.GetCommandRegistry().Register(types.CommandTypeConfigShow, func() rpc.RpcRawParams { return NewConfigShowConfigRaw() })
}
