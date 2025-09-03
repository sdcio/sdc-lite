package params

import (
	dsTypes "github.com/sdcio/data-server/pkg/tree/types"
)

type UpdateInsertFlagsRaw struct {
	New          bool `json:"new" yaml:"new"`
	Delete       bool `json:"delete" yaml:"delete"`
	OnlyIntended bool `json:"only_intended" yaml:"only_intended"`
}

func (u *UpdateInsertFlagsRaw) UnRaw() *dsTypes.UpdateInsertFlags {
	uif := dsTypes.NewUpdateInsertFlags()
	switch {
	case u.New:
		uif.SetNewFlag()
	case u.Delete:
		if u.OnlyIntended {
			uif.SetDeleteOnlyUpdatedFlag()
		} else {
			uif.SetDeleteFlag()
		}
	}
	return uif
}
