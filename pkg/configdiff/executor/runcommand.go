package executor

import (
	"context"

	"github.com/sdcio/sdc-lite/cmd/interfaces"
	"github.com/sdcio/sdc-lite/pkg/configdiff/params"
)

type RunCommand interface {
	Run(ctx context.Context, cde params.Executor) (interfaces.Output, error)
	String() string
}
