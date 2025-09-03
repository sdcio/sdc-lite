package params

import (
	"context"

	"github.com/sdcio/sdc-lite/cmd/interfaces"
	"github.com/sdcio/sdc-lite/pkg/types"
)

type ConfigLoad struct {
	intent *types.Intent
}

func NewConfigLoad(intent *types.Intent) *ConfigLoad {
	return &ConfigLoad{intent: intent}
}

func (cl *ConfigLoad) GetIntent() *types.Intent {
	return cl.intent
}

func (c *ConfigLoad) String() string {
	return types.CommandTypeConfigLoad
}

func (c *ConfigLoad) Run(ctx context.Context, cde Executor) (interfaces.Output, error) {
	return nil, cde.TreeLoadData(ctx, c)
}
