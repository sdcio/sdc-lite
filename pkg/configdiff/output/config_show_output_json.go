package output

import (
	"encoding/json"
	"io"
	"strings"

	"github.com/sdcio/sdc-lite/cmd/interfaces"
)

type ConfigShowJsonOutput struct {
	tree             TreeToJson
	onlyNewOrUpdated bool
}

var _ interfaces.Output = (*ConfigShowJsonOutput)(nil)

func NewConfigShowJsonOutput(root TreeToJson) *ConfigShowJsonOutput {
	return &ConfigShowJsonOutput{
		tree: root,
	}
}

func (c *ConfigShowJsonOutput) SetOnlyNewOrUpdated(v bool) {
	c.onlyNewOrUpdated = v
}

func (o *ConfigShowJsonOutput) ToString() (string, error) {
	sb := &strings.Builder{}
	err := o.WriteToJson(sb)
	if err != nil {
		return "", err
	}
	return sb.String(), nil
}
func (o *ConfigShowJsonOutput) ToStringDetails() (string, error) {
	return o.ToString()
}
func (o *ConfigShowJsonOutput) ToStruct() (any, error) {
	return o.tree.ToJson(o.onlyNewOrUpdated)
}
func (o *ConfigShowJsonOutput) WriteToJson(w io.Writer) error {
	jenc := json.NewEncoder(w)
	jVal, err := o.ToStruct()
	if err != nil {
		return err
	}
	return jenc.Encode(jVal)
}

type TreeToJson interface {
	ToJson(onlyNewOrUpdated bool) (any, error)
}
