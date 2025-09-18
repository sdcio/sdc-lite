package output

import (
	"encoding/json"
	"io"
	"strings"

	"github.com/sdcio/sdc-lite/cmd/interfaces"
)

type ConfigShowJsonIetfOutput struct {
	tree             TreeToJsonIetf
	onlyNewOrUpdated bool
}

var _ interfaces.Output = (*ConfigShowJsonIetfOutput)(nil)

func NewConfigShowJsonIetfOutput(root TreeToJsonIetf) *ConfigShowJsonIetfOutput {
	return &ConfigShowJsonIetfOutput{
		tree: root,
	}
}

func (c *ConfigShowJsonIetfOutput) SetOnlyNewOrUpdated(v bool) {
	c.onlyNewOrUpdated = v
}

func (o *ConfigShowJsonIetfOutput) ToString() (string, error) {
	sb := &strings.Builder{}
	err := o.WriteToJson(sb)
	if err != nil {
		return "", err
	}
	return sb.String(), nil
}
func (o *ConfigShowJsonIetfOutput) ToStringDetails() (string, error) {
	return o.ToString()
}
func (o *ConfigShowJsonIetfOutput) ToStruct() (any, error) {
	return o.tree.ToJsonIETF(o.onlyNewOrUpdated)
}
func (o *ConfigShowJsonIetfOutput) WriteToJson(w io.Writer) error {
	jenc := json.NewEncoder(w)
	jVal, err := o.ToStruct()
	if err != nil {
		return err
	}
	return jenc.Encode(jVal)
}

type TreeToJsonIetf interface {
	ToJsonIETF(onlyNewOrUpdated bool) (any, error)
}
