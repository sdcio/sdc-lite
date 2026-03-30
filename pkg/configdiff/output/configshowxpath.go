package output

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"slices"
	"strings"

	"github.com/sdcio/data-server/pkg/tree/api"
	"github.com/sdcio/data-server/pkg/utils"
	"github.com/sdcio/sdc-lite/cmd/interfaces"
)

type ConfigShowXPathOutput struct {
	data TreeToXPath
}

var _ interfaces.Output = (*ConfigShowXPathOutput)(nil)

func NewConfigShowXPathOutput(data TreeToXPath) *ConfigShowXPathOutput {
	return &ConfigShowXPathOutput{
		data: data,
	}
}

func (o *ConfigShowXPathOutput) ToString(ctx context.Context) (string, error) {
	sb := &strings.Builder{}
	mapData := make(map[string]string)

	for _, v := range o.data.GetHighestPrecedence(ctx, false, false, false) {
		path := v.Update.SdcpbPath().ToXPath(false)
		value := v.Update.Value().ToString()
		mapData[path] = value
	}

	keys := make([]string, 0, len(mapData))
	for k := range mapData {
		keys = append(keys, k)
	}
	slices.Sort(keys)

	for _, k := range keys {
		fmt.Fprintf(sb, "%s -> %s\n", k, mapData[k])
	}
	return sb.String(), nil
}
func (o *ConfigShowXPathOutput) ToStringDetails(ctx context.Context) (string, error) {
	return o.ToString(ctx)
}
func (o *ConfigShowXPathOutput) ToStruct(ctx context.Context) (any, error) {
	mapData := make(map[string]any)
	for _, v := range o.data.GetHighestPrecedence(ctx, false, false, false) {
		path := v.Update.SdcpbPath().ToXPath(false)
		value, err := utils.GetJsonValue(v.Update.Value(), false)
		if err != nil {
			return nil, err
		}
		mapData[path] = value
	}
	return mapData, nil
}
func (o *ConfigShowXPathOutput) WriteToJson(ctx context.Context, w io.Writer) error {
	jEnc := json.NewEncoder(w)
	jEnc.SetIndent("", "  ")

	jVal, err := o.ToStruct(ctx)
	if err != nil {
		return err
	}
	return jEnc.Encode(jVal)
}

type TreeToXPath interface {
	GetHighestPrecedence(ctx context.Context, onlyNewOrUpdated bool, includeDefaults bool, includeExplicitDelete bool) api.LeafVariantSlice
}
