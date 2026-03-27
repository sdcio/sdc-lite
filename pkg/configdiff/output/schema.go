package output

import (
	"context"
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

func (s *SchemaOutput) WriteToJson(ctx context.Context, w io.Writer) error {
	jEnc := json.NewEncoder(w)
	jEnc.SetIndent("", "  ")
	return jEnc.Encode(s)
}

func (s *SchemaOutput) ToString(_ context.Context) (string, error) {
	return fmt.Sprintf("Vendor: %s, Version: %s\n", s.Vendor, s.Version), nil
}

func (s *SchemaOutput) ToStringDetails(ctx context.Context) (string, error) {
	return s.ToString(ctx)
}

func (s *SchemaOutput) ToStruct(_ context.Context) (any, error) {
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

func (s SchemaOutputSlice) ToString(ctx context.Context) (string, error) {
	sb := &strings.Builder{}
	for _, schema := range s {
		schemaString, err := schema.ToString(ctx)
		if err != nil {
			return "", err
		}
		fmt.Fprint(sb, schemaString)
	}
	return sb.String(), nil
}

func (s SchemaOutputSlice) ToStringDetails(ctx context.Context) (string, error) {
	return s.ToString(ctx)
}

func (s SchemaOutputSlice) WriteToJson(ctx context.Context, w io.Writer) error {
	jEnc := json.NewEncoder(w)
	jEnc.SetIndent("", "  ")
	return jEnc.Encode(s)
}

func (s SchemaOutputSlice) ToStruct(_ context.Context) (any, error) {
	return s, nil
}
