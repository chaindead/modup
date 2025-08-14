package tui

import (
	"fmt"
	"go-mod-upgrade/internal/deps"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		if m.mode == modeList {
			m.list.SetSize(m.width, m.height)
		}
		return m, nil
	case errMsg:
		m.err = msg
		return m, tea.Quit
	case tea.KeyMsg:
		if m.mode == modeList {
			switch msg.String() {
			case "q", "esc", "ctrl+c":
				return m, tea.Quit
			case " ":
				idx := m.list.Index()
				if idx >= 0 && idx < len(m.items) {
					m.items[idx].Selected = !m.items[idx].Selected
					m.list.SetItem(idx, m.items[idx])
				}
			case "a":
				all := true
				for _, it := range m.items {
					if !it.Selected {
						all = false
						break
					}
				}
				for i := range m.items {
					m.items[i].Selected = !all
					m.list.SetItem(i, m.items[i])
				}
			case "enter", "u":
				var sel []deps.Module
				for _, it := range m.items {
					if it.Selected {
						sel = append(sel, it.Module)
					}
				}
				if len(sel) == 0 {
					break
				}
				m.mode = modeUpdate
				m.updated = 0
				m.total = len(sel)
				return m, m.startUpdate(sel)
			}
			var cmd tea.Cmd
			m.list, cmd = m.list.Update(msg)
			return m, cmd
		}
		return m, nil
	case scanTotalMsg:
		m.total = int(msg)
		m.scanProgCh = make(chan int, 32)
		m.scanModsCh = make(chan []deps.Module, 1)
		m.scanErrCh = make(chan error, 1)
		go func() {
			mods, err := deps.DiscoverOutdatedWithProgress(m.scanProgCh)
			if err != nil {
				m.scanErrCh <- err
			} else {
				m.scanModsCh <- mods
			}
			close(m.scanProgCh)
			close(m.scanModsCh)
			close(m.scanErrCh)
		}()
		return m, m.waitScanEventCmd()
	case scanProgressMsg:
		m.scanned += int(msg)
		return m, m.waitScanEventCmd()
	case scanDoneMsg:
		mods := []deps.Module(msg)
		if ok, _ := deps.ToolsSupported(); ok {
			if tools, err := deps.DiscoverToolUpdates(); err == nil && len(tools) > 0 {
				mods = append(mods, tools...)
			}
		}
		if m.opts.Force {
			m.mode = modeUpdate
			m.updated = 0
			m.total = len(mods)
			return m, m.startUpdate(mods)
		}
		l, items := newList(mods, m.opts.PageSize, m.keys)
		if m.width > 0 && m.height > 0 {
			l.SetSize(m.width, m.height)
		}
		m.list = l
		m.items = items
		m.mode = modeList
		return m, nil
	case updateProgressMsg:
		m.updated = int(msg)
		return m, m.waitUpdateEventCmd()
	case updatedModuleMsg:
		// increment progress and append a pretty line for the module
		m.updated++
		um := deps.Module(msg)
		line := m.renderUpdateLine(um)
		m.updatedLines = append(m.updatedLines, line)
		return m, m.waitUpdateEventCmd()
	case updateDoneMsg:
		m.mode = modeDone
		return m, tea.Quit
	}
	if m.mode == modeScan || m.mode == modeUpdate {
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m Model) View() string {
	switch m.mode {
	case modeScan:
		percent := 0.0
		if m.total > 0 {
			percent = float64(m.scanned) / float64(m.total)
		}
		bar := m.progress.ViewAs(percent)
		title := lipgloss.NewStyle().Bold(true).Render("Scanning dependencies…")
		return lipgloss.NewStyle().Padding(1, 2).Render(fmt.Sprintf("%s\n%s\n%d/%d", title, bar, m.scanned, m.total))
	case modeList:
		hint := lipgloss.NewStyle().Faint(true).Render("space: toggle  •  a: toggle-all  •  enter: update  •  q: quit")
		return lipgloss.NewStyle().Padding(1, 2).Render(m.list.View() + "\n" + hint)
	case modeUpdate:
		percent := 0.0
		if m.total > 0 {
			percent = float64(m.updated) / float64(m.total)
		}
		bar := m.progress.ViewAs(percent)
		title := lipgloss.NewStyle().Bold(true).Render("Updating modules…")
		// join rendered updated lines
		body := ""
		for _, ln := range m.updatedLines {
			body += ln + "\n"
		}
		return lipgloss.NewStyle().Padding(1, 2).Render(fmt.Sprintf("%s\n%s\n%d/%d\n%s", title, bar, m.updated, m.total, body))
	case modeDone:
		return lipgloss.NewStyle().Padding(1, 2).Render("Done")
	}
	return ""
}
