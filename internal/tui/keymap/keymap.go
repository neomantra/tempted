package keymap

import (
	"github.com/charmbracelet/bubbles/key"
)

type keyMap struct {
	Back    key.Binding
	Exec    key.Binding
	Exit    key.Binding
	Filter  key.Binding
	Forward key.Binding
	Reload  key.Binding
	Task    key.Binding
	Term    key.Binding
	Wrap    key.Binding
}

var KeyMap = keyMap{
	Back: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "back"),
	),
	Exec: key.NewBinding(
		key.WithKeys("e"),
		key.WithHelp("e", "exec"),
	),
	Exit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q/ctrl+c", "exit"),
	),
	Filter: key.NewBinding(
		key.WithKeys("/"),
		key.WithHelp("/", "filter"),
	),
	Forward: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "enter"),
	),
	Reload: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "reload"),
	),
	Term: key.NewBinding(
		key.WithKeys("t"),
		key.WithHelp("t", "term"),
	),
	Wrap: key.NewBinding(
		key.WithKeys("ctrl+w"),
		key.WithHelp("ctrl+w", "toggle wrap"),
	),
}
