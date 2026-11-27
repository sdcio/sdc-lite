package output

import (
	"context"
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
	stats  *types.ValidationStats
}

var _ interfaces.Output = (*ConfigValidateOutput)(nil)

func NewConfigValidateOutput(result types.ValidationResults, stats *types.ValidationStats) *ConfigValidateOutput {
	return &ConfigValidateOutput{
		result: result,
		stats:  stats,
	}
}

func (cvo *ConfigValidateOutput) ToString(ctx context.Context) (string, error) {
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

func (cvo *ConfigValidateOutput) ToStringDetails(ctx context.Context) (string, error) {
	sb := &strings.Builder{}

	toString, err := cvo.ToString(ctx)
	if err != nil {
		return "", err
	}
	fmt.Fprint(sb, toString)

	fmt.Fprintln(sb, "Validations performed:")
	// sort the map, by getting the keys first
	keys := make([]types.StatType, 0, len(cvo.stats.GetCounter()))
	for typ := range cvo.stats.GetCounter() {
		keys = append(keys, typ)
	}

	// sorting the keys
	sort.Slice(keys, func(i, j int) bool {
		return keys[i].String() < keys[j].String()
	})

	indent := "  "
	// printing the stats in the sorted order
	for _, typ := range keys {
		fmt.Fprintf(sb, "%s%s: %d\n", indent, typ, cvo.stats.GetCounter()[typ])
	}
	return sb.String(), nil
}

func (cvo *ConfigValidateOutput) WriteToJson(ctx context.Context, w io.Writer) error {
	jenc := json.NewEncoder(w)

	jVal, err := cvo.ToStruct(ctx)
	if err != nil {
		return err
	}
	return jenc.Encode(jVal)
}

// ValidateResult holds structured YANG validation findings and rule-check counters.
type ValidateResult struct {
	Errors   []string          `json:"errors"`
	Warnings []string          `json:"warnings"`
	Stats    map[string]uint32 `json:"stats"`
}

func (cvo *ConfigValidateOutput) ToStruct(ctx context.Context) (any, error) {
	errs := cvo.result.ErrorsStr()
	if errs == nil {
		errs = []string{}
	}
	warns := cvo.result.WarningsStr()
	if warns == nil {
		warns = []string{}
	}
	result := ValidateResult{
		Errors:   errs,
		Warnings: warns,
		Stats:    map[string]uint32{},
	}
	if cvo.stats != nil {
		for k, v := range cvo.stats.Counter {
			if v != nil {
				result.Stats[k.String()] = *v
			}
		}
	}

	return result, nil
}
