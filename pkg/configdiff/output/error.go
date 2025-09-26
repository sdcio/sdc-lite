package output

import (
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

func (c *ErrorOutput) ToString() (string, error) {
	return c.Error.Error(), nil
}

func (c *ErrorOutput) ToStringDetails() (string, error) {
	return c.Error.Error(), nil
}

func (c *ErrorOutput) ToStruct() (any, error) {
	return c, nil
}

func (c *ErrorOutput) WriteToJson(w io.Writer) error {
	jenc := json.NewEncoder(w)

	jVal, err := c.ToStruct()
	if err != nil {
		return err
	}
	return jenc.Encode(jVal)
}
