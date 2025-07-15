package types

import (
	"fmt"

	dsTypes "github.com/sdcio/data-server/pkg/tree/types"
)

type IntentInfo struct {
	Name   string
	Prio   int32
	Flag   *dsTypes.UpdateInsertFlags
	Format ConfigFormat
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

func (ii *IntentInfo) GetFormat() ConfigFormat {
	return ii.Format
}

func (ii *IntentInfo) SetData(format ConfigFormat, data []byte) *IntentInfo {
	ii.Format = format
	ii.Data = data
	return ii
}

func (ii *IntentInfo) String() string {
	return fmt.Sprintf("Name: %s, Prio: %d, Flag: %s, Format: %s", ii.GetName(), ii.GetPrio(), ii.GetFlag(), ii.GetFormat())
}

type IntentInfos map[string]*IntentInfo

func (i *IntentInfos) AddIntentInfo(ii *IntentInfo) {
	if *i == nil {
		*i = make(IntentInfos)
	}
	(*i)[ii.GetName()] = ii
}
