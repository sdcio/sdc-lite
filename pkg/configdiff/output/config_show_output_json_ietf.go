package output

import (
	"encoding/json"
	"io"
	"strings"

	"github.com/sdcio/data-server/pkg/tree"
	"github.com/sdcio/sdc-lite/cmd/interfaces"
)

type ConfigShowJsonIetfOutput struct {
	root             tree.Entry
	onlyNewOrUpdated bool
}

var _ interfaces.Output = (*ConfigShowJsonIetfOutput)(nil)

func NewConfigShowJsonIetfOutput(root tree.Entry) *ConfigShowJsonIetfOutput {
	return &ConfigShowJsonIetfOutput{
		root: root,
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
	return o.root.ToJsonIETF(o.onlyNewOrUpdated)
}
func (o *ConfigShowJsonIetfOutput) WriteToJson(w io.Writer) error {
	jenc := json.NewEncoder(w)
	jVal, err := o.ToStruct()
	if err != nil {
		return err
	}
	return jenc.Encode(jVal)
}
