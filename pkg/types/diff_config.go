package types

import (
	"fmt"
	"strings"

	"golang.org/x/term"
)

type DiffConfig struct {
	diffType     DiffType
	format       ConfigFormat
	contextLines int
	width        int
	color        bool
	showHeader   bool
}

func NewDiffConfig(dt DiffType) *DiffConfig {
	return &DiffConfig{
		diffType:     dt,
		color:        true,
		showHeader:   true,
		contextLines: 2,
	}
}

func (d *DiffConfig) GetWidth() int {
	if d.width == 0 {
		width, _, err := term.GetSize(0)
		if err != nil {
			d.width = 160
		}
		d.width = width
	}
	return d.width
}

func (d *DiffConfig) GetColor() bool {
	return d.color
}

func (d *DiffConfig) GetDiffType() DiffType {
	return d.diffType
}

func (d *DiffConfig) GetContextLines() int {
	return d.contextLines
}

func (d *DiffConfig) SetWidth(w int) *DiffConfig {
	d.width = w
	return d
}

func (d *DiffConfig) GetShowHeader() bool {
	return d.showHeader
}

func (d *DiffConfig) SetShowHeader(b bool) *DiffConfig {
	d.showHeader = b
	return d
}

func (d *DiffConfig) SetContextLines(l int) *DiffConfig {
	d.contextLines = l
	return d
}

func (d *DiffConfig) SetColor(b bool) *DiffConfig {
	d.color = b
	return d
}

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
