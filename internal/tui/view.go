package tui

import (
	"fmt"
	"strings"

	"github.com/Lotto724/herdr-system-metrics/internal/metrics"
)

const (
	minimumWidth  = 40
	minimumHeight = 7
)

func renderView(model Model) string {
	if model.width < minimumWidth || model.height < minimumHeight {
		return "terminal too small\n"
	}
	content := []string{
		fitLine("System Metrics (Linux/WSL best effort)", model.width),
		fitLine("CPU: "+cpuLabel(model.snapshot.CPU), model.width),
		fitLine("Memory: "+memoryLabel(model.snapshot.Memory), model.width),
		fitLine("Swap: "+swapLabel(model.snapshot.Memory), model.width),
	}
	if model.err != nil {
		content = append(content, fitLine("Error: "+model.err.Error(), model.width))
	}
	action := "pause"
	if model.paused {
		action = "resume"
	}
	footer := []string{
		strings.Repeat("-", model.width),
		fitLine("[space] "+action+" [q] quit", model.width),
	}
	lines := append(content, make([]string, model.height-len(content)-len(footer))...)
	lines = append(lines, footer...)
	return strings.Join(lines, "\n") + "\n"
}

func fitLine(value string, width int) string {
	if len(value) <= width {
		return value
	}
	return value[:width]
}

func cpuLabel(value metrics.Value[float64]) string {
	if value.Status != metrics.Available {
		return string(value.Status)
	}
	return fmt.Sprintf("%.1f%%", value.Value)
}

func memoryLabel(value metrics.Value[metrics.MemoryUsage]) string {
	if value.Status != metrics.Available {
		return string(value.Status)
	}
	return formatGB(value.Value.Used) + " / " + formatGB(value.Value.Total) + " GB"
}

func swapLabel(value metrics.Value[metrics.MemoryUsage]) string {
	if value.Status != metrics.Available {
		return string(value.Status)
	}
	return formatGB(value.Value.SwapUsed) + " / " + formatGB(value.Value.SwapTotal) + " GB"
}

func formatGB(bytes uint64) string { return fmt.Sprintf("%.1f", float64(bytes)/1_000_000_000) }
