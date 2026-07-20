package tui

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Lotto724/herdr-system-metrics/internal/metrics"
	tea "github.com/charmbracelet/bubbletea"
)

type sourceStub struct{}

func (sourceStub) Read(context.Context) (metrics.RawSample, error) { return metrics.RawSample{}, nil }

func TestModelInitSchedulesOneSecondTick(t *testing.T) {
	model := NewModel(sourceStub{}, metrics.NewState(2))
	started := time.Now()
	msg := model.Init()()
	if _, ok := msg.(tickMsg); !ok {
		t.Fatalf("Init message = %T, want tickMsg", msg)
	}
	if elapsed := time.Since(started); elapsed < 900*time.Millisecond || elapsed > 2*time.Second {
		t.Fatalf("tick delay = %v, want one-second cadence", elapsed)
	}
}

func TestModelUpdateDoesNotOverlapReads(t *testing.T) {
	model := NewModel(sourceStub{}, metrics.NewState(2))
	updated, _ := model.Update(tickMsg{})
	model = updated.(Model)
	if !model.reading {
		t.Fatal("first tick did not start a read")
	}
	updated, command := model.Update(tickMsg{})
	model = updated.(Model)
	if !model.reading || command == nil {
		t.Fatal("overlapping tick did not preserve in-flight read and cadence")
	}
}

func TestModelUpdatePauseDiscardsGenerationAndResumeWarmsCPU(t *testing.T) {
	state := metrics.NewState(2)
	model := NewModel(sourceStub{}, state)
	updated, _ := model.Update(tickMsg{})
	model = updated.(Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeySpace})
	model = updated.(Model)
	if !model.paused || model.generation != 1 {
		t.Fatalf("pause = paused:%t generation:%d, want true:1", model.paused, model.generation)
	}
	updated, _ = model.Update(sampleMsg{generation: 0, raw: sample(100, 20)})
	model = updated.(Model)
	if model.snapshot.CPU.Status != "" {
		t.Fatalf("discarded sample changed snapshot: %#v", model.snapshot)
	}
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeySpace})
	model = updated.(Model)
	if model.paused {
		t.Fatal("resume did not unpause")
	}
	updated, _ = model.Update(sampleMsg{generation: 1, raw: sample(200, 40)})
	model = updated.(Model)
	if model.snapshot.CPU.Status != metrics.WarmingUp {
		t.Fatalf("resumed CPU = %s, want %s", model.snapshot.CPU.Status, metrics.WarmingUp)
	}
}

func TestModelUpdateIgnoresResetKey(t *testing.T) {
	model := NewModel(sourceStub{}, metrics.NewState(2))
	for _, raw := range []metrics.RawSample{sample(100, 20), sample(200, 60)} {
		updated, _ := model.Update(sampleMsg{raw: raw})
		model = updated.(Model)
	}
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")})
	model = updated.(Model)
	if len(model.state.CPUHistory()) != 1 || model.snapshot.Memory.Status != metrics.Available {
		t.Fatalf("reset key changed history=%v snapshot=%#v", model.state.CPUHistory(), model.snapshot)
	}
}

func TestModelUpdateQuitsForExitKeys(t *testing.T) {
	for _, key := range []tea.KeyMsg{{Type: tea.KeyRunes, Runes: []rune("q")}, {Type: tea.KeyEsc}, {Type: tea.KeyCtrlC}} {
		_, command := NewModel(sourceStub{}, metrics.NewState(1)).Update(key)
		if command == nil || command() != tea.Quit() {
			t.Fatalf("key %q did not return tea.Quit", key.String())
		}
	}
}

func TestModelUpdateMarksMetricsUnavailableAfterSampleErrorAndRecovers(t *testing.T) {
	model := NewModel(sourceStub{}, metrics.NewState(1))
	for _, raw := range []metrics.RawSample{sample(100, 20), sample(200, 60)} {
		updated, _ := model.Update(sampleMsg{raw: raw})
		model = updated.(Model)
	}
	updated, _ := model.Update(sampleMsg{err: errors.New("proc read failed")})
	model = updated.(Model)
	if model.err == nil || model.snapshot.CPU.Status != metrics.Unavailable || model.snapshot.Memory.Status != metrics.Unavailable {
		t.Fatalf("sample error = %v snapshot=%#v, want retained error and unavailable metrics", model.err, model.snapshot)
	}
	updated, _ = model.Update(sampleMsg{raw: sample(100, 20)})
	model = updated.(Model)
	if model.err != nil || model.snapshot.CPU.Status != metrics.WarmingUp || model.snapshot.Memory.Status != metrics.Available {
		t.Fatalf("first recovery error=%v snapshot=%#v", model.err, model.snapshot)
	}
	updated, _ = model.Update(sampleMsg{raw: sample(200, 60)})
	model = updated.(Model)
	if model.snapshot.CPU.Status != metrics.Available || model.snapshot.CPU.Value != 40 || model.snapshot.Memory.Status != metrics.Available {
		t.Fatalf("second recovery snapshot=%#v", model.snapshot)
	}
}

func sample(total, busy uint64) metrics.RawSample {
	return metrics.RawSample{CPU: metrics.CPUCounters{Total: total, Busy: busy}, Memory: metrics.Memory{Total: 100, Available: 40}}
}
