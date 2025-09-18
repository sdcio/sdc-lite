package rawparams

import (
	"github.com/sdcio/sdc-lite/pkg/configdiff/command_registry"
	"github.com/sdcio/sdc-lite/pkg/configdiff/executor"
	"github.com/sdcio/sdc-lite/pkg/configdiff/params"
	"github.com/sdcio/sdc-lite/pkg/configdiff/rpc"
	"github.com/sdcio/sdc-lite/pkg/types"
)

type ConfigValidateRaw struct{}

func NewConfigValidateRaw() *ConfigValidateRaw {
	return &ConfigValidateRaw{}
}

func (c *ConfigValidateRaw) GetMethod() types.CommandType {
	return types.CommandTypeConfigValidate
}

func (c *ConfigValidateRaw) UnRaw() (executor.RunCommand, error) {
	return params.NewConfigValidate(), nil
}

func init() {
	command_registry.GetCommandRegistry().Register(types.CommandTypeConfigValidate, func() rpc.RpcRawParams { return NewConfigValidateRaw() })
}
