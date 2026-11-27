package executor

import (
	"context"

	"github.com/sdcio/sdc-lite/cmd/interfaces"
	"github.com/sdcio/sdc-lite/pkg/configdiff/params"
	"github.com/sdcio/sdc-lite/pkg/types"
)

type RunCommand interface {
	Run(ctx context.Context, cde params.Executor) (interfaces.Output, error)
	String() string
}

// FormattedCommand is implemented by commands whose output varies by a caller-selected ConfigFormat.
type FormattedCommand interface {
	ResultFormat() types.ConfigFormat
}
