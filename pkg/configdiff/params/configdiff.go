package params

import (
	"context"

	"github.com/sdcio/sdc-lite/cmd/interfaces"
	"github.com/sdcio/sdc-lite/pkg/configdiff/enum"
	"github.com/sdcio/sdc-lite/pkg/configdiff/output"
	"github.com/sdcio/sdc-lite/pkg/types"
	sdcpb "github.com/sdcio/sdc-protos/sdcpb"

	"golang.org/x/term"
)

type DiffConfig struct {
	diffType     enum.DiffType
	format       types.ConfigFormat
	contextLines int
	width        int
	color        bool
	showHeader   bool
	path         *sdcpb.Path
}

func NewDiffConfig(dt enum.DiffType) *DiffConfig {
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

func (d *DiffConfig) GetDiffType() enum.DiffType {
	return d.diffType
}

func (d *DiffConfig) GetContextLines() int {
	return d.contextLines
}

func (d *DiffConfig) GetShowHeader() bool {
	return d.showHeader
}

func (d *DiffConfig) Run(ctx context.Context, cde Executor) (interfaces.Output, error) {
	diff, err := cde.GetDiff(ctx, d)
	return output.NewConfigDiffOutput(diff), err
}

func (d *DiffConfig) String() string {
	return types.CommandTypeConfigDiff
}
