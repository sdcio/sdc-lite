package output

import (
	"context"
	"encoding/json"
	"io"

	"github.com/sdcio/sdc-lite/cmd/interfaces"
)

type ConfigDiffOutput struct {
	diff string
}

var _ interfaces.Output = (*ConfigDiffOutput)(nil)

func NewConfigDiffOutput(s string) *ConfigDiffOutput {
	return &ConfigDiffOutput{
		diff: s,
	}
}

func (o *ConfigDiffOutput) ToString(_ context.Context) (string, error) {
	return o.diff, nil
}
func (o *ConfigDiffOutput) ToStringDetails(ctx context.Context) (string, error) {
	return o.ToString(ctx)
}
func (o *ConfigDiffOutput) ToStruct(_ context.Context) (any, error) {
	return struct{ Diff string }{Diff: o.diff}, nil
}
func (o *ConfigDiffOutput) WriteToJson(ctx context.Context, w io.Writer) error {
	jenc := json.NewEncoder(w)
	jVal, err := o.ToStruct(ctx)
	if err != nil {
		return err
	}
	return jenc.Encode(jVal)
}
