package main

import (
	"fmt"
	"os"

	"github.com/spf13/pflag"

	"go-mod-upgrade/internal/tui"
)

func main() {
	var (
		dryRun   bool
		force    bool
		pageSize int
		hook     string
	)

	pflag.BoolVar(&dryRun, "dry-run", false, "show planned updates without applying changes")
	pflag.BoolVar(&force, "force", false, "select all updates and apply without interactive selection")
	pflag.IntVar(&pageSize, "pagesize", 12, "page size for interactive list")
	pflag.StringVar(&hook, "hook", "", "optional executable to run per updated module: hook <path> <from> <to>")
	pflag.Parse()

	if err := tui.RunApp(tui.Options{DryRun: dryRun, Force: force, PageSize: pageSize, Hook: hook}); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
