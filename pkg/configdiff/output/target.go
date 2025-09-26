package output

import (
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

func (t *TargetOutput) ToString() (string, error) {
	schemaString, err := t.Schema.ToString()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s [ %s ]\n", t.TargetName, schemaString), nil
}

func (t *TargetOutput) ToStringDetails() (string, error) {
	indent := "  "
	sb := &strings.Builder{}
	fmt.Fprintf(sb, "Target: %s ( %s )\n", t.TargetName, t.TargetPath)
	if t.Schema != nil {
		fmt.Fprintf(sb, "%sSchema:\n", indent)
		schemaString, err := t.Schema.ToStringDetails()
		if err != nil {
			return "", err
		}
		fmt.Fprintf(sb, "%[1]s%[1]s%[2]s\n", indent, schemaString)
	}
	fmt.Fprintf(sb, "%sIntents:\n", indent)
	for _, i := range t.Intents {
		intentString, err := i.ToStringDetails()
		if err != nil {
			return "", err
		}
		fmt.Fprintf(sb, "%[1]s%[1]s%[2]s\n", indent, intentString)
	}
	return sb.String(), nil
}

func (t *TargetOutput) WriteToJson(w io.Writer) error {
	jEnc := json.NewEncoder(w)
	jEnc.SetIndent("", "  ")
	return jEnc.Encode(t)
}

func (t *TargetOutput) ToStruct() (any, error) {
	return t, nil
}

type TargetOutputSlice []*TargetOutput

var _ interfaces.Output = (TargetOutputSlice)(nil)

func (t TargetOutputSlice) ToString() (string, error) {
	sb := &strings.Builder{}
	for _, target := range t {
		targetString, err := target.ToString()
		if err != nil {
			return "", err
		}
		fmt.Fprint(sb, targetString)
	}

	return sb.String(), nil
}

func (t TargetOutputSlice) ToStringDetails() (string, error) {
	sb := &strings.Builder{}
	for _, target := range t {
		targetString, err := target.ToStringDetails()
		if err != nil {
			return "", err
		}
		fmt.Fprint(sb, targetString)
	}

	return sb.String(), nil
}

func (t TargetOutputSlice) WriteToJson(w io.Writer) error {
	jEnc := json.NewEncoder(w)
	jEnc.SetIndent("", "  ")
	return jEnc.Encode(t)
}

func (t TargetOutputSlice) ToStruct() (any, error) {
	return t, nil
}
