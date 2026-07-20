package tui

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/Lotto724/herdr-system-metrics/internal/metrics"
	tea "github.com/charmbracelet/bubbletea"
)

func TestViewRendersResponsiveLayouts(t *testing.T) {
	model := viewModel()
	for _, test := range []struct {
		name   string
		width  int
		height int
		want   []string
		omit   []string
	}{
		{"full", 80, 16, []string{"System Metrics", "CPU: 25.0%", "Memory: 60/100", "Load (1/5/15): 1.00 2.00 3.00", "CPU trend:", "Memory trend:"}, nil},
		{"compact", 64, 12, []string{"System Metrics", "CPU: 25.0%", "Memory: 60/100", "CPU trend:", "Memory trend:"}, []string{"Swap:", "Load (1/5/15):"}},
		{"too small", 39, 12, []string{"terminal too small"}, nil},
	} {
		t.Run(test.name, func(t *testing.T) {
			updated, _ := model.Update(tea.WindowSizeMsg{Width: test.width, Height: test.height})
			view := updated.(Model).View()
			for _, text := range test.want {
				if !strings.Contains(view, text) {
					t.Fatalf("View() missing %q:\n%s", text, view)
				}
			}
			for _, text := range test.omit {
				if strings.Contains(view, text) {
					t.Fatalf("View() unexpectedly contains %q:\n%s", text, view)
				}
			}
		})
	}
}

func TestModelUpdateResizeRecomputesView(t *testing.T) {
	model := viewModel()
	updated, _ := model.Update(tea.WindowSizeMsg{Width: 80, Height: 16})
	model = updated.(Model)
	if !strings.Contains(model.View(), "CPU trend:") {
		t.Fatalf("full view = %q, want graph", model.View())
	}
	updated, _ = model.Update(tea.WindowSizeMsg{Width: 39, Height: 12})
	model = updated.(Model)
	if got := model.View(); got != "terminal too small\n" {
		t.Fatalf("resized view = %q, want terminal-too-small message", got)
	}
}

func TestCompactViewLinesFitTerminalWidth(t *testing.T) {
	for _, width := range []int{40, 64} {
		t.Run(fmt.Sprintf("%d columns", width), func(t *testing.T) {
			model := viewModel()
			updated, _ := model.Update(tea.WindowSizeMsg{Width: width, Height: 12})
			for _, line := range strings.Split(strings.TrimSuffix(updated.(Model).View(), "\n"), "\n") {
				if len(line) > width {
					t.Fatalf("compact line %q has width %d, want at most %d", line, len(line), width)
				}
			}
		})
	}
}

func TestViewUsesVisibleHistoryTailAndIsDeterministic(t *testing.T) {
	values := []float64{1, 2, 3, 4, 5}
	if got := visibleTail(values, 3); fmt.Sprint(got) != "[3 4 5]" {
		t.Fatalf("visibleTail(%v, 3) = %v, want [3 4 5]", values, got)
	}
	model := viewModel()
	updated, _ := model.Update(tea.WindowSizeMsg{Width: 64, Height: 12})
	model = updated.(Model)
	if first, second := model.View(), model.View(); first != second {
		t.Fatalf("View() was nondeterministic:\nfirst: %q\nsecond: %q", first, second)
	}
}

func TestViewLabelsWarmingUnavailableAndReadError(t *testing.T) {
	model := NewModel(sourceStub{}, metrics.NewState(1))
	model.snapshot = metrics.Snapshot{
		CPU:    metrics.Value[float64]{Status: metrics.WarmingUp},
		Memory: metrics.Value[metrics.MemoryUsage]{Status: metrics.Unavailable},
	}
	model.err = errors.New("proc read failed")
	updated, _ := model.Update(tea.WindowSizeMsg{Width: 80, Height: 16})
	view := updated.(Model).View()
	for _, text := range []string{"CPU: warming up", "Memory: unavailable", "Error: proc read failed"} {
		if !strings.Contains(view, text) {
			t.Fatalf("View() missing %q:\n%s", text, view)
		}
	}
}

func viewModel() Model {
	model := NewModel(sourceStub{}, metrics.NewState(8))
	model.snapshot = metrics.Snapshot{
		CPU: metrics.Value[float64]{Status: metrics.Available, Value: 25},
		Memory: metrics.Value[metrics.MemoryUsage]{
			Status: metrics.Available,
			Value:  metrics.MemoryUsage{Used: 60, Available: 40, Total: 100, SwapUsed: 20, SwapTotal: 50},
		},
		Load: metrics.Value[metrics.Load]{Status: metrics.Available, Value: metrics.Load{One: 1, Five: 2, Fifteen: 3}},
	}
	for _, raw := range []metrics.RawSample{sample(100, 20), sample(200, 70), sample(300, 145), sample(400, 245), sample(500, 370)} {
		model.state.Accept(raw)
	}
	return model
}
