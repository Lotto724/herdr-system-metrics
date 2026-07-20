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
		{"full", 70, 7, []string{"System Metrics", "CPU: 25.0%", "Memory: 0.0 / 0.0 GB", "Swap: 0.0 / 0.0 GB", "[space] pause [q] quit"}, legacyDashboardText()},
		{"compact", 40, 7, []string{"System Metrics", "CPU: 25.0%", "Memory: 0.0 / 0.0 GB", "Swap: 0.0 / 0.0 GB", "[space] pause [q] quit"}, legacyDashboardText()},
		{"too narrow", 39, 7, []string{"terminal too small"}, nil},
		{"too short for separated footer", 40, 6, []string{"terminal too small"}, nil},
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

func TestDashboardOmitsLegacyTrendLoadAndResetUI(t *testing.T) {
	model := viewModel()
	updated, _ := model.Update(tea.WindowSizeMsg{Width: 70, Height: 7})
	view := updated.(Model).View()
	for _, text := range legacyDashboardText() {
		if strings.Contains(view, text) {
			t.Fatalf("View() unexpectedly contains %q:\n%s", text, view)
		}
	}
}

func legacyDashboardText() []string {
	return []string{"Load (1/5/15):", "CPU trend", "Memory trend", "[r]", "|", "+---"}
}

func TestModelUpdateResizeRecomputesView(t *testing.T) {
	model := viewModel()
	updated, _ := model.Update(tea.WindowSizeMsg{Width: 70, Height: 7})
	model = updated.(Model)
	if !strings.Contains(model.View(), "Swap:") {
		t.Fatalf("full view = %q, want current metrics", model.View())
	}
	updated, _ = model.Update(tea.WindowSizeMsg{Width: 39, Height: 7})
	model = updated.(Model)
	if got := model.View(); got != "terminal too small\n" {
		t.Fatalf("resized view = %q, want terminal-too-small message", got)
	}
}

func TestCompactViewLinesFitTerminalWidth(t *testing.T) {
	for _, width := range []int{40, 70} {
		t.Run(fmt.Sprintf("%d columns", width), func(t *testing.T) {
			model := viewModel()
			updated, _ := model.Update(tea.WindowSizeMsg{Width: width, Height: 7})
			for _, line := range strings.Split(strings.TrimSuffix(updated.(Model).View(), "\n"), "\n") {
				if len(line) > width {
					t.Fatalf("compact line %q has width %d, want at most %d", line, len(line), width)
				}
			}
		})
	}
}

func TestViewLinesFitFullAndCompactTerminalBoundaries(t *testing.T) {
	for _, test := range []struct {
		name   string
		width  int
		height int
	}{
		{"full", 70, 7},
		{"compact", 40, 7},
	} {
		t.Run(test.name, func(t *testing.T) {
			model := viewModel()
			updated, _ := model.Update(tea.WindowSizeMsg{Width: test.width, Height: test.height})
			lines := strings.Split(strings.TrimSuffix(updated.(Model).View(), "\n"), "\n")
			if len(lines) > test.height {
				t.Fatalf("View() rendered %d lines at %dx%d, want at most %d:\n%s", len(lines), test.width, test.height, test.height, updated.(Model).View())
			}
			for _, line := range lines {
				if len(line) > test.width {
					t.Fatalf("View() rendered line %q at %dx%d, want at most %d columns", line, test.width, test.height, test.width)
				}
			}
		})
	}
}

func TestViewReadErrorFitsAtLayoutThresholds(t *testing.T) {
	for _, test := range []struct {
		name   string
		width  int
		height int
		want   []string
	}{
		{"full", 70, 7, []string{"CPU:", "Memory:", "Swap:", "Error: proc read failed", "[space] pause [q] quit"}},
		{"compact", 40, 7, []string{"CPU:", "Memory:", "Swap:", "Error: proc read failed", "[space] pause [q] quit"}},
	} {
		t.Run(test.name, func(t *testing.T) {
			model := viewModel()
			model.err = errors.New("proc read failed")
			updated, _ := model.Update(tea.WindowSizeMsg{Width: test.width, Height: test.height})
			view := updated.(Model).View()
			lines := strings.Split(strings.TrimSuffix(view, "\n"), "\n")
			if len(lines) > test.height {
				t.Fatalf("View() rendered %d lines at %dx%d, want at most %d:\n%s", len(lines), test.width, test.height, test.height, view)
			}
			for _, text := range test.want {
				if !strings.Contains(view, text) {
					t.Fatalf("View() missing %q:\n%s", text, view)
				}
			}
			for _, text := range legacyDashboardText() {
				if strings.Contains(view, text) {
					t.Fatalf("View() unexpectedly contains %q:\n%s", text, view)
				}
			}
		})
	}
}

func TestViewFormatsMemoryAndSwapInDecimalGB(t *testing.T) {
	for _, test := range []struct {
		name  string
		used  uint64
		total uint64
		want  string
	}{
		{"decimal gigabytes", 6_300_000_000, 15_600_000_000, "6.3 / 15.6 GB"},
		{"zero and sub-gigabyte values", 1, 999_999_999, "0.0 / 1.0 GB"},
	} {
		t.Run(test.name, func(t *testing.T) {
			value := metrics.Value[metrics.MemoryUsage]{Status: metrics.Available, Value: metrics.MemoryUsage{Used: test.used, Total: test.total, SwapUsed: test.used, SwapTotal: test.total}}
			if got := memoryLabel(value); got != test.want {
				t.Fatalf("memoryLabel() = %q, want %q", got, test.want)
			}
			if got := swapLabel(value); got != test.want {
				t.Fatalf("swapLabel() = %q, want %q", got, test.want)
			}
		})
	}
}

func TestViewShowsOnlyAvailablePauseAction(t *testing.T) {
	model := viewModel()
	updated, _ := model.Update(tea.WindowSizeMsg{Width: 80, Height: 16})
	paused := updated.(Model)
	paused.paused = true
	for _, test := range []struct {
		name  string
		model Model
		want  string
		omit  string
	}{
		{"running", updated.(Model), "[space] pause [q] quit", "[space] resume"},
		{"paused", paused, "[space] resume [q] quit", "[space] pause"},
	} {
		t.Run(test.name, func(t *testing.T) {
			view := test.model.View()
			if !strings.Contains(view, test.want) || strings.Contains(view, "[r]") || strings.Contains(view, test.omit) {
				t.Fatalf("View() controls = %q, want %q without reset or %q", view, test.want, test.omit)
			}
		})
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

func TestViewHidesStaleMetricsAfterReadErrorAndPreservesControls(t *testing.T) {
	model := NewModel(sourceStub{}, metrics.NewState(1))
	for _, raw := range []metrics.RawSample{sample(100, 20), sample(200, 60)} {
		updated, _ := model.Update(sampleMsg{raw: raw})
		model = updated.(Model)
	}
	updated, _ := model.Update(sampleMsg{err: errors.New("proc read failed")})
	model = updated.(Model)
	updated, _ = model.Update(tea.WindowSizeMsg{Width: 70, Height: 7})
	view := updated.(Model).View()

	for _, text := range []string{"CPU: unavailable", "Memory: unavailable", "Swap: unavailable", "Error: proc read failed", "[space] pause [q] quit"} {
		if !strings.Contains(view, text) {
			t.Fatalf("View() missing %q after read error:\n%s", text, view)
		}
	}
	for _, stale := range []string{"CPU: 40.0%", "Memory: 0.0 / 0.0 GB", "Swap: 0.0 / 0.0 GB"} {
		if strings.Contains(view, stale) {
			t.Fatalf("View() retained stale value %q after read error:\n%s", stale, view)
		}
	}
}

func TestViewAnchorsSeparatedFooterAtBottom(t *testing.T) {
	for _, test := range []struct {
		name       string
		width      int
		height     int
		paused     bool
		err        error
		wantLines  int
		wantAction string
	}{
		{"running at minimum", 40, 7, false, nil, 7, "[space] pause [q] quit"},
		{"paused at taller height", 70, 10, true, nil, 10, "[space] resume [q] quit"},
		{"error at minimum", 40, 7, false, errors.New("proc read failed"), 7, "[space] pause [q] quit"},
		{"error at taller height", 70, 10, false, errors.New("proc read failed"), 10, "[space] pause [q] quit"},
	} {
		t.Run(test.name, func(t *testing.T) {
			model := viewModel()
			model.paused = test.paused
			model.err = test.err
			updated, _ := model.Update(tea.WindowSizeMsg{Width: test.width, Height: test.height})
			lines := strings.Split(strings.TrimSuffix(updated.(Model).View(), "\n"), "\n")

			if len(lines) != test.wantLines {
				t.Fatalf("View() rendered %d lines at %dx%d, want %d:\n%s", len(lines), test.width, test.height, test.wantLines, updated.(Model).View())
			}
			if got := lines[len(lines)-2]; got != strings.Repeat("-", test.width) {
				t.Fatalf("footer separator = %q, want width-sized ASCII separator", got)
			}
			if got := lines[len(lines)-1]; got != test.wantAction {
				t.Fatalf("footer action = %q, want %q", got, test.wantAction)
			}
			for _, line := range lines {
				if len(line) > test.width {
					t.Fatalf("View() clipped line %q at width %d", line, test.width)
				}
			}
			if test.height > 7 && lines[len(lines)-3] != "" {
				t.Fatalf("line before footer = %q, want blank padding", lines[len(lines)-3])
			}
		})
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
	}
	for _, raw := range []metrics.RawSample{sample(100, 20), sample(200, 70), sample(300, 145), sample(400, 245), sample(500, 370)} {
		model.state.Accept(raw)
	}
	return model
}
