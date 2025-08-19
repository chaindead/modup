//go:build !fake

package tui

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/chaindead/modup/internal/deps"
)

func getPkgInfo(pkg string) tea.Cmd {
	return func() tea.Msg {
		mod, err := deps.GetModuleInfo(pkg)
		return getPackageInfoMsg{mod, err}
	}
}

func upgradeModule(mod deps.Module) tea.Cmd {
	return func() tea.Msg {
		err := deps.Upgrade(mod)
		return upgradeModuleResultMsg{mod: mod, err: err}
	}
}

func getPackageList() tea.Cmd {
	return func() tea.Msg {
		pkgs, err := deps.ListAllModulePaths()
		return getPackageListMsg{pkgs, err}
	}
}
