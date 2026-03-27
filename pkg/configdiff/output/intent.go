package output

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/sdcio/sdc-lite/cmd/interfaces"
)

type IntentOutput struct {
	Name     string `json:"target"`
	Priority int32  `json:"priority"`
	// Flags    *FlagsOutput `json:"flag"`
}

var _ interfaces.Output = (*IntentOutput)(nil)

func (i *IntentOutput) ToString(_ context.Context) (string, error) {
	return fmt.Sprintf("Name: %s, Priority: %d", i.Name, i.Priority), nil
}

func (i *IntentOutput) ToStringDetails(ctx context.Context) (string, error) {
	return i.ToString(ctx)
}

func (i *IntentOutput) WriteToJson(_ context.Context, w io.Writer) error {
	jEnc := json.NewEncoder(w)
	jEnc.SetIndent("", "  ")
	return jEnc.Encode(i)
}

func (i *IntentOutput) ToStruct(_ context.Context) (any, error) {
	return i, nil
}
