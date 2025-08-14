package tui

import (
	"fmt"

	"go-mod-upgrade/internal/deps"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
)

// Implement concrete item rendering with deps.Module
type listModuleItem struct {
	Module   deps.Module
	Selected bool
}

func (i listModuleItem) Title() string {
	box := "[ ]"
	if i.Selected {
		box = "[x]"
	}

	cat := lipgloss.NewStyle().Foreground(lipgloss.Color("#04B575")).Render(i.Module.UpdateCategory)
	name := lipgloss.NewStyle().Bold(true).Render(i.Module.Path)

	return fmt.Sprintf("%s %s", box, name+" "+cat)
}

func (i listModuleItem) Description() string {
	from := lipgloss.NewStyle().Foreground(lipgloss.Color("#6C91C2")).Render("v" + i.Module.Current.String())
	to := lipgloss.NewStyle().Foreground(lipgloss.Color("#22C55E")).Render("v" + i.Module.Latest.String())
	return fmt.Sprintf("%s -> %s", from, to)
}

func (i listModuleItem) FilterValue() string { return i.Module.Path }

func newList(modules []deps.Module, keys listKeys) (list.Model, []listModuleItem) {
	d := list.NewDefaultDelegate()
	d.ShowDescription = true
	items := make([]list.Item, 0, len(modules))
	mitems := make([]listModuleItem, 0, len(modules))
	for _, m := range modules {
		mi := listModuleItem{Module: m}
		mitems = append(mitems, mi)
		items = append(items, mi)
	}

	l := list.New(items, d, 0, 0)
	l.SetShowStatusBar(true)
	l.SetShowTitle(true)
	l.Title = "go-mod-upgrade"
	l.SetFilteringEnabled(true)
	l.SetShowHelp(true)
	l.Help = help.New()
	l.AdditionalFullHelpKeys = func() []key.Binding { return []key.Binding{keys.toggle, keys.all, keys.apply, keys.quit} }
	l.SetShowPagination(true)

	return l, mitems
}
