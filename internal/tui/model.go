package tui

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/chaindead/modup/internal/deps"
)

type model struct {
	mode int
	// scan mode
	packages modules
	modules  []deps.Module
	scanning namedSpinners

	// choose mode
	list  list.Model
	items []list.Item

	// upgrade mode
	upgrading         []deps.Module
	upgradeIndex      int
	upgradeFailures   int
	upgradedSucceeded []deps.Module
	upgradedFailed    []deps.Module

	//common
	width    int
	height   int
	spinner  spinner.Model
	progress progress.Model
	done     bool
}

func NewModel() model {
	return model{
		spinner:  newSpinner(),
		progress: newProgress(),
		scanning: nil,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		stepPrint("Loading packages list"),
		getPackageList(),
		m.spinner.Tick)
}
