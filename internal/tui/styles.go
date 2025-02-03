package tui

import (
	"fmt"

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

func stepPrint(template string, args ...interface{}) tea.Cmd {
	line := stepIcon.Render() + " " + fmt.Sprintf(template, args...)

	return tea.Println(titleStyle.Render(line))
}

func textPrint(template string, args ...interface{}) tea.Cmd {
	line := fmt.Sprintf(template, args...)

	return tea.Println(printStyle.Render(line))
}

func printDone(m model) []tea.Cmd {
	cmds := []tea.Cmd{
		stepPrint("Done"),
		textPrint("%s %d succeeded", checkMark, len(m.upgrading)),
	}

	if len(m.upgradedFailed) > 0 {
		textPrint("%s %d failed", failMark, len(m.upgradedFailed))
		cmds = append(cmds, stepPrint("Failed"))
	}

	for _, mod := range m.upgradedFailed {
		cmds = append(cmds, textPrint("%s %s", failMark, mod.Path))
	}

	return cmds
}
