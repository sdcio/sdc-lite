package cmd

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

func makeCompletionCmd(root *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:       "completion [bash|zsh|fish|powershell]",
		Short:     "Generate completion script",
		ValidArgs: []string{"bash", "zsh", "fish", "powershell"},
		RunE: func(cmd *cobra.Command, args []string) error {
			switch args[0] {
			case "bash":
				return EmitBashCompletionWithWordbreakMods(root, os.Stdout)
			case "zsh":
				return root.GenZshCompletion(os.Stdout)
			case "fish":
				return root.GenFishCompletion(os.Stdout, true)
			case "powershell":
				return root.GenPowerShellCompletion(os.Stdout)
			default:
				return fmt.Errorf("unsupported shell: %s", args[0])
			}
		},
	}
	return cmd
}

// EmitBashCompletionWithWordbreakMods writes a bash completion script for root
// to out, but replaces the registration so a wrapper is used which temporarily
// modifies COMP_WORDBREAKS.
//
//   - `adds` is a slice of single-character strings to append to COMP_WORDBREAKS
//     if not already present (e.g. []string{"/", ":"}).
//   - `removes` is a slice of single-character strings to remove from
//     COMP_WORDBREAKS while completing (e.g. []string{"=", "[", "]", "/"}).
//
// The wrapper restores the original COMP_WORDBREAKS after the underlying
// completion function returns.
func EmitBashCompletionWithWordbreakMods(root *cobra.Command, out io.Writer) error {

	comp_breakwords_adds := []string{"/", "[", "]", "="}
	comp_breakwords_removes := []string{}

	var buf bytes.Buffer
	if err := root.GenBashCompletion(&buf); err != nil {
		return err
	}
	script := buf.String()

	cmdName := root.Name()
	origFunc := fmt.Sprintf("__start_%s", cmdName)
	wrapperFunc := fmt.Sprintf("_%s_wrapper", cmdName)

	// Build wrapper body
	var w bytes.Buffer
	fmt.Fprintf(&w, "\n# Wrapper to temporarily adjust COMP_WORDBREAKS for %s\n", cmdName)
	fmt.Fprintf(&w, "%s() {\n", wrapperFunc)
	fmt.Fprintf(&w, "    local __old_COMP_WORDBREAKS=\"$COMP_WORDBREAKS\"\n\n")

	// Removals: produce lines like COMP_WORDBREAKS=${COMP_WORDBREAKS//\[/}
	for _, ch := range comp_breakwords_removes {
		if len(ch) == 0 {
			continue
		}
		esc := bashEscapePatternChar(ch)
		fmt.Fprintf(&w, "    COMP_WORDBREAKS=${COMP_WORDBREAKS//%s/}\n", esc)
	}
	if len(comp_breakwords_removes) > 0 {
		fmt.Fprintln(&w)
	}

	// Adds: add the char if not present
	// Use pattern matching: if [[ \"$COMP_WORDBREAKS\" != *"<ch>"* ]]; then COMP_WORDBREAKS="$COMP_WORDBREAKS<ch>"; fi
	for _, ch := range comp_breakwords_adds {
		if len(ch) == 0 {
			continue
		}
		// Put the literal char inside double quotes in pattern; safe because we produce literal file content.
		fmt.Fprintf(&w, "    if [[ \"$COMP_WORDBREAKS\" != *\"%s\"* ]]; then\n", ch)
		// append char
		// But if char is special for bash (e.g. backslash), it's okay inside double-quotes for literal.
		fmt.Fprintf(&w, "        COMP_WORDBREAKS=\"$COMP_WORDBREAKS%s\"\n", ch)
		fmt.Fprintln(&w, "    fi")
	}
	if len(comp_breakwords_adds) > 0 {
		fmt.Fprintln(&w)
	}

	// Call the original function, then restore
	fmt.Fprintf(&w, "    # call original cobra-generated completion function\n")
	fmt.Fprintf(&w, "    %s \"$@\"\n\n", origFunc)
	fmt.Fprintf(&w, "    # restore original COMP_WORDBREAKS\n")
	fmt.Fprintf(&w, "    COMP_WORDBREAKS=\"$__old_COMP_WORDBREAKS\"\n")
	fmt.Fprintf(&w, "}\n")

	// Replace the registration line to point to the wrapper
	// Common pattern in cobra: "complete -F _mycmd mycmd"
	origCompleteLine := fmt.Sprintf("%s %s", origFunc, cmdName)
	newCompleteLine := fmt.Sprintf("%s %s", wrapperFunc, cmdName)

	if strings.Contains(script, origCompleteLine) {
		script = strings.ReplaceAll(script, origCompleteLine, newCompleteLine)
	} else {
		// lenient fallback
		script = strings.ReplaceAll(script, fmt.Sprintf("complete -F %s", origFunc), fmt.Sprintf("complete -F %s", wrapperFunc))
	}

	script = strings.ReplaceAll(script,
		"_init_completion -s",
		"_init_completion -n ':=[]/' -s")
	script = strings.ReplaceAll(script,
		"__config-diff_init_completion -n \"=\" || return",
		"__config-diff_init_completion -n ':=[]/' || return")

	// Write script then wrapper
	if _, err := io.WriteString(out, script); err != nil {
		return err
	}
	if _, err := io.WriteString(out, w.String()); err != nil {
		return err
	}
	return nil
}

// bashEscapePatternChar escapes characters that need backslash in ${VAR//PAT/} patterns.
// We only escape the minimal set commonly problematic in this context.
func bashEscapePatternChar(ch string) string {
	// only single char expected - but handle longer gracefully by escaping every char
	var b strings.Builder
	for _, r := range ch {
		switch r {
		case '[', ']', '\\', '/', '$', '`':
			b.WriteRune('\\')
			b.WriteRune(r)
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}
