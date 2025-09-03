package params

import (
	"context"

	"github.com/sdcio/sdc-lite/cmd/interfaces"
	"github.com/sdcio/sdc-lite/pkg/configdiff/output"
	"github.com/sdcio/sdc-lite/pkg/types"
)

type SchemaLoadConfig struct {
	Schema []byte `json:"schema,omitempty" yaml:"schema,omitempty"`
}

func NewSchemaLoadConfig() *SchemaLoadConfig {
	return &SchemaLoadConfig{}
}

func (s *SchemaLoadConfig) SetSchema(schema []byte) *SchemaLoadConfig {
	s.Schema = schema
	return s
}

func (s *SchemaLoadConfig) GetSchema() []byte {
	return s.Schema
}

func (s *SchemaLoadConfig) String() string {
	return types.CommandTypeSchemaLoad
}

func (s *SchemaLoadConfig) Run(ctx context.Context, cde Executor) (interfaces.Output, error) {
	schema, err := cde.SchemaDownload(ctx, s)
	return output.NewSchemaOutput(schema), err
}
