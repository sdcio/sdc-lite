package params

import (
	"context"

	"github.com/sdcio/sdc-lite/cmd/interfaces"
	"github.com/sdcio/sdc-lite/pkg/types"
)

var registry CommandRegistry

func GetCommandRegistry() CommandRegistry {
	if registry == nil {
		registry = newCommandRegistryImpl()
	}
	return registry
}

type CommandFactory func() RpcRawParams

type CommandRegistry interface {
	Register(ct types.CommandType, cf CommandFactory)
	GetCommandFactory(ct types.CommandType) (CommandFactory, error)
}

type RunCommand interface {
	Run(ctx context.Context, cde Executor) (interfaces.Output, error)
	String() string
}
