package params

import (
	"context"

	"github.com/sdcio/sdc-lite/cmd/interfaces"
	"github.com/sdcio/sdc-lite/pkg/configdiff/output"
	"github.com/sdcio/sdc-lite/pkg/types"
)

type ConfigValidate struct {
}

func NewConfigValidate() *ConfigValidate {
	return &ConfigValidate{}
}

func (c *ConfigValidate) String() string {
	return types.CommandTypeConfigValidate
}

func (c *ConfigValidate) Run(ctx context.Context, cde Executor) (interfaces.Output, error) {
	result, stats, err := cde.TreeValidate(ctx)
	out := output.NewConfigValidateOutput(result, stats)
	return out, err
}
