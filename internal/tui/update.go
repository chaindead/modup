package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
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
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case progress.FrameMsg:
		progressModel, cmd := m.progress.Update(msg)
		m.progress = progressModel.(progress.Model)
		return m, cmd

	// get pkgs commands
	case getPackageListMsg:
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

		m.scanning = append(m.scanning, pkg)
		return m, getPkgInfo(pkg)
	case getPackageInfoMsg:
		m.index++
		if msg.mod.Updatable {
			m.modules = append(m.modules, msg.mod)
		}

		pkg := msg.mod.Path
		mark := checkMark
		if msg.err != nil {
			pkg = fmt.Sprintf("%s (%s)", msg.mod.Path, msg.err.Error())
			mark = failMark
		}

		if m.packages.cnt == m.index {
			return m, tea.Sequence(
				textPrint("%s %s", mark, pkg),
				changeModeList(),
			)
		}

		progressCmd := m.progress.SetPercent(float64(m.index) / float64(m.packages.cnt))

		// remove from scanning
		if pkg != "" {
			for i, p := range m.scanning {
				if p == pkg {
					m.scanning = append(m.scanning[:i], m.scanning[i+1:]...)
					break
				}
			}
		}

		return m, tea.Batch(progressCmd, textPrint("%s %s", mark, pkg), moduleStartedCmd())

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
		m.progress = progress.New(
			progress.WithDefaultGradient(),
			progress.WithWidth(40),
			progress.WithoutPercentage(),
		)

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
