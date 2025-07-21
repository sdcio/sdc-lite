package diff

import (
	"fmt"
	"math"
	"strings"
)

// ANSI color codes
const (
	colorReset = "\033[0m"
	colorRed   = "\033[31m"
	colorGreen = "\033[32m"
	colorGray  = "\033[90m"
)

type diffLine struct {
	typ  byte // ' ', '+', '-'
	line string
}

func Diff(oldLines, newLines []string) []diffLine {
	m, n := len(oldLines), len(newLines)
	lcs := make([][]int, m+1)
	for i := range lcs {
		lcs[i] = make([]int, n+1)
	}
	for i := 0; i < m; i++ {
		for j := 0; j < n; j++ {
			if oldLines[i] == newLines[j] {
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
		case i > 0 && j > 0 && oldLines[i-1] == newLines[j-1]:
			result = append([]diffLine{{' ', oldLines[i-1]}}, result...)
			i--
			j--
		case j > 0 && (i == 0 || lcs[i][j-1] >= lcs[i-1][j]):
			result = append([]diffLine{{'+', newLines[j-1]}}, result...)
			j--
		case i > 0:
			result = append([]diffLine{{'-', oldLines[i-1]}}, result...)
			i--
		}
	}
	return result
}

func colorLine(d diffLine) string {
	switch d.typ {
	case '+':
		return colorGreen + "+ " + d.line + colorReset
	case '-':
		return colorRed + "- " + d.line + colorReset
	default:
		return colorGray + "  " + d.line + colorReset
	}
}

func GenerateFullDiffString(oldLines, newLines []string, showHeader bool) string {
	diff := Diff(oldLines, newLines)
	var sb strings.Builder
	if showHeader {
		fmt.Fprintf(&sb, "@@ -1,%d +1,%d @@\n", len(oldLines), len(newLines))
	}
	for _, d := range diff {
		sb.WriteString(colorLine(d) + "\n")
	}
	return sb.String()
}

func GenerateContextDiffString(oldLines, newLines []string, context int, showHeader bool) string {
	diff := Diff(oldLines, newLines)
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
		start := max(changeIndices[i]-context, 0)
		end := min(changeIndices[i]+context+1, len(diff))

		for i+1 < len(changeIndices) && changeIndices[i+1]-end < context {
			end = min(changeIndices[i+1]+context+1, len(diff))
			i++
		}
		i++

		if showHeader {
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
				sb.WriteString(colorLine(diff[j]) + "\n")
				seen[j] = true
			}
		}
		sb.WriteString(colorGray + "---" + colorReset + "\n")
	}

	return sb.String()
}

func min(a, b int) int { return int(math.Min(float64(a), float64(b))) }
func max(a, b int) int { return int(math.Max(float64(a), float64(b))) }

func GenerateSideBySideDiffString(oldLines, newLines []string, width int, showColor bool) string {
	diff := Diff(oldLines, newLines)
	var sb strings.Builder

	leftPad := func(s string) string {
		if len(s) > width {
			return s[:width]
		}
		return fmt.Sprintf("%-*s", width, s)
	}

	colorize := func(prefix byte, line string) string {
		if !showColor {
			return fmt.Sprintf("%c %s", prefix, line)
		}
		switch prefix {
		case '+':
			return colorGreen + "+ " + line + colorReset
		case '-':
			return colorRed + "- " + line + colorReset
		default:
			return colorGray + "  " + line + colorReset
		}
	}

	i := 0
	for i < len(diff) {
		left := ""
		right := ""
		ltyp, rtyp := byte(' '), byte(' ')

		if i < len(diff) {
			d := diff[i]
			switch d.typ {
			case ' ':
				left, right = d.line, d.line
				ltyp, rtyp = ' ', ' '
				i++
			case '-':
				left = d.line
				ltyp = '-'
				if i+1 < len(diff) && diff[i+1].typ == '+' {
					right = diff[i+1].line
					rtyp = '+'
					i += 2
				} else {
					right = ""
					rtyp = ' '
					i++
				}
			case '+':
				right = d.line
				rtyp = '+'
				left = ""
				ltyp = ' '
				i++
			}
		}

		leftStr := colorize(ltyp, leftPad(left))
		rightStr := colorize(rtyp, right)
		sb.WriteString(fmt.Sprintf("%s | %s\n", leftStr, rightStr))
	}

	return sb.String()
}

func GenerateSideBySideContextDiffString(oldLines, newLines []string, context, width int, showHeader, showColor bool) string {
	diff := Diff(oldLines, newLines)
	var sb strings.Builder

	type lineState struct {
		idx   int
		left  string
		right string
		ltype byte
		rtype byte
	}

	var changeIndices []int
	for i, d := range diff {
		if d.typ == '+' || d.typ == '-' {
			changeIndices = append(changeIndices, i)
		}
	}
	if len(changeIndices) == 0 {
		return ""
	}

	leftPad := func(s string) string {
		if len(s) > width {
			return s[:width]
		}
		return fmt.Sprintf("%-*s", width, s)
	}
	colorize := func(prefix byte, text string) string {
		if !showColor {
			return fmt.Sprintf("%c %s", prefix, text)
		}
		switch prefix {
		case '+':
			return colorGreen + "+ " + text + colorReset
		case '-':
			return colorRed + "- " + text + colorReset
		default:
			return colorGray + "  " + text + colorReset
		}
	}

	i := 0
	for i < len(changeIndices) {
		start := max(changeIndices[i]-context, 0)
		end := min(changeIndices[i]+context+1, len(diff))

		// Extend to merge nearby blocks
		for i+1 < len(changeIndices) && changeIndices[i+1]-end < context {
			end = min(changeIndices[i+1]+context+1, len(diff))
			i++
		}
		i++

		// Header logic
		if showHeader {
			oldLine, newLine := 1, 1
			for j := 0; j < start; j++ {
				if diff[j].typ != '+' {
					oldLine++
				}
				if diff[j].typ != '-' {
					newLine++
				}
			}
			oldCount, newCount := 0, 0
			for j := start; j < end; j++ {
				if diff[j].typ != '+' {
					oldCount++
				}
				if diff[j].typ != '-' {
					newCount++
				}
			}
			fmt.Fprintf(&sb, colorGray+"@@ -%d,%d +%d,%d @@"+colorReset+"\n", oldLine, oldCount, newLine, newCount)
		}

		// Render diff block
		j := start
		for j < end {
			var left, right string
			ltyp, rtyp := byte(' '), byte(' ')
			if j < len(diff) {
				d := diff[j]
				switch d.typ {
				case ' ':
					left = d.line
					right = d.line
					ltyp, rtyp = ' ', ' '
					j++
				case '-':
					left = d.line
					ltyp = '-'
					if j+1 < len(diff) && diff[j+1].typ == '+' {
						right = diff[j+1].line
						rtyp = '+'
						j += 2
					} else {
						right = ""
						rtyp = ' '
						j++
					}
				case '+':
					right = d.line
					rtyp = '+'
					left = ""
					ltyp = ' '
					j++
				}
			}
			leftStr := colorize(ltyp, leftPad(left))
			rightStr := colorize(rtyp, right)
			sb.WriteString(fmt.Sprintf("%s | %s\n", leftStr, rightStr))
		}
		sb.WriteString(colorGray + "---" + colorReset + "\n")
	}

	return sb.String()
}
