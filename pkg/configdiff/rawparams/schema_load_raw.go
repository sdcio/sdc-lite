package rawparams

import (
	"strings"

	"github.com/sdcio/sdc-lite/pkg/configdiff/command_registry"
	"github.com/sdcio/sdc-lite/pkg/configdiff/executor"
	"github.com/sdcio/sdc-lite/pkg/configdiff/params"
	"github.com/sdcio/sdc-lite/pkg/configdiff/rpc"
	"github.com/sdcio/sdc-lite/pkg/types"
	"github.com/sdcio/sdc-lite/pkg/utils"
)

type SchemaLoadConfigRaw struct {
	File string `json:"file,omitempty" yaml:"file,omitempty"`
	Data []byte `json:"data,omitempty" yaml:"data,omitempty"`
}

func NewSchemaLoadConfigRaw() *SchemaLoadConfigRaw {
	result := &SchemaLoadConfigRaw{}
	return result
}

func (s *SchemaLoadConfigRaw) SetFile(file string) error {
	if strings.TrimSpace(file) == "-" {
		fw := utils.NewFileWrapper(file)

		schemaBytes, err := fw.Bytes()
		if err != nil {
			return err
		}
		s.SetData(schemaBytes)
		return nil
	}

	s.File = file
	return nil
}

func (s *SchemaLoadConfigRaw) SetData(data []byte) *SchemaLoadConfigRaw {
	s.Data = data
	return s
}

func (s *SchemaLoadConfigRaw) GetMethod() types.CommandType {
	return types.CommandTypeSchemaLoad
}

func (s *SchemaLoadConfigRaw) UnRaw() (executor.RunCommand, error) {
	var err error
	result := params.NewSchemaLoadConfig()
	data := s.Data

	// if data does not contain data, read the file reference
	if len(data) == 0 {
		fw := utils.NewFileWrapper(s.File)
		data, err = fw.Bytes()
		if err != nil {
			return nil, err
		}
	}
	// set the data
	result.SetSchema(data)
	return result, nil
}

func init() {
	command_registry.GetCommandRegistry().Register(types.CommandTypeSchemaLoad, func() rpc.RpcRawParams { return NewSchemaLoadConfigRaw() })
}
