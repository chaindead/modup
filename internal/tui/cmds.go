//go:build !test

package tui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/chaindead/modup/internal/deps"
)

type getPackageListMsg struct {
	packages []string
	err      error
}

type getPackageInfoMsg struct {
	mod deps.Module
	err error
}

type upgradeModuleResultMsg struct {
	mod deps.Module
	err error
}

type changeModeListMsg bool

func changeModeList() tea.Cmd {
	return tea.Tick(time.Second/2, func(time.Time) tea.Msg {
		return changeModeListMsg(true)
	})
}

type beginUpgradeMsg struct {
	modules []deps.Module
}

func beginUpgradeCmd(selected []deps.Module) tea.Cmd {
	return func() tea.Msg { return beginUpgradeMsg{modules: selected} }
}

type moduleStartedMsg struct{}

func moduleStartedCmd() tea.Cmd {
	return func() tea.Msg {
		return moduleStartedMsg{}
	}
}
