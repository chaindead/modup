package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/pflag"

	"github.com/chaindead/modup/internal/tui"
)

func main() {
	pflag.Parse()

	if _, err := tea.NewProgram(tui.NewModel()).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
