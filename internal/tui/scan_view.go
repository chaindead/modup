package tui

import (
	"go-mod-upgrade/internal/deps"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) startScanCmd() tea.Cmd {
	return func() tea.Msg {
		total, err := deps.CountModules()
		if err != nil {
			return errMsg(err)
		}
		return scanTotalMsg(total)
	}
}

func (m Model) waitScanEventCmd() tea.Cmd {
	return func() tea.Msg {
		ticker := time.NewTicker(50 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case n, ok := <-m.scanProgCh:
				if !ok {
					continue
				}
				if n <= 0 {
					n = 1
				}
				return scanProgressMsg(n)
			case mods, ok := <-m.scanModsCh:
				if !ok {
					continue
				}
				return scanDoneMsg(mods)
			case err, ok := <-m.scanErrCh:
				if !ok {
					continue
				}
				return errMsg(err)
			case <-ticker.C:
				continue
			}
		}
	}
}
