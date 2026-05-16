package navigator

import "github.com/charmbracelet/bubbles/key"

type KeyMap struct {
	Bottom      key.Binding
	Top         key.Binding
	SectionDown key.Binding
	SectionUp   key.Binding
	Down        key.Binding
	Up          key.Binding
	Quit        key.Binding

	Search         key.Binding
	SearchNext     key.Binding
	SearchPrevious key.Binding
	SearchConfirm  key.Binding
	SearchQuit     key.Binding

	Copy             key.Binding
	Get              key.Binding
	Edit             key.Binding
	Delete           key.Binding
	RemoveFinalizers key.Binding
	Describe         key.Binding
	Pause            key.Binding
	Unpause          key.Binding
	Help             key.Binding
	CloseFullHelp    key.Binding
}

// DefaultKeyMap returns a default set of keybindings.
func DefaultKeyMap() KeyMap {
	return KeyMap{
		Bottom: key.NewBinding(
			key.WithKeys("bottom"),
			key.WithHelp("end", "bottom"),
		),
		Top: key.NewBinding(
			key.WithKeys("top"),
			key.WithHelp("home", "top"),
		),
		SectionDown: key.NewBinding(
			key.WithKeys("secdown"),
			key.WithHelp("secdown", "section down"),
		),
		SectionUp: key.NewBinding(
			key.WithKeys("secup"),
			key.WithHelp("secup", "section up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),

		Search: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "search"),
		),
		SearchNext: key.NewBinding(
			key.WithKeys("n"),
			key.WithHelp("n", "search next"),
		),
		SearchPrevious: key.NewBinding(
			key.WithKeys("N"),
			key.WithHelp("N", "search previous"),
		),
		SearchConfirm: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "search confirm"),
		),
		SearchQuit: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "search quit"),
		),

		Copy: key.NewBinding(
			key.WithKeys("c"),
			key.WithHelp("c", "copy"),
		),
		Get: key.NewBinding(
			key.WithKeys("y"),
			key.WithHelp("y", "get (yaml)"),
		),
		Edit: key.NewBinding(
			key.WithKeys("e"),
			key.WithHelp("e", "edit"),
		),
		Delete: key.NewBinding(
			key.WithKeys("ctrl+d"),
			key.WithHelp("ctrl+d", "delete"),
		),
		RemoveFinalizers: key.NewBinding(
			key.WithKeys("ctrl+x"),
			key.WithHelp("ctrl+x", "rm finalisers"),
		),
		Describe: key.NewBinding(
			key.WithKeys("enter", "d"),
			key.WithHelp("d", "describe")),
		Pause: key.NewBinding(
			key.WithKeys("p"),
			key.WithHelp("p", "pause")),
		Unpause: key.NewBinding(
			key.WithKeys("u"),
			key.WithHelp("u", "unpause")),
		Help: key.NewBinding(
			key.WithKeys("?", "h"),
			key.WithHelp("?/h", "toogle help"),
		),

		Quit: key.NewBinding(
			key.WithKeys("q"),
			key.WithHelp("q", "quit"),
		),
	}
}
