package shell

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type Cmd struct {
	logger *slog.Logger
}

func New(logger *slog.Logger) *Cmd {
	return &Cmd{logger: logger}
}

func (s *Cmd) Exec(c string, args ...string) tea.Cmd {
	s.logger.Info("executing shell", "cmd", c, "args", args)

	cmd := exec.Command(c, args...) //nolint:gosec // intentional shell execution utility
	// Inherit environment so $EDITOR is respected
	cmd.Env = os.Environ()
	// Attach to the user's terminal
	// cmd.Stdin = os.Stdin
	// cmd.Stdout = os.Stdout
	// cmd.Stderr = os.Stderr

	return tea.ExecProcess(cmd, func(_ error) tea.Msg {
		return nil
	})
}

func (s *Cmd) Pager(c string, args ...string) tea.Cmd {
	cmd := c + " " + strings.Join(args, " ")
	pager := os.Getenv("PAGER")
	// Default for those who never thought about it
	if pager == "" {
		pager = "less"
	}
	// If we don't do this, it will not render the output as YAML,
	// since stdin does not tell us much about the format
	if pager == "bat" {
		pager = "bat -l yaml --style=plain --paging always"
	}
	viewCmd := fmt.Sprintf("%s | %s", cmd, pager)

	return s.Exec(os.Getenv("SHELL"), "-c", viewCmd)
}
