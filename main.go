package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/pflag"

	"github.com/chaindead/modup/internal/tui"
)

var (
	buildTag    = "dev"
	showVersion = pflag.BoolP("version", "v", false, "show version")
)

func main() {
	pflag.Parse()

	if *showVersion {
		fmt.Println("modup", buildTag)
		os.Exit(0)
	}

	if _, err := tea.NewProgram(tui.NewModel()).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
