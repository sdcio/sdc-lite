package output

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/sdcio/data-server/pkg/tree/types"
	"github.com/sdcio/sdc-lite/cmd/interfaces"
)

type ConfigValidateOutput struct {
	result types.ValidationResults
	stats  *types.ValidationStatOverall
}

func NewConfigValidateOutput(result types.ValidationResults, stats *types.ValidationStatOverall) *ConfigValidateOutput {
	return &ConfigValidateOutput{
		result: result,
		stats:  stats,
	}
}

func (cvo *ConfigValidateOutput) ToString() (string, error) {
	sb := &strings.Builder{}

	if cvo.result.HasErrors() {
		fmt.Fprintf(sb, "Errors:\n%s", strings.Join(cvo.result.ErrorsStr(), "\n"))
	} else {
		fmt.Fprintln(sb, "Successfully validated!")
	}

	if cvo.result.HasWarnings() {
		fmt.Fprintf(sb, "Warnings:\n%s", strings.Join(cvo.result.WarningsStr(), "\n"))
	}

	return sb.String(), nil
}

func (cvo *ConfigValidateOutput) ToStringDetails() (string, error) {
	sb := &strings.Builder{}

	toString, err := cvo.ToString()
	if err != nil {
		return "", err
	}
	fmt.Fprint(sb, toString)

	fmt.Fprintln(sb, "Validations performed:")
	// sort the map, by getting the keys first
	keys := make([]string, 0, len(cvo.stats.GetCounter()))
	for typ := range cvo.stats.GetCounter() {
		keys = append(keys, typ)
	}

	// sorting the keys
	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})

	indent := "  "
	// printing the stats in the sorted order
	for _, typ := range keys {
		fmt.Fprintf(sb, "%s%s: %d\n", indent, typ, cvo.stats.GetCounter()[typ])
	}
	return sb.String(), nil
}

func (cvo *ConfigValidateOutput) WriteToJson(w io.Writer) error {
	jenc := json.NewEncoder(w)

	jVal, err := cvo.ToStruct()
	if err != nil {
		return err
	}
	return jenc.Encode(jVal)
}

func (cvo *ConfigValidateOutput) ToStruct() (any, error) {

	result := struct {
		Errors   []string
		Warnings []string
		Stats    *types.ValidationStatOverall
	}{
		Errors:   cvo.result.ErrorsStr(),
		Warnings: cvo.result.WarningsStr(),
		Stats:    cvo.stats,
	}

	return result, nil
}

var _ interfaces.Output = (*ConfigValidateOutput)(nil)
