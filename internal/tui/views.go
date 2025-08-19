package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	modeScan = iota
	modeList
	modeUpgrade
)

func (m model) View() string {
	switch m.mode {
	case modeScan:
		return m.viewScan()
	case modeList:
		return m.viewList()
	case modeUpgrade:
		return m.viewUpgrade()
	default:
		panic("unreachable")
	}
}

func (m model) viewScan() string {
	n := m.packages.cnt
	w := lipgloss.Width(fmt.Sprintf("%d", n))

	pkgCount := fmt.Sprintf(" %*d/%*d", w, m.packages.current, w, n)
	prog := m.progress.View()

	var lines []string
	for _, p := range m.scanning {
		name := currentPkgNameStyle.Render(p.name)
		lines = append(lines, p.spin.View()+" Scanning "+name)
	}

	body := strings.Join(lines, "\n")

	// bottom-right align progress and counter
	footer := prog + pkgCount
	cellsRemaining := max(0, m.width-lipgloss.Width(footer))
	gap := strings.Repeat(" ", cellsRemaining)

	return body + "\n" + gap + footer
}

func (m model) viewList() string {
	return appStyle.Render(m.list.View())
}
func (m model) viewUpgrade() string {
	n := len(m.upgrading)
	w := lipgloss.Width(fmt.Sprintf("%d", n))

	count := fmt.Sprintf(" %*d/%*d", w, m.upgradeIndex, w, n)
	spin := m.spinner.View() + " "
	prog := m.progress.View()
	cellsAvail := max(0, m.width-lipgloss.Width(spin+prog+count))

	var info string
	if n != 0 && m.upgradeIndex < n {
		pkgName := currentPkgNameStyle.Render(m.upgrading[m.upgradeIndex].Path)
		info = lipgloss.NewStyle().MaxWidth(cellsAvail).Render("Upgrading " + pkgName)
	}

	cellsRemaining := max(0, m.width-lipgloss.Width(spin+info+prog+count))
	gap := strings.Repeat(" ", cellsRemaining)

	return spin + info + gap + prog + count
}

func (m model) printDone() []tea.Cmd {
	cmds := []tea.Cmd{
		stepPrint("Done"),
		textPrint("%s %d succeeded!", checkMark, len(m.upgrading)),
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
