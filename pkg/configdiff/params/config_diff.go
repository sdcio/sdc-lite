package params

import (
	"context"
	"fmt"
	"strings"

	"github.com/sdcio/sdc-lite/cmd/interfaces"
	"github.com/sdcio/sdc-lite/pkg/types"
	sdcpb "github.com/sdcio/sdc-protos/sdcpb"

	"golang.org/x/term"
)

type DiffConfig struct {
	diffType     DiffType
	format       types.ConfigFormat
	contextLines int
	width        int
	color        bool
	showHeader   bool
	path         *sdcpb.Path
}

func NewDiffConfig(dt DiffType) *DiffConfig {
	return &DiffConfig{
		diffType:     dt,
		color:        true,
		showHeader:   true,
		contextLines: 2,
	}
}

func (d *DiffConfig) SetConfig(c types.ConfigFormat) *DiffConfig {
	d.format = c
	return d
}

func (d *DiffConfig) SetPath(p *sdcpb.Path) *DiffConfig {
	d.path = p
	return d
}

func (d *DiffConfig) SetWidth(w int) *DiffConfig {
	d.width = w
	return d
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

func (d *DiffConfig) GetPath() *sdcpb.Path {
	return d.path
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

func (d *DiffConfig) GetFormat() types.ConfigFormat {
	return d.format
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

func (d *DiffConfig) GetShowHeader() bool {
	return d.showHeader
}

func (d *DiffConfig) Run(ctx context.Context, cde Executor) (interfaces.Output, error) {
	// TODO: fix this
	diff, err := cde.GetDiff(ctx, d)
	fmt.Println(diff)
	return nil, err
}

func (d *DiffConfig) String() string {
	return types.CommandTypeConfigDiff
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
