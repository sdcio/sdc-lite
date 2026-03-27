package output

import (
	"context"
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

func (o *ConfigShowJsonIetfOutput) ToString(ctx context.Context) (string, error) {
	sb := &strings.Builder{}
	err := o.WriteToJson(ctx, sb)
	if err != nil {
		return "", err
	}
	return sb.String(), nil
}
func (o *ConfigShowJsonIetfOutput) ToStringDetails(ctx context.Context) (string, error) {
	return o.ToString(ctx)
}
func (o *ConfigShowJsonIetfOutput) ToStruct(ctx context.Context) (any, error) {
	return o.tree.ToJsonIETF(ctx, o.onlyNewOrUpdated)
}
func (o *ConfigShowJsonIetfOutput) WriteToJson(ctx context.Context, w io.Writer) error {
	jenc := json.NewEncoder(w)
	jenc.SetIndent("", "  ")
	jVal, err := o.ToStruct(ctx)
	if err != nil {
		return err
	}
	return jenc.Encode(jVal)
}

type TreeToJsonIetf interface {
	ToJsonIETF(ctx context.Context, onlyNewOrUpdated bool) (any, error)
}
