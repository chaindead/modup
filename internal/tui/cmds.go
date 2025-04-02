package tui

import (
	"fmt"
	"math/rand"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/chaindead/modup/internal/deps"
)

type getPackageListMsg struct {
	packages []string
	err      error
}

func getPackageList() tea.Cmd {
	if *test {
		time.Sleep(randomTestDelay())

		return func() tea.Msg {
			packages := make([]string, 0, len(fakeDeps))
			for pkg := range fakeDeps {
				packages = append(packages, pkg)
			}
			return getPackageListMsg{packages, nil}
		}
	}

	return func() tea.Msg {
		pkgs, err := deps.ListAllModulePaths()
		return getPackageListMsg{pkgs, err}
	}
}

type getPackageInfoMsg struct {
	mod deps.Module
	err error
}

func getPkgInfo(pkg string) tea.Cmd {
	return func() tea.Msg {
		if *test {
			time.Sleep(randomTestDelay())

			if mod, exists := fakeDeps[pkg]; exists {
				return getPackageInfoMsg{mod, nil}
			}
			return getPackageInfoMsg{deps.Module{Path: pkg}, nil}
		}

		mod, err := deps.GetModuleInfo(pkg)
		return getPackageInfoMsg{mod, err}
	}
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

type upgradeModuleResultMsg struct {
	mod deps.Module
	err error
}

func upgradeModule(mod deps.Module) tea.Cmd {
	return func() tea.Msg {
		if *test {
			time.Sleep(randomTestDelay())
			// 10% simulated failure
			r := rand.New(rand.NewSource(time.Now().UnixNano()))
			if r.Float64() < 0.10 {
				return upgradeModuleResultMsg{mod: mod, err: fmt.Errorf("simulated upgrade error")}
			}
			return upgradeModuleResultMsg{mod: mod, err: nil}
		}

		err := deps.Upgrade(mod)
		return upgradeModuleResultMsg{mod: mod, err: err}
	}
}

type moduleStartedMsg struct{}

func moduleStartedCmd() tea.Cmd {
	return func() tea.Msg {
		return moduleStartedMsg{}
	}
}
