package main

import (
	"context"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/brunoluiz/xpdig/internal/bubbles/action/kubectl"
	"github.com/brunoluiz/xpdig/internal/bubbles/action/shell"
	"github.com/brunoluiz/xpdig/internal/bubbles/app"
	"github.com/brunoluiz/xpdig/internal/bubbles/component/navigator"
	"github.com/brunoluiz/xpdig/internal/bubbles/component/statusbar"
	"github.com/brunoluiz/xpdig/internal/bubbles/component/table"
	"github.com/brunoluiz/xpdig/internal/bubbles/layout/xpnavigator"
	"github.com/brunoluiz/xpdig/internal/xplane"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/urfave/cli/v3"
)

// nolint: funlen
func cmdTrace() *cli.Command {
	return &cli.Command{
		Usage: `Explore tracing from Crossplane. Usage is available through arguments or data stream
1. To load it straight from a live resource using the crossplane CLI, do 'xpdig trace <object name>'
2. To load it from a trace JSON file, do 'crossplane resource trace -o json <> | xpdig trace --stdin'

Live mode is only available for (1) through the use of --watch / --watch-interval (see flag usage below)`,
		Name:    "trace",
		Aliases: []string{"t"},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "cmd",
				Usage: "Which binary should it use to generate the JSON trace",
				Value: "crossplane resource trace -o json",
			},
			&cli.StringFlag{Name: "context", Aliases: []string{"ctx"}, Usage: "Kubernetes context to be used"},
			&cli.StringFlag{Name: "namespace", Aliases: []string{"n", "ns"}, Usage: "Kubernetes namespace to be used"},
			&cli.BoolFlag{Name: "stdin", Aliases: []string{"in"}, Usage: "Specify in case file is piped into stdin"},
			&cli.BoolFlag{Name: "short", Usage: "Return short result columns for small screens"},
			&cli.BoolFlag{Name: "no-watch", Aliases: []string{"nw"}, Usage: "Disable automatic resource watcher", Value: false},
			&cli.DurationFlag{
				Name:    "watch-interval",
				Aliases: []string{"wi"},
				Usage:   "Refresh interval for the watcher feature",
				Value:   5 * time.Second,
			},
		},
		Action: runTrace,
	}
}

func runTrace(ctx context.Context, c *cli.Command) error {
	tracer, err := getTracer(c, logger.With("component", "tracer"))
	if err != nil {
		return err
	}

	logger.Info("starting xpdig",
		"component", "main",
		"info", map[string]any{
			"version": version,
			"args":    c.Args().Slice(),
			"flags":   getFlags(c),
		})

	watch := !c.Bool("no-watch")
	program := tea.NewProgram(
		app.New(
			logger.With("component", "bubbles/app"),
			kubectl.New(c.String("context"), shell.New(logger.With("component", "bubbles/action/shell"))),
			xpnavigator.New(
				logger.With("component", "bubbles/layout/xpnavigator"),
				navigator.New(
					logger.With("component", "bubbles/component/navigator"),
					table.New(
						table.WithFocused(true),
						table.WithStyles(func() table.Styles {
							s := table.DefaultStyles()
							s.Selected = lipgloss.NewStyle().
								Foreground(lipgloss.ANSIColor(ansi.Black)).
								Background(lipgloss.ANSIColor(ansi.White))
							return s
						}()),
					),
					textinput.New(),
				),
				statusbar.New(),
				tracer,
				xpnavigator.WithWatch(watch),
				xpnavigator.WithWatchInterval(c.Duration("watch-interval")),
				xpnavigator.WithShortColumns(c.Bool("short")),
			),
		),
		tea.WithAltScreen(),
		tea.WithContext(ctx),
	)

	_, err = program.Run()
	return err
}

type ErrInvalidArgument struct{}

func (e *ErrInvalidArgument) Error() string {
	return "trace for is not possible: argument must be on the format '<kind>/<name>' or '<kind> <name>'"
}

func getTracer(c *cli.Command, logger *slog.Logger) (xpnavigator.Tracer, error) {
	if c.Bool("stdin") {
		return xplane.NewReaderTraceQuerier(os.Stdin), nil
	}

	var kind, object string
	switch c.Args().Len() {
	case 1:
		n1 := c.Args().First()
		res := strings.Split(n1, "/")
		if len(res) != 2 {
			return nil, &ErrInvalidArgument{}
		}
		kind, object = res[0], res[1]
	case 2:
		kind, object = c.Args().Get(0), c.Args().Get(1)
	default:
		return nil, &ErrInvalidArgument{}
	}

	return xplane.NewCLITraceQuerier(
		logger,
		c.String("cmd"),
		c.String("namespace"),
		c.String("context"),
		kind, object,
	), nil
}
