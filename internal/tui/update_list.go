package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/chaindead/modup/internal/deps"
)

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

func (i listModuleItem) FilterValue() string {
	// Use path without the domain to allow matching owner/repo (and subpaths) but avoid noisy domain matches
	p := i.Module.Path
	if slash := strings.IndexByte(p, '/'); slash >= 0 && slash+1 < len(p) {
		return p[slash+1:]
	}
	return p
}

func listItemSelected(listItem list.Item) bool {
	return listItem.(listModuleItem).Selected
}

func listItemSetSelected(listItem list.Item, selected bool) list.Item {
	item := listItem.(listModuleItem)
	item.Selected = selected

	return item
}

func findItemIndexByPath(items []list.Item, modulePath string) int {
	for idx, it := range items {
		if lm, ok := it.(listModuleItem); ok {
			if lm.Module.Path == modulePath {
				return idx
			}
		}
	}
	return -1
}

type listKeyMap struct {
	toggleItem key.Binding
	toggleAll  key.Binding
	update     key.Binding
}

func newListKeyMap() *listKeyMap {
	return &listKeyMap{
		toggleItem: key.NewBinding(
			key.WithKeys(" "),
			key.WithHelp("space", "toggle item"),
		),
		toggleAll: key.NewBinding(
			key.WithKeys("a"),
			key.WithHelp("a", "toggle all"),
		),
		update: key.NewBinding(
			key.WithKeys("enter", "u"),
			key.WithHelp("enter", "update selected"),
		),
	}
}

func newItemDelegate(keys *listKeyMap) list.DefaultDelegate {
	d := list.NewDefaultDelegate()
	d.ShowDescription = true
	d.Styles.FilterMatch = lipgloss.NewStyle()

	d.UpdateFunc = func(msg tea.Msg, m *list.Model) tea.Cmd {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			if m.FilterState() == list.Filtering {
				return nil
			}

			switch {
			case key.Matches(msg, keys.toggleItem):
				idx := m.GlobalIndex()
				if idx >= 0 && idx < len(m.Items()) {
					item := m.Items()[idx]
					newItem := listItemSetSelected(item, !listItemSelected(item))
					setCmd := m.SetItem(idx, newItem)
					statusCmd := m.NewStatusMessage(statusMessageStyle("Selected " + newItem.(listModuleItem).Module.Path))
					return tea.Batch(setCmd, statusCmd)
				}

			case key.Matches(msg, keys.toggleAll):
				visible := m.VisibleItems()
				allVisibleSelected := true
				for _, it := range visible {
					if !listItemSelected(it) {
						allVisibleSelected = false
						break
					}
				}

				var cmds []tea.Cmd
				for _, vit := range visible {
					lm, ok := vit.(listModuleItem)
					if !ok {
						continue
					}
					underlyingIdx := findItemIndexByPath(m.Items(), lm.Module.Path)
					if underlyingIdx < 0 {
						continue
					}
					newItem := listItemSetSelected(m.Items()[underlyingIdx], !allVisibleSelected)
					cmds = append(cmds, m.SetItem(underlyingIdx, newItem))
				}

				action := "Selected all"
				if allVisibleSelected {
					action = "Deselected all"
				}
				statusCmd := m.NewStatusMessage(statusMessageStyle(action))
				cmds = append(cmds, statusCmd)
				return tea.Batch(cmds...)

			case key.Matches(msg, keys.update):
				selected := make([]deps.Module, 0)
				for _, it := range m.VisibleItems() {
					if listItemSelected(it) {
						lm := it.(listModuleItem)
						selected = append(selected, lm.Module)
					}
				}

				if len(selected) == 0 {
					return m.NewStatusMessage(statusMessageStyle("No packages selected to update"))
				}

				return beginUpgradeCmd(selected)
			}
		}

		return nil
	}

	help := []key.Binding{keys.toggleItem, keys.toggleAll, keys.update}

	d.ShortHelpFunc = func() []key.Binding {
		return help
	}

	d.FullHelpFunc = func() [][]key.Binding {
		return [][]key.Binding{help}
	}

	return d
}

func (m model) newList() (list.Model, []list.Item) {
	modules := m.modules

	keys := newListKeyMap()
	d := newItemDelegate(keys)

	items := make([]list.Item, 0, len(modules))
	for _, m := range modules {
		items = append(items, listModuleItem{Module: m})
	}

	l := list.New(items, d, 0, 0)
	l.Title = "Go Modules available to upgrade"
	l.Help = help.New()
	l.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{
			keys.toggleItem,
			keys.toggleAll,
			keys.update,
		}
	}

	l.SetStatusBarItemName("package", "packages")
	l.SetShowTitle(true)
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)
	l.SetShowPagination(true)

	h, v := appStyle.GetFrameSize()
	l.SetSize(m.width-h, m.height-v)

	return l, items
}

var statusMessageStyle = lipgloss.NewStyle().
	Foreground(lipgloss.AdaptiveColor{Light: "#04B575", Dark: "#04B575"}).
	Render

func (m model) listUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
	newListModel, cmd := m.list.Update(msg)
	m.list = newListModel

	m.items = m.list.Items()

	return m, cmd
}
