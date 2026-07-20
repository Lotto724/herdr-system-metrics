package tui

import (
	"fmt"
	"math"
	"strings"

	"github.com/Lotto724/herdr-system-metrics/internal/metrics"
)

const (
	minimumWidth  = 40
	minimumHeight = 8
	fullWidth     = 70
	fullHeight    = 14
)

type layoutMode uint8

const (
	layoutTooSmall layoutMode = iota
	layoutCompact
	layoutFull
)

type layout struct {
	mode       layoutMode
	graphWidth int
}

func layoutFor(width, height int) layout {
	if width < minimumWidth || height < minimumHeight {
		return layout{mode: layoutTooSmall}
	}
	graphWidth := min(width-24, 32)
	if width >= fullWidth && height >= fullHeight {
		return layout{mode: layoutFull, graphWidth: graphWidth}
	}
	return layout{mode: layoutCompact, graphWidth: min(graphWidth, 20)}
}

func renderView(model Model) string {
	layout := layoutFor(model.width, model.height)
	if layout.mode == layoutTooSmall {
		return "terminal too small\n"
	}
	lines := []string{
		"System Metrics (Linux/WSL best effort)",
		"CPU: " + cpuLabel(model.snapshot.CPU),
		"Memory: " + memoryLabel(model.snapshot.Memory),
	}
	if layout.mode == layoutFull {
		lines = append(lines, "Swap: "+swapLabel(model.snapshot.Memory), "Load (1/5/15): "+loadLabel(model.snapshot.Load))
	}
	lines = append(
		lines,
		"CPU trend: "+graph(model.state.CPUHistory(), layout.graphWidth),
		"Memory trend: "+memoryGraph(model.state.MemoryHistory(), layout.graphWidth),
	)
	if model.err != nil {
		lines = append(lines, "Error: "+model.err.Error())
	}
	lines = append(lines, "[space] pause/resume [r] reset [q] quit")
	return strings.Join(lines, "\n") + "\n"
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
	return fmt.Sprintf("%d/%d", value.Value.Used, value.Value.Total)
}

func swapLabel(value metrics.Value[metrics.MemoryUsage]) string {
	if value.Status != metrics.Available {
		return string(value.Status)
	}
	return fmt.Sprintf("%d/%d", value.Value.SwapUsed, value.Value.SwapTotal)
}

func loadLabel(value metrics.Value[metrics.Load]) string {
	if value.Status != metrics.Available {
		return string(value.Status)
	}
	return fmt.Sprintf("%.2f %.2f %.2f", value.Value.One, value.Value.Five, value.Value.Fifteen)
}

func visibleTail(values []float64, capacity int) []float64 { return tail(values, capacity) }

func graph(values []float64, width int) string {
	values = visibleTail(values, width)
	if len(values) == 0 {
		return strings.Repeat(".", width)
	}
	low, high := values[0], values[0]
	for _, value := range values[1:] {
		low, high = min(low, value), max(high, value)
	}
	var graph strings.Builder
	graph.WriteString(strings.Repeat(" ", width-len(values)))
	for _, value := range values {
		if high == low {
			graph.WriteByte('=')
			continue
		}
		level := int(math.Round((value - low) * 8 / (high - low)))
		graph.WriteByte(".:-=+*#%@"[level])
	}
	return graph.String()
}

func memoryGraph(values []uint64, width int) string {
	points := make([]float64, len(values))
	for index, value := range values {
		points[index] = float64(value)
	}
	return graph(points, width)
}

func tail(values []float64, capacity int) []float64 {
	if capacity <= 0 {
		return nil
	}
	if len(values) > capacity {
		values = values[len(values)-capacity:]
	}
	return values
}
