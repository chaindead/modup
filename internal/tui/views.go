package tui

import (
	"fmt"
	"strings"

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
	n := len(m.packages)
	w := lipgloss.Width(fmt.Sprintf("%d", n))

	pkgCount := fmt.Sprintf(" %*d/%*d", w, m.index, w, n)

	spin := m.spinner.View() + " "
	prog := m.progress.View()
	cellsAvail := max(0, m.width-lipgloss.Width(spin+prog+pkgCount))

	info := "Loading modules list\n"
	if n != 0 {
		pkgName := currentPkgNameStyle.Render(m.packages[m.index])
		info = lipgloss.NewStyle().MaxWidth(cellsAvail).Render("Scanning " + pkgName)
	}

	cellsRemaining := max(0, m.width-lipgloss.Width(spin+info+prog+pkgCount))
	gap := strings.Repeat(" ", cellsRemaining)

	return spin + info + gap + prog + pkgCount
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
