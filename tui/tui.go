package tui

import (
	"errors"
	"os"
	"runtime"

	tea "github.com/charmbracelet/bubbletea"
	"golang.org/x/term"
)

// errCancelled is returned when the user aborts the form (Esc / ctrl+c).
var errCancelled = errors.New("cancelled")

// IsInteractive reports whether fscanx should launch the TUI instead of the
// classic CLI. Conditions:
//   - stdin is a terminal (TTY)
//   - not in -std pipe mode (no stdin redirection intended for masscan|fscanx)
//   - NOT on Windows: legacy Win7 conhost does not reliably render ANSI escape
//     sequences that bubbletea emits, so we keep the safe CLI path there.
//
// This preserves the Win7 / CI / pipe paths: when there is no capable terminal
// or input is piped, fscanx falls back to the original flag-driven CLI.
//
// force, when true, overrides the platform/pipe guard (used with explicit -tui).
func IsInteractive(force bool) bool {
	if force {
		return true
	}
	if runtime.GOOS == "windows" {
		return false
	}
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return false
	}
	// If the user already passed flags or a -std pipe, stay in CLI mode.
	for _, a := range os.Args[1:] {
		if a == "-std" || a == "--std" {
			return false
		}
	}
	return true
}

// RunTUI launches the interactive form. On success it returns an os.Args-style
// slice (without the program name) built from the user's input. The caller
// should set os.Args = append([]string{"fscanx"}, args...) and proceed with the
// normal common.Flag + Plugins.Scan flow.
//
// Returns errCancelled if the user aborts.
func RunTUI() ([]string, error) {
	m := newModel()
	p := tea.NewProgram(m)
	res, err := p.Run()
	if err != nil {
		return nil, err
	}
	final, ok := res.(model)
	if !ok {
		return nil, errors.New("unexpected tui result")
	}
	if final.err != nil {
		return nil, final.err
	}
	return final.BuildArgs(), nil
}
