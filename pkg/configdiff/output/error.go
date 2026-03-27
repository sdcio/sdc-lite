package output

import (
	"context"
	"encoding/json"
	"io"

	"github.com/sdcio/sdc-lite/cmd/interfaces"
)

type ErrorOutput struct {
	Error error
}

var _ interfaces.Output = (*ErrorOutput)(nil)

func NewErrorOutput(err error) *ErrorOutput {
	return &ErrorOutput{Error: err}
}

func (c *ErrorOutput) ToString(_ context.Context) (string, error) {
	return c.Error.Error(), nil
}

func (c *ErrorOutput) ToStringDetails(_ context.Context) (string, error) {
	return c.Error.Error(), nil
}

func (c *ErrorOutput) ToStruct(_ context.Context) (any, error) {
	return c, nil
}

func (c *ErrorOutput) WriteToJson(ctx context.Context, w io.Writer) error {
	jenc := json.NewEncoder(w)

	jVal, err := c.ToStruct(ctx)
	if err != nil {
		return err
	}
	return jenc.Encode(jVal)
}
