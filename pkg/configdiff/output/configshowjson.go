package output

import (
	"context"
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

func (o *ConfigShowJsonOutput) ToString(ctx context.Context) (string, error) {
	sb := &strings.Builder{}
	err := o.WriteToJson(ctx, sb)
	if err != nil {
		return "", err
	}
	return sb.String(), nil
}
func (o *ConfigShowJsonOutput) ToStringDetails(ctx context.Context) (string, error) {
	return o.ToString(ctx)
}
func (o *ConfigShowJsonOutput) ToStruct(ctx context.Context) (any, error) {
	return o.tree.ToJson(ctx, o.onlyNewOrUpdated)
}
func (o *ConfigShowJsonOutput) WriteToJson(ctx context.Context, w io.Writer) error {
	jenc := json.NewEncoder(w)
	jenc.SetIndent("", "  ")
	jVal, err := o.ToStruct(ctx)
	if err != nil {
		return err
	}
	return jenc.Encode(jVal)
}

type TreeToJson interface {
	ToJson(ctx context.Context, onlyNewOrUpdated bool) (any, error)
}
