package output

import (
	"context"
	"encoding/json"
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

func (c *ConfigShowYamlOutput) SetOnlyNewOrUpdated(v bool) *ConfigShowYamlOutput {
	c.onlyNewOrUpdated = v
	return c
}

func (o *ConfigShowYamlOutput) ToString(ctx context.Context) (string, error) {
	jVal, err := o.tree.ToJson(ctx, o.onlyNewOrUpdated)
	if err != nil {
		return "", err
	}
	sb := &strings.Builder{}
	if err := yaml.NewEncoder(sb).Encode(jVal); err != nil {
		return "", err
	}
	return sb.String(), nil
}
func (o *ConfigShowYamlOutput) ToStringDetails(ctx context.Context) (string, error) {
	return o.ToString(ctx)
}
func (o *ConfigShowYamlOutput) ToStruct(ctx context.Context) (any, error) {
	return o.ToString(ctx)
}
func (o *ConfigShowYamlOutput) WriteToJson(ctx context.Context, w io.Writer) error {
	jVal, err := o.ToStruct(ctx)
	if err != nil {
		return err
	}
	return json.NewEncoder(w).Encode(jVal)
}
