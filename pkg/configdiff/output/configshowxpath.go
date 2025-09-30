package output

import (
	"encoding/json"
	"fmt"
	"io"
	"slices"
	"strings"

	"github.com/sdcio/sdc-lite/cmd/interfaces"
)

type ConfigShowXPathOutput struct {
	data map[string]any
}

var _ interfaces.Output = (*ConfigShowXPathOutput)(nil)

func NewConfigShowXPath(data map[string]any) *ConfigShowXPathOutput {
	return &ConfigShowXPathOutput{
		data: data,
	}
}

func (o *ConfigShowXPathOutput) ToString() (string, error) {
	sb := &strings.Builder{}

	keys := make([]string, 0, len(o.data))
	for k := range o.data {
		keys = append(keys, k)
	}
	slices.Sort(keys)

	for _, k := range keys {
		fmt.Fprintf(sb, "%s -> %v\n", k, o.data[k])
	}
	return sb.String(), nil
}
func (o *ConfigShowXPathOutput) ToStringDetails() (string, error) {
	return o.ToString()
}
func (o *ConfigShowXPathOutput) ToStruct() (any, error) {
	return o.data, nil
}
func (o *ConfigShowXPathOutput) WriteToJson(w io.Writer) error {
	jEnc := json.NewEncoder(w)
	jEnc.SetIndent("", "  ")

	jVal, err := o.ToStruct()
	if err != nil {
		return err
	}
	return jEnc.Encode(jVal)
}
