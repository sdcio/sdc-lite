package types

import (
	"fmt"

	dsTypes "github.com/sdcio/data-server/pkg/tree/types"
	"github.com/sdcio/sdc-lite/pkg/configdiff/output"
)

type Intent struct {
	Name     string
	Prio     int32
	Flag     *dsTypes.UpdateInsertFlags
	BasePath string // basepath for the config data (no filepath)
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

func (i *Intent) GetName() string {
	return i.Name
}

func (i *Intent) GetPrio() int32 {
	return i.Prio
}

func (i *Intent) GetFlag() *dsTypes.UpdateInsertFlags {
	return i.Flag
}

func (i *Intent) GetData() []byte {
	return i.Data
}

func (i *Intent) GetBasePath() string {
	return i.BasePath
}

func (i *Intent) SetBasePath(p string) *Intent {
	i.BasePath = p
	return i
}

func (i *Intent) GetFormat() ConfigFormat {
	return i.Format
}

func (i *Intent) SetPrio(p int32) *Intent {
	i.Prio = p
	return i
}

func (i *Intent) SetData(format ConfigFormat, data []byte) *Intent {
	i.Format = format
	i.Data = data
	return i
}

func (i *Intent) String() string {
	return fmt.Sprintf("Name: %s, Prio: %d, Flag: %s, Format: %s", i.GetName(), i.GetPrio(), i.GetFlag(), i.GetFormat())
}

func (i *Intent) Export() *output.IntentOutput {

	return &output.IntentOutput{
		Name:     i.GetName(),
		Priority: i.GetPrio(),
		// Flags:    NewFlagsOutput(i.GetFlag()),
	}
}

type Intents map[string]*Intent

func (i *Intents) AddIntent(ii *Intent) {
	if *i == nil {
		*i = make(Intents)
	}
	(*i)[ii.GetName()] = ii
}

func (i Intents) Export() []*output.IntentOutput {
	result := make([]*output.IntentOutput, 0, len(i))
	for _, intent := range i {
		result = append(result, intent.Export())
	}
	return result
}
