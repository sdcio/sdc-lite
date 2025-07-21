package types

import (
	"fmt"

	dsTypes "github.com/sdcio/data-server/pkg/tree/types"
)

type Intent struct {
	Name     string
	Prio     int32
	BasePath string
	Flag     *dsTypes.UpdateInsertFlags
	Format   ConfigFormat
	Data     []byte
}

func NewIntent(name string, prio int32, flag *dsTypes.UpdateInsertFlags) *Intent {
	return &Intent{
		Name: name,
		Prio: prio,
		Flag: flag,
	}
}

func (ii *Intent) GetName() string {
	return ii.Name
}

func (ii *Intent) GetPrio() int32 {
	return ii.Prio
}

func (ii *Intent) GetFlag() *dsTypes.UpdateInsertFlags {
	return ii.Flag
}

func (ii *Intent) GetData() []byte {
	return ii.Data
}

func (ii *Intent) GetBasePath() string {
	return ii.BasePath
}

func (ii *Intent) SetBasePath(p string) *Intent {
	ii.BasePath = p
	return ii
}

func (ii *Intent) GetFormat() ConfigFormat {
	return ii.Format
}

func (ii *Intent) SetData(format ConfigFormat, data []byte) *Intent {
	ii.Format = format
	ii.Data = data
	return ii
}

func (ii *Intent) String() string {
	return fmt.Sprintf("Name: %s, Prio: %d, Flag: %s, Format: %s", ii.GetName(), ii.GetPrio(), ii.GetFlag(), ii.GetFormat())
}

type Intents map[string]*Intent

func (i *Intents) AddIntent(ii *Intent) {
	if *i == nil {
		*i = make(Intents)
	}
	(*i)[ii.GetName()] = ii
}
