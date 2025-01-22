package tui

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"

	"go-mod-upgrade/internal/deps"
)

type appMode int

const (
	modeScan appMode = iota
	modeList
	modeUpdate
	modeDone
)

type Options struct {
	DryRun bool
}

type (
	scanTotalMsg      int
	scanProgressMsg   int
	scanDoneMsg       []deps.Module
	updateProgressMsg int
	updateDoneMsg     struct{}
	errMsg            error
)

type Model struct {
	mode    appMode
	opts    Options
	total   int
	scanned int
	updated int
	width   int
	height  int

	spinner  spinner.Model
	progress progress.Model
	list     list.Model
	items    []listModuleItem
	keys     listKeys
	err      error

	// scan channels
	scanProgCh chan int
	scanModsCh chan []deps.Module
	scanErrCh  chan error
	// update channels
	updProgCh    chan int
	updErrCh     chan error
	updDetailCh  chan deps.Module
	updatedLines []string
}

func NewApp(opts Options) Model {
	sp := spinner.New()
	sp.Spinner = spinner.MiniDot
	pr := progress.New(progress.WithDefaultGradient())
	return Model{mode: modeScan, opts: opts, spinner: sp, progress: pr, keys: newListKeys()}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, m.startScanCmd())
}

// RunApp starts the full application lifecycle
func RunApp(opts Options) error {
	m := NewApp(opts)
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}
