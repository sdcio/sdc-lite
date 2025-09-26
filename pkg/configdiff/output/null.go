package output

import (
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

func (c *NullOutput) ToString() (string, error) {
	return "", nil
}

func (c *NullOutput) ToStringDetails() (string, error) {
	return "", nil
}

func (c *NullOutput) ToStruct() (any, error) {
	return c, nil
}

func (c *NullOutput) WriteToJson(w io.Writer) error {
	jenc := json.NewEncoder(w)

	jVal, err := c.ToStruct()
	if err != nil {
		return err
	}
	return jenc.Encode(jVal)
}
