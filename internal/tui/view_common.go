package tui

import (
	"github.com/charmbracelet/bubbles/key"
)

type listKeys struct {
	toggle key.Binding
	all    key.Binding
	apply  key.Binding
	quit   key.Binding
}

func newListKeys() listKeys {
	return listKeys{
		toggle: key.NewBinding(key.WithKeys(" "), key.WithHelp("space", "toggle")),
		all:    key.NewBinding(key.WithKeys("a"), key.WithHelp("a", "toggle all")),
		apply:  key.NewBinding(key.WithKeys("enter", "u"), key.WithHelp("enter", "update")),
		quit:   key.NewBinding(key.WithKeys("q", "esc", "ctrl+c"), key.WithHelp("q", "quit")),
	}
}
