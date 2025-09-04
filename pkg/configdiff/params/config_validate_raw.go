package params

import (
	"github.com/sdcio/sdc-lite/pkg/types"
)

type ConfigValidateRaw struct{}

func NewConfigValidateRaw() *ConfigValidateRaw {
	return &ConfigValidateRaw{}
}

func (c *ConfigValidateRaw) GetMethod() types.CommandType {
	return types.CommandTypeConfigValidate
}

func (c *ConfigValidateRaw) UnRaw() (RunCommand, error) {
	return NewConfigValidate(), nil
}
