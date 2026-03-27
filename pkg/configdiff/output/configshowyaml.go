package output

import (
	"context"
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

func (o *ConfigShowYamlOutput) ToString(ctx context.Context) (string, error) {
	sb := &strings.Builder{}
	err := o.WriteToJson(ctx, sb)
	if err != nil {
		return "", err
	}
	return sb.String(), nil
}
func (o *ConfigShowYamlOutput) ToStringDetails(ctx context.Context) (string, error) {
	return o.ToString(ctx)
}
func (o *ConfigShowYamlOutput) ToStruct(ctx context.Context) (any, error) {
	return o.tree.ToJson(ctx, o.onlyNewOrUpdated)
}
func (o *ConfigShowYamlOutput) WriteToJson(ctx context.Context, w io.Writer) error {
	yEnc := yaml.NewEncoder(w)

	jVal, err := o.ToStruct(ctx)
	if err != nil {
		return err
	}
	return yEnc.Encode(jVal)
}
