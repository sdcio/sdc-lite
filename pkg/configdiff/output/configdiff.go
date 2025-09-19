package output

import (
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

func (o *ConfigDiffOutput) ToString() (string, error) {
	return o.diff, nil
}
func (o *ConfigDiffOutput) ToStringDetails() (string, error) {
	return o.ToString()
}
func (o *ConfigDiffOutput) ToStruct() (any, error) {
	return struct{ Diff string }{Diff: o.diff}, nil
}
func (o *ConfigDiffOutput) WriteToJson(w io.Writer) error {
	jenc := json.NewEncoder(w)
	jVal, err := o.ToStruct()
	if err != nil {
		return err
	}
	return jenc.Encode(jVal)
}
