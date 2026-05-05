package output

import (
	"context"
	"encoding/json"
	"io"

	"github.com/sdcio/sdc-lite/cmd/interfaces"
)

type NullOutput struct {
}

var _ interfaces.Output = (*NullOutput)(nil)

func NewNullOutput() *NullOutput {
	return &NullOutput{}
}

func (c *NullOutput) ToString(_ context.Context) (string, error) {
	return "", nil
}

func (c *NullOutput) ToStringDetails(_ context.Context) (string, error) {
	return "", nil
}

func (c *NullOutput) ToStruct(_ context.Context) (any, error) {
	return c, nil
}

func (c *NullOutput) WriteToJson(ctx context.Context, w io.Writer) error {
	jenc := json.NewEncoder(w)

	jVal, err := c.ToStruct(ctx)
	if err != nil {
		return err
	}
	return jenc.Encode(jVal)
}
