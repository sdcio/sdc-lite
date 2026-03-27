package output

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/sdcio/sdc-lite/cmd/interfaces"
)

type TargetOutput struct {
	TargetName string          `json:"target"`
	TargetPath string          `json:"target-path"`
	Schema     *SchemaOutput   `json:"schema"`
	Intents    []*IntentOutput `json:"intents"`
}

var _ interfaces.Output = (*TargetOutput)(nil)

func (t *TargetOutput) ToString(ctx context.Context) (string, error) {
	schemaString, err := t.Schema.ToString(ctx)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s [ %s ]\n", t.TargetName, schemaString), nil
}

func (t *TargetOutput) ToStringDetails(ctx context.Context) (string, error) {
	indent := "  "
	sb := &strings.Builder{}
	fmt.Fprintf(sb, "Target: %s ( %s )\n", t.TargetName, t.TargetPath)
	if t.Schema != nil {
		fmt.Fprintf(sb, "%sSchema:\n", indent)
		schemaString, err := t.Schema.ToStringDetails(ctx)
		if err != nil {
			return "", err
		}
		fmt.Fprintf(sb, "%[1]s%[1]s%[2]s\n", indent, schemaString)
	}
	fmt.Fprintf(sb, "%sIntents:\n", indent)
	for _, i := range t.Intents {
		intentString, err := i.ToStringDetails(ctx)
		if err != nil {
			return "", err
		}
		fmt.Fprintf(sb, "%[1]s%[1]s%[2]s\n", indent, intentString)
	}
	return sb.String(), nil
}

func (t *TargetOutput) WriteToJson(_ context.Context, w io.Writer) error {
	jEnc := json.NewEncoder(w)
	jEnc.SetIndent("", "  ")
	return jEnc.Encode(t)
}

func (t *TargetOutput) ToStruct(_ context.Context) (any, error) {
	return t, nil
}

type TargetOutputSlice []*TargetOutput

var _ interfaces.Output = (TargetOutputSlice)(nil)

func (t TargetOutputSlice) ToString(ctx context.Context) (string, error) {
	sb := &strings.Builder{}
	for _, target := range t {
		targetString, err := target.ToString(ctx)
		if err != nil {
			return "", err
		}
		fmt.Fprint(sb, targetString)
	}

	return sb.String(), nil
}

func (t TargetOutputSlice) ToStringDetails(ctx context.Context) (string, error) {
	sb := &strings.Builder{}
	for _, target := range t {
		targetString, err := target.ToStringDetails(ctx)
		if err != nil {
			return "", err
		}
		fmt.Fprint(sb, targetString)
	}

	return sb.String(), nil
}

func (t TargetOutputSlice) WriteToJson(_ context.Context, w io.Writer) error {
	jEnc := json.NewEncoder(w)
	jEnc.SetIndent("", "  ")
	return jEnc.Encode(t)
}

func (t TargetOutputSlice) ToStruct(_ context.Context) (any, error) {
	return t, nil
}
