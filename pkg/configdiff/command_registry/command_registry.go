package command_registry

import (
	"github.com/sdcio/sdc-lite/pkg/configdiff/rpc"
	"github.com/sdcio/sdc-lite/pkg/types"
)

var registry CommandRegistry

func GetCommandRegistry() CommandRegistry {
	if registry == nil {
		registry = newCommandRegistryImpl()
	}
	return registry
}

type CommandFactory func() rpc.RpcRawParams

type CommandRegistry interface {
	Register(ct types.CommandType, cf CommandFactory)
	GetCommandFactory(ct types.CommandType) (CommandFactory, error)
}
