package params

import (
	"github.com/sdcio/sdc-lite/pkg/types"
	sdcpb "github.com/sdcio/sdc-protos/sdcpb"
)

type DiffConfigRaw struct {
	DiffType     string `json:"diff_type" yaml:"diff_type"`
	Format       string `json:"format" yaml:"format"`
	ContextLines int    `json:"context_lines" yaml:"context_lines"`
	Width        int    `json:"width" yaml:"width"`
	NoColor      bool   `json:"no_color" yaml:"no_color"`
	NoShowHeader bool   `json:"no_show_header" yaml:"no_show_header"`
	Path         string `json:"path" yaml:"path"`
}

func NewDiffConfigRaw() *DiffConfigRaw {
	result := &DiffConfigRaw{}
	return result
}

func (d *DiffConfigRaw) SetConfig(c string) *DiffConfigRaw {
	d.Format = c
	return d
}

func (d *DiffConfigRaw) SetDiffType(dt string) *DiffConfigRaw {
	d.DiffType = dt
	return d
}

func (d *DiffConfigRaw) SetPath(p string) *DiffConfigRaw {
	d.Path = p
	return d
}

func (d *DiffConfigRaw) SetWidth(w int) *DiffConfigRaw {
	d.Width = w
	return d
}

func (d *DiffConfigRaw) SetShowHeader(b bool) *DiffConfigRaw {
	d.NoShowHeader = !b
	return d
}

func (d *DiffConfigRaw) SetContextLines(l int) *DiffConfigRaw {
	d.ContextLines = l
	return d
}

func (d *DiffConfigRaw) SetNoColor(b bool) *DiffConfigRaw {
	d.NoColor = b
	return d
}

func (c *DiffConfigRaw) GetMethod() types.CommandType {
	return types.CommandTypeConfigDiff
}

func (d *DiffConfigRaw) UnRaw() (RunCommand, error) {
	var err error
	var dt DiffType = DiffTypePatch
	if d.DiffType != "" {
		dt, err = ParseDiffType(d.DiffType)
		if err != nil {
			return nil, err
		}
	}

	var f types.ConfigFormat = types.ConfigFormatJson
	if d.Format != "" {
		f, err = types.ParseConfigFormat(d.Format)
		if err != nil {
			return nil, err
		}
	}

	var path *sdcpb.Path
	if d.Path != "" {
		path, err = sdcpb.ParsePath(d.Path)
		if err != nil {
			return nil, err
		}
	}

	dc := NewDiffConfig(dt).SetColor(!d.NoColor).SetConfig(f).SetPath(path).SetShowHeader(!d.NoShowHeader)

	if dc.contextLines != 0 {
		dc.SetContextLines(d.ContextLines)
	}
	if d.Width != 0 {
		dc.SetWidth(d.Width)
	}

	return dc, nil
}
