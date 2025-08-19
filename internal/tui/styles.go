package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	// icons
	checkMark = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).SetString("✓")
	failMark  = lipgloss.NewStyle().Foreground(lipgloss.Color("1")).SetString("x")
	stepIcon  = lipgloss.NewStyle().Foreground(lipgloss.Color("33")).SetString("➤")

	// styles
	appStyle   = lipgloss.NewStyle().Padding(1, 2)
	titleStyle = lipgloss.NewStyle().
			MarginTop(1).
			Bold(true)
	printStyle          = lipgloss.NewStyle().MarginLeft(1)
	currentPkgNameStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("211"))
)

func newProgress() progress.Model {
	return progress.New(
		progress.WithDefaultGradient(),
		progress.WithWidth(40),
		progress.WithoutPercentage(),
	)
}

func newSpinner() spinner.Model {
	s := spinner.New()
	s.Spinner = spinner.MiniDot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("63"))
	return s
}

func stepPrint(template string, args ...interface{}) tea.Cmd {
	line := stepIcon.Render() + " " + fmt.Sprintf(template, args...)

	return tea.Println(titleStyle.Render(line))
}

func textPrint(template string, args ...interface{}) tea.Cmd {
	line := fmt.Sprintf(template, args...)

	return tea.Println(printStyle.Render(line))
}
