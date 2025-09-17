package params

import (
	"context"

	"github.com/sdcio/sdc-lite/cmd/interfaces"
	"github.com/sdcio/sdc-lite/pkg/configdiff/output"
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
	err := cde.TreeLoadData(ctx, c)
	if err != nil {
		return nil, err
	}

	result := output.NewErrorOutput(err)

	return result, nil
}
