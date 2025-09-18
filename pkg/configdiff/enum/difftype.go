package enum

import (
	"fmt"
	"strings"
)

type DiffType string

const (
	DiffTypeUndefined       DiffType = "undefined"
	DiffTypeFull            DiffType = "full"
	DiffTypePatch           DiffType = "patch"
	DiffTypeSideBySide      DiffType = "side-by-side"
	DiffTypeSideBySidePatch DiffType = "side-by-side-patch"
)

var DiffTypeList = DiffTypes{DiffTypeFull, DiffTypePatch, DiffTypeSideBySide, DiffTypeSideBySidePatch}

func ParseDiffType(s string) (DiffType, error) {
	for _, x := range DiffTypeList {
		if strings.EqualFold(string(x), s) {
			return x, nil
		}
	}
	return DiffTypeUndefined, fmt.Errorf("unknown diff format: %s", s)
}

type DiffTypes []DiffType

func (d DiffTypes) StringSlice() []string {
	result := make([]string, 0, len(d))
	for _, x := range d {
		result = append(result, string(x))
	}
	return result
}
