package workspace

import (
	"github.com/sdcio/config-diff/pkg/types"
	dsTypes "github.com/sdcio/data-server/pkg/tree/types"
)

type IntentInfo struct {
	Name   string
	Prio   int32
	Flag   *dsTypes.UpdateInsertFlags
	Format types.ConfigFormat
	Data   []byte
}

func NewIntentInfo(name string, prio int32, flag *dsTypes.UpdateInsertFlags) *IntentInfo {
	return &IntentInfo{
		Name: name,
		Prio: prio,
		Flag: flag,
	}
}

func (ii *IntentInfo) GetName() string {
	return ii.Name
}

func (ii *IntentInfo) GetPrio() int32 {
	return ii.Prio
}

func (ii *IntentInfo) GetFlag() *dsTypes.UpdateInsertFlags {
	return ii.Flag
}

func (ii *IntentInfo) GetData() []byte {
	return ii.Data
}

func (ii *IntentInfo) GetFormat() types.ConfigFormat {
	return ii.Format
}

func (ii *IntentInfo) SetData(format types.ConfigFormat, data []byte) *IntentInfo {
	ii.Format = format
	ii.Data = data
	return ii
}

type IntentInfos map[string]*IntentInfo
