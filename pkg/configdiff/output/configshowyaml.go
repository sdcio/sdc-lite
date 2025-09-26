package output

import (
	"io"
	"strings"

	"github.com/sdcio/sdc-lite/cmd/interfaces"
	yaml "sigs.k8s.io/yaml/goyaml.v2"
)

type ConfigShowYamlOutput struct {
	tree             TreeToJson
	onlyNewOrUpdated bool
}

var _ interfaces.Output = (*ConfigShowYamlOutput)(nil)

func NewConfigShowYamlOutput(root TreeToJson) *ConfigShowYamlOutput {
	return &ConfigShowYamlOutput{
		tree: root,
	}
}

func (c *ConfigShowYamlOutput) SetOnlyNewOrUpdated(v bool) {
	c.onlyNewOrUpdated = v
}

func (o *ConfigShowYamlOutput) ToString() (string, error) {
	sb := &strings.Builder{}
	err := o.WriteToJson(sb)
	if err != nil {
		return "", err
	}
	return sb.String(), nil
}
func (o *ConfigShowYamlOutput) ToStringDetails() (string, error) {
	return o.ToString()
}
func (o *ConfigShowYamlOutput) ToStruct() (any, error) {
	return o.tree.ToJson(o.onlyNewOrUpdated)
}
func (o *ConfigShowYamlOutput) WriteToJson(w io.Writer) error {
	yEnc := yaml.NewEncoder(w)

	jVal, err := o.ToStruct()
	if err != nil {
		return err
	}
	return yEnc.Encode(jVal)
}
