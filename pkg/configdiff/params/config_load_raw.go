package params

import (
	"fmt"

	"github.com/sdcio/sdc-lite/pkg/types"
	"github.com/sdcio/sdc-lite/pkg/utils"
)

type ConfigLoadRaw struct {
	Name     string                `json:"name" yaml:"name"`
	Prio     int32                 `json:"prio" yaml:"prio"`
	Flags    *UpdateInsertFlagsRaw `json:"flags" yaml:"flags"`
	Format   string                `json:"format" yaml:"format"`
	BasePath string                `json:"base-path" yaml:"base-path"`
	// either File or Data
	File string `json:"file,omitempty" yaml:"file,omitempty"`
	Data []byte `json:"data,omitempty" yaml:"data,omitempty"`
}

func NewConfigLoadRaw() *ConfigLoadRaw {
	result := &ConfigLoadRaw{}
	return result
}

func (i *ConfigLoadRaw) SetName(name string) *ConfigLoadRaw {
	i.Name = name
	return i
}

func (i *ConfigLoadRaw) SetPrio(prio int32) *ConfigLoadRaw {
	i.Prio = prio
	return i
}

func (i *ConfigLoadRaw) SetFlags(flags *UpdateInsertFlagsRaw) *ConfigLoadRaw {
	i.Flags = flags
	return i
}

func (i *ConfigLoadRaw) SetFormat(format string) *ConfigLoadRaw {
	i.Format = format
	return i
}

func (i *ConfigLoadRaw) SetFile(file string) *ConfigLoadRaw {
	i.File = file
	return i
}

func (i *ConfigLoadRaw) SetData(data []byte) *ConfigLoadRaw {
	i.Data = data
	return i
}

func (i *ConfigLoadRaw) SetBasePath(bp string) *ConfigLoadRaw {
	i.BasePath = bp
	return i
}

func (i *ConfigLoadRaw) GetName() string {
	return i.Name
}

func (i *ConfigLoadRaw) GetPrio() int32 {
	return i.Prio
}

func (i *ConfigLoadRaw) GetFlags() *UpdateInsertFlagsRaw {
	return i.Flags
}

func (i *ConfigLoadRaw) GetFormat() string {
	return i.Format
}

func (i *ConfigLoadRaw) GetFile() string {
	return i.File
}

func (i *ConfigLoadRaw) GetData() []byte {
	return i.Data
}

func (i *ConfigLoadRaw) GetMethod() types.CommandType {
	return types.CommandTypeConfigLoad
}

func (i *ConfigLoadRaw) GetBasePath() string {
	return i.BasePath
}

func (i *ConfigLoadRaw) UnRaw() (*ConfigLoad, error) {
	cf, err := types.ParseConfigFormat(i.Format)
	if err != nil {
		return nil, err
	}

	var data []byte

	switch {
	case len(i.Data) > 0:
		data = i.Data
	case i.File != "":
		fw := utils.NewFileWrapper(i.File)
		data, err = fw.Bytes()
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("no data provided")
	}

	intent := types.NewIntent(i.Name, i.Prio, i.Flags.UnRaw()).SetData(cf, data)

	return NewConfigLoad(intent), nil
}
