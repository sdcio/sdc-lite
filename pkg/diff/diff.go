package diff

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"

	"github.com/sdcio/sdc-lite/pkg/configdiff/enum"
	"github.com/sdcio/sdc-lite/pkg/configdiff/params"
)

// ANSI color codes
const (
	colorReset               = "\033[0m"
	colorRed                 = "\033[31m"
	colorGreen               = "\033[32m"
	colorGray                = "\033[90m"
	sideBySideSeparatorWidth = 3
)

type diffLine struct {
	typ  byte // ' ', '+', '-'
	line string
}

type Differ struct {
	old    []string
	new    []string
	config *params.DiffConfig
}

func NewDiffer(old, new []string) *Differ {
	return &Differ{
		old: old,
		new: new,
	}
}

func (d *Differ) SetConfig(dc *params.DiffConfig) *Differ {
	d.config = dc
	return d
}

func NewDifferJson(old, new any) (*Differ, error) {
	oldByte, err := json.MarshalIndent(old, "", "  ")
	if err != nil {
		return nil, err
	}
	newByte, err := json.MarshalIndent(new, "", "  ")
	if err != nil {
		return nil, err
	}

	oldStr := string(oldByte)
	oldStrSl := strings.Split(oldStr, "\n")

	newStr := string(newByte)
	newStrSl := strings.Split(newStr, "\n")

	return NewDiffer(oldStrSl, newStrSl), nil
}

func (d *Differ) Diff() (string, error) {
	switch d.config.GetDiffType() {
	case enum.DiffTypeFull:
		return d.generateFullDiffString(), nil
	case enum.DiffTypePatch:
		return d.generateContextDiffString(), nil
	case enum.DiffTypeSideBySidePatch:
		return d.generateSideBySideContextDiffString(), nil
	default:
		// case types.DiffTypeSideBySide:
		return d.generateSideBySideDiffString(), nil
	}
}

func (d *Differ) calcDiff() []diffLine {
	m, n := len(d.old), len(d.new)
	lcs := make([][]int, m+1)
	for i := range lcs {
		lcs[i] = make([]int, n+1)
	}
	for i := 0; i < m; i++ {
		for j := 0; j < n; j++ {
			if d.old[i] == d.new[j] {
				lcs[i+1][j+1] = lcs[i][j] + 1
			} else {
				lcs[i+1][j+1] = max(lcs[i+1][j], lcs[i][j+1])
			}
		}
	}

	var result []diffLine
	i, j := m, n
	for i > 0 || j > 0 {
		switch {
		case i > 0 && j > 0 && d.old[i-1] == d.new[j-1]:
			result = append([]diffLine{{' ', d.old[i-1]}}, result...)
			i--
			j--
		case j > 0 && (i == 0 || lcs[i][j-1] >= lcs[i-1][j]):
			result = append([]diffLine{{'+', d.new[j-1]}}, result...)
			j--
		case i > 0:
			result = append([]diffLine{{'-', d.old[i-1]}}, result...)
			i--
		}
	}
	return result
}

func colorLine(dl diffLine, color bool) string {
	colorPlus := ""
	colorMinus := ""

	if color {
		colorPlus = colorGreen
		colorMinus = colorRed
	}
	switch dl.typ {
	case '+':
		return colorPlus + "+" + dl.line + colorReset
	case '-':
		return colorMinus + "-" + dl.line + colorReset
	default:
		return " " + dl.line
	}
}

func (d *Differ) generateFullDiffString() string {
	diffs := d.calcDiff()
	var sb strings.Builder
	if d.config.GetShowHeader() {
		fmt.Fprintf(&sb, "@@ -1,%d +1,%d @@\n", len(d.old), len(d.new))
	}
	for _, diff := range diffs {
		sb.WriteString(colorLine(diff, d.config.GetColor()) + "\n")
	}
	return sb.String()
}

func (d *Differ) generateContextDiffString() string {
	diff := d.calcDiff()
	var sb strings.Builder

	var changeIndices []int
	for i, d := range diff {
		if d.typ == '+' || d.typ == '-' {
			changeIndices = append(changeIndices, i)
		}
	}
	if len(changeIndices) == 0 {
		return ""
	}

	seen := make(map[int]bool)
	for i := 0; i < len(changeIndices); {
		start := max(changeIndices[i]-d.config.GetContextLines(), 0)
		end := min(changeIndices[i]+d.config.GetContextLines()+1, len(diff))

		for i+1 < len(changeIndices) && changeIndices[i+1]-end < d.config.GetContextLines() {
			end = min(changeIndices[i+1]+d.config.GetContextLines()+1, len(diff))
			i++
		}
		i++

		if d.config.GetShowHeader() {
			// Compute old and new line ranges for header
			oldStart, oldCount := 0, 0
			newStart, newCount := 0, 0
			oldIndex, newIndex := 1, 1 // unified diff is 1-based

			for j := 0; j < start; j++ {
				if diff[j].typ != '+' {
					oldIndex++
				}
				if diff[j].typ != '-' {
					newIndex++
				}
			}

			oldStart = oldIndex
			newStart = newIndex

			for j := start; j < end; j++ {
				if diff[j].typ != '+' {
					oldCount++
				}
				if diff[j].typ != '-' {
					newCount++
				}
			}

			fmt.Fprintf(&sb, colorGray+"@@ -%d,%d +%d,%d @@"+colorReset+"\n", oldStart, oldCount, newStart, newCount)
		}

		for j := start; j < end; j++ {
			if !seen[j] {
				sb.WriteString(colorLine(diff[j], d.config.GetColor()) + "\n")
				seen[j] = true
			}
		}
		sb.WriteString(colorGray + "---" + colorReset + "\n")
	}

	return sb.String()
}

func min(a, b int) int { return int(math.Min(float64(a), float64(b))) }
func max(a, b int) int { return int(math.Max(float64(a), float64(b))) }

func formatSideBySideLine(left, right string, width int) string {
	leftLines := wrapLine(left, width)
	rightLines := wrapLine(right, width)
	maxLines := max(len(leftLines), len(rightLines))

	var sb strings.Builder
	for i := 0; i < maxLines; i++ {
		var l, r string
		if i < len(leftLines) {
			l = leftLines[i]
		}
		if i < len(rightLines) {
			r = rightLines[i]
		}
		sb.WriteString(fmt.Sprintf("%-*s │ %s\n", width, l, r))
	}
	return sb.String()
}

func wrapLine(s string, width int) []string {
	runes := []rune(s)
	var lines []string
	for i := 0; i < len(runes); i += width {
		end := i + width
		if end > len(runes) {
			end = len(runes)
		}
		lines = append(lines, string(runes[i:end]))
	}
	return lines
}

func (d *Differ) generateSideBySideDiffString() string {
	diff := d.calcDiff()
	var sb strings.Builder

	width := (d.config.GetWidth() - sideBySideSeparatorWidth) / 2

	formatLine := func(left, right string) string {
		return formatSideBySideLine(left, right, width)
	}

	leftLine, rightLine := "", ""

	colorPlus := ""
	colorMinus := ""
	if d.config.GetColor() {
		colorPlus = colorGreen
		colorMinus = colorRed
	}

	for _, dl := range diff {
		switch dl.typ {
		case ' ':
			leftLine = " " + dl.line
			rightLine = " " + dl.line
		case '-':
			leftLine = colorMinus + "-" + dl.line + colorReset
			rightLine = ""
		case '+':
			leftLine = ""
			rightLine = colorPlus + "+" + dl.line + colorReset
		}
		sb.WriteString(formatLine(leftLine, rightLine))
	}

	return sb.String()
}

func (d *Differ) generateSideBySideContextDiffString() string {
	diff := d.calcDiff()
	var sb strings.Builder

	width := (d.config.GetWidth() - sideBySideSeparatorWidth) / 2
	formatLine := func(left, right string) string {
		return formatSideBySideLine(left, right, width)
	}

	var changeIndices []int
	for i, dl := range diff {
		if dl.typ != ' ' {
			changeIndices = append(changeIndices, i)
		}
	}
	if len(changeIndices) == 0 {
		return ""
	}

	seen := make(map[int]bool)
	for i := 0; i < len(changeIndices); {
		start := max(changeIndices[i]-d.config.GetContextLines(), 0)
		end := min(changeIndices[i]+d.config.GetContextLines()+1, len(diff))

		for i+1 < len(changeIndices) && changeIndices[i+1]-end < d.config.GetContextLines() {
			end = min(changeIndices[i+1]+d.config.GetContextLines()+1, len(diff))
			i++
		}
		i++

		if d.config.GetShowHeader() {
			sb.WriteString(colorGray + fmt.Sprintf("@@ Context lines %d-%d @@\n", start+1, end) + colorReset)
		}

		colorPlus := ""
		colorMinus := ""

		if d.config.GetColor() {
			colorMinus = colorRed
			colorPlus = colorGreen
		}

		for j := start; j < end; j++ {
			if seen[j] {
				continue
			}
			dl := diff[j]
			var leftLine, rightLine string
			switch dl.typ {
			case ' ':
				leftLine = " " + dl.line
				rightLine = " " + dl.line
			case '-':
				leftLine = colorMinus + "-" + dl.line + colorReset
				rightLine = ""
			case '+':
				leftLine = ""
				rightLine = colorPlus + "+" + dl.line + colorReset
			}
			sb.WriteString(formatLine(leftLine, rightLine))
			seen[j] = true
		}
		sb.WriteString(colorGray + "---\n" + colorReset)
	}

	return sb.String()
}
