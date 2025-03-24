package tui

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/chaindead/modup/internal/deps"
)

type model struct {
	mode int
	// scan mode
	packages modules
	index    int
	modules  []deps.Module
	scanning []string

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
	p := progress.New(
		progress.WithDefaultGradient(),
		progress.WithWidth(40),
		progress.WithoutPercentage(),
	)
	s := spinner.New()
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("63"))
	return model{
		spinner:  s,
		progress: p,
		scanning: nil,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(getPackageList(), m.spinner.Tick)
}
