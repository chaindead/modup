package tui

import (
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"go-mod-upgrade/internal/deps"
)

type updatedModuleMsg deps.Module

func (m *Model) startUpdate(sel []deps.Module) tea.Cmd {
	// legacy numeric progress channel remains for percent, but also capture detailed events
	m.updProgCh = make(chan int, 32)
	m.updErrCh = make(chan error, 1)
	detailed := make(chan deps.Module, 32)
	m.updDetailCh = detailed
	go func() {
		// run updater emitting detailed modules
		err := deps.ApplyProgressDetailed(sel, detailed)
		if err != nil {
			m.updErrCh <- err
		} else {
			m.updErrCh <- nil
		}
		close(detailed)
		close(m.updProgCh)
		close(m.updErrCh)
	}()

	return m.waitUpdateEventCmd()
}

func (m Model) waitUpdateEventCmd() tea.Cmd {
	return func() tea.Msg {
		ticker := time.NewTicker(50 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case u, ok := <-m.updDetailCh:
				if !ok {
					continue
				}
				return updatedModuleMsg(u)
			case n, ok := <-m.updProgCh:
				if !ok {
					continue
				}
				if n <= 0 {
					n = 1
				}
				return updateProgressMsg(m.updated + n)
			case err, ok := <-m.updErrCh:
				if !ok {
					continue
				}
				if err != nil {
					return errMsg(err)
				}
				return updateDoneMsg{}
			case <-ticker.C:
				continue
			}
		}
	}
}

// augment View to print per-item success lines similar to package-manager example
var (
	checkMark = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).SetString("✓")
	arrow     = lipgloss.NewStyle().Foreground(lipgloss.Color("63")).SetString("→")
)

func (m Model) renderUpdateLine(mod deps.Module) string {
	from := "v" + mod.Current.String()
	to := "v" + mod.Latest.String()
	name := lipgloss.NewStyle().Foreground(lipgloss.Color("211")).Render(mod.Path)
	return strings.Join([]string{checkMark.String(), name, from, arrow.String(), to}, " ")
}
