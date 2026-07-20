package main

import (
	"fmt"
	"os"

	"github.com/Lotto724/herdr-system-metrics/internal/metrics"
	"github.com/Lotto724/herdr-system-metrics/internal/procfs"
	"github.com/Lotto724/herdr-system-metrics/internal/tui"
	tea "github.com/charmbracelet/bubbletea"
)

const historyCapacity = 60

type programRunner interface {
	Run() (tea.Model, error)
}

func main() {
	if err := run(tea.NewProgram(newModel())); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func newModel() tui.Model {
	return tui.NewModel(procfs.NewReader("/proc"), metrics.NewState(historyCapacity))
}

func run(program programRunner) error {
	_, err := program.Run()
	return err
}
