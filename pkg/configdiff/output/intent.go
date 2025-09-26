package output

import (
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

func (i *IntentOutput) ToString() (string, error) {
	return fmt.Sprintf("Name: %s, Priority: %d", i.Name, i.Priority), nil
}

func (i *IntentOutput) ToStringDetails() (string, error) {
	return i.ToString()
}

func (i *IntentOutput) WriteToJson(w io.Writer) error {
	jEnc := json.NewEncoder(w)
	jEnc.SetIndent("", "  ")
	return jEnc.Encode(i)
}

func (i *IntentOutput) ToStruct() (any, error) {
	return i, nil
}
