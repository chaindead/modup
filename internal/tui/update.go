package tui

import (
	"fmt"
	"sort"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/chaindead/modup/internal/deps"
)

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(tea.WindowSizeMsg); ok {
		m.width, m.height = msg.Width, msg.Height

		if len(m.list.Items()) != 0 {
			h, v := appStyle.GetFrameSize()
			m.list.SetSize(msg.Width-h, msg.Height-v)
		}
	}

	_, listFinished := msg.(beginUpgradeMsg)

	if m.mode == modeList && !listFinished {
		return m.listUpdate(msg)
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc", "q":
			return m, tea.Quit
		}

	case spinner.TickMsg:
		var cmd tea.Cmd

		if msg.ID == m.spinner.ID() {
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}

		for i := range m.scanning {
			if msg.ID == m.scanning[i].spin.ID() {
				m.scanning[i].spin, cmd = m.scanning[i].spin.Update(msg)
				break
			}
		}

		return m, cmd
	case progress.FrameMsg:
		progressModel, cmd := m.progress.Update(msg)
		m.progress = progressModel.(progress.Model)
		return m, cmd

	// get pkgs commands
	case getPackageListMsg:
		if msg.err != nil {
			return m, tea.Batch(
				tea.Println("list used packages:", msg.err),
				tea.Quit,
			)
		}
		m.packages = createModules(msg.packages)

		cmds := []tea.Cmd{
			stepPrint("Getting info about %d packages", m.packages.cnt),
		}
		for i := uint(0); i < *workerCnt; i++ {
			cmds = append(cmds, moduleStartedCmd())
		}

		return m, tea.Batch(cmds...)
	case moduleStartedMsg:
		pkg, ok := m.packages.next()
		if !ok {
			return m, nil
		}

		m.scanning = append(m.scanning, namedSpinner{
			name: pkg,
			spin: newSpinner(),
		})
		return m, tea.Batch(
			getPkgInfo(pkg),
			m.scanning.lastSpinner().Tick,
		)
	case getPackageInfoMsg:
		m.packages.current++
		m.scanning = m.scanning.remove(msg.mod.Path)
		if msg.mod.Updatable {
			m.modules = append(m.modules, msg.mod)
		}

		pkg := msg.mod.Path
		mark := checkMark
		if msg.err != nil {
			pkg = fmt.Sprintf("%s (%s)", msg.mod.Path, msg.err.Error())
			mark = failMark
		}

		if !m.packages.isFinished() {
			progressCmd := m.progress.SetPercent(m.packages.progressFloat())
			return m, tea.Batch(progressCmd, textPrint("%s %s", mark, pkg), moduleStartedCmd())
		}

		if len(m.modules) == 0 {
			return m, tea.Sequence(
				stepPrint("Everything is up-to-date"),
				tea.Quit,
			)
		}
		m.modules = sortModules(m.modules)

		return m, tea.Sequence(
			textPrint("%s %s", mark, pkg),
			changeModeList(),
		)

	case changeModeListMsg:
		l, items := m.newList()
		m.list = l
		m.items = items
		m.mode = modeList

		return m, nil

	// Begin upgrade flow
	case beginUpgradeMsg:
		m.upgrading = msg.modules
		m.upgradeIndex = 0
		m.upgradeFailures = 0
		m.upgradedSucceeded = nil
		m.upgradedFailed = nil
		m.mode = modeUpgrade
		m.progress = newProgress()

		cmds := []tea.Cmd{
			stepPrint("Upgrading %d packages", len(m.upgrading)),
			m.spinner.Tick,
		}
		for _, mod := range m.upgrading {
			cmds = append(cmds, upgradeModule(mod))
		}
		return m, tea.Sequence(cmds...)

	case upgradeModuleResultMsg:
		mark := checkMark
		if msg.err != nil {
			mark = failMark
			m.upgradeFailures++
			m.upgradedFailed = append(m.upgradedFailed, msg.mod)
		} else {
			m.upgradedSucceeded = append(m.upgradedSucceeded, msg.mod)
		}
		m.upgradeIndex++
		progressCmd := m.progress.SetPercent(float64(m.upgradeIndex) / float64(len(m.upgrading)))
		if m.upgradeIndex >= len(m.upgrading) {
			doneCmds := []tea.Cmd{
				progressCmd,
				textPrint("%s %s", mark, msg.mod.Path),
			}
			doneCmds = append(doneCmds, m.printDone()...)
			doneCmds = append(doneCmds, tea.Quit)

			return m, tea.Sequence(doneCmds...)
		}
		return m, tea.Batch(
			progressCmd,
			textPrint("%s %s", mark, msg.mod.Path),
		)
	}

	return m, nil
}

var categoryMap = map[string]int{
	"minor":      1,
	"patch":      2,
	"prerelease": 3,
	"metadata":   4,
}

func sortModules(ms []deps.Module) []deps.Module {
	sort.Slice(ms, func(i, j int) bool {
		return categoryMap[ms[i].UpdateCategory] < categoryMap[ms[j].UpdateCategory]
	})

	return ms
}
