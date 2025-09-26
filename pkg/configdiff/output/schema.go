package output

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/sdcio/sdc-lite/cmd/interfaces"
	sdcpb "github.com/sdcio/sdc-protos/sdcpb"
)

type SchemaOutput struct {
	Vendor  string `json:"vendor"`
	Version string `json:"version"`
}

func NewSchemaOutput(s *sdcpb.Schema) *SchemaOutput {
	return &SchemaOutput{
		Vendor:  s.GetVendor(),
		Version: s.GetVersion(),
	}

}

func (s *SchemaOutput) WriteToJson(w io.Writer) error {
	jEnc := json.NewEncoder(w)
	jEnc.SetIndent("", "  ")
	return jEnc.Encode(s)
}

func (s *SchemaOutput) ToString() (string, error) {
	return fmt.Sprintf("Vendor: %s, Version: %s", s.Vendor, s.Version), nil
}

func (s *SchemaOutput) ToStringDetails() (string, error) {
	return s.ToString()
}

func (s *SchemaOutput) ToStruct() (any, error) {
	return s, nil
}

type SchemaOutputSlice []*SchemaOutput

var _ interfaces.Output = SchemaOutputSlice{}

func NewSchemaOutputList(ss []*sdcpb.Schema) SchemaOutputSlice {
	result := SchemaOutputSlice{}
	for _, s := range ss {
		result = append(result, NewSchemaOutput(s))
	}
	return result
}

func (s SchemaOutputSlice) ToString() (string, error) {
	sb := &strings.Builder{}
	for _, schema := range s {
		schemaString, err := schema.ToString()
		if err != nil {
			return "", err
		}
		fmt.Fprint(sb, schemaString)
	}
	return sb.String(), nil
}

func (s SchemaOutputSlice) ToStringDetails() (string, error) {
	return s.ToString()
}

func (s SchemaOutputSlice) WriteToJson(w io.Writer) error {
	jEnc := json.NewEncoder(w)
	jEnc.SetIndent("", "  ")
	return jEnc.Encode(s)
}

func (s SchemaOutputSlice) ToStruct() (any, error) {
	return s, nil
}
