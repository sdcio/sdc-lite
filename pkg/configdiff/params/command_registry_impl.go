package params

import (
	"fmt"

	"github.com/sdcio/sdc-lite/pkg/types"
)

type CommandRegistryImpl struct {
	init map[types.CommandType]CommandFactory
}

func newCommandRegistryImpl() *CommandRegistryImpl {
	return &CommandRegistryImpl{
		init: map[types.CommandType]CommandFactory{},
	}
}

func (c CommandRegistryImpl) Register(ct types.CommandType, cFactory CommandFactory) {
	c.init[ct] = cFactory
}

func (c *CommandRegistryImpl) GetCommandFactory(ct types.CommandType) (CommandFactory, error) {
	cf, ok := c.init[ct]
	if !ok {
		return nil, fmt.Errorf("command factory not registered: %s", ct)
	}
	return cf, nil
}
