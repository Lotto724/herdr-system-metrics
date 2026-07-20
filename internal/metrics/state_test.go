package metrics

import "testing"

func TestHistoryUsesFixedCapacityRing(t *testing.T) {
	history := newHistory[int](3)
	if len(history.values) != 3 || cap(history.values) != 3 {
		t.Fatalf("ring backing = len %d cap %d, want len/cap 3", len(history.values), cap(history.values))
	}
}

func TestHistoryWraparoundOrderingAndZeroCapacity(t *testing.T) {
	t.Run("wraparound retains newest values in order", func(t *testing.T) {
		history := newHistory[int](3)
		for _, value := range []int{1, 2, 3, 4, 5} {
			history.add(value)
		}
		if got, want := history.all(), []int{3, 4, 5}; !equalInts(got, want) {
			t.Fatalf("history = %v, want %v", got, want)
		}
	})

	t.Run("zero capacity discards values", func(t *testing.T) {
		history := newHistory[int](0)
		history.add(1)
		if got := history.all(); len(got) != 0 {
			t.Fatalf("zero-capacity history = %v, want empty", got)
		}
	})
}

func equalInts(got, want []int) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}

func TestStateAcceptCPUWarmupValidationAndRecovery(t *testing.T) {
	state := NewState(2)
	mem := Memory{Total: 100, Available: 40, SwapTotal: 20, SwapFree: 5}
	tests := []struct {
		name        string
		cpu         CPUCounters
		want        Status
		wantPercent float64
	}{
		{"first valid sample warms up", CPUCounters{Total: 100, Busy: 25}, WarmingUp, 0},
		{"second valid sample calculates delta", CPUCounters{Total: 200, Busy: 75}, Available, 50},
		{"regressing sample is unavailable", CPUCounters{Total: 150, Busy: 70}, Unavailable, 0},
		{"recovery starts with a new warmup", CPUCounters{Total: 300, Busy: 100}, WarmingUp, 0},
		{"later recovery sample calculates delta", CPUCounters{Total: 400, Busy: 150}, Available, 50},
		{"zero delta is unavailable", CPUCounters{Total: 400, Busy: 151}, Unavailable, 0},
		{"zero total is unavailable", CPUCounters{}, Unavailable, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := state.Accept(RawSample{CPU: tt.cpu, Memory: mem})
			if got.CPU.Status != tt.want || got.CPU.Value != tt.wantPercent {
				t.Fatalf("CPU = %#v, want %s/%v", got.CPU, tt.want, tt.wantPercent)
			}
		})
	}
}

func TestStateUsesMemoryAvailableAndBoundsHistory(t *testing.T) {
	state := NewState(2)
	for _, sample := range []RawSample{
		{CPU: CPUCounters{Total: 100, Busy: 20}, Memory: Memory{Total: 100, Available: 60, SwapTotal: 20, SwapFree: 15}},
		{CPU: CPUCounters{Total: 200, Busy: 60}, Memory: Memory{Total: 100, Available: 40, SwapTotal: 20, SwapFree: 10}},
		{CPU: CPUCounters{Total: 300, Busy: 120}, Memory: Memory{Total: 100, Available: 20, SwapTotal: 20, SwapFree: 5}},
	} {
		state.Accept(sample)
	}
	got := state.Snapshot()
	if got.Memory.Status != Available || got.Memory.Value.Used != 80 || got.Memory.Value.Available != 20 || got.Memory.Value.SwapUsed != 15 {
		t.Fatalf("memory = %#v", got.Memory)
	}
	if history := state.MemoryHistory(); len(history) != 2 || history[0] != 60 || history[1] != 80 {
		t.Fatalf("memory history = %#v", history)
	}
	if history := state.CPUHistory(); len(history) != 2 || history[0] != 40 || history[1] != 60 {
		t.Fatalf("CPU history = %#v", history)
	}
}

func TestStateInvalidatesBaseline(t *testing.T) {
	state := NewState(2)
	state.Accept(RawSample{CPU: CPUCounters{Total: 100, Busy: 20}, Memory: Memory{Total: 100, Available: 50}})
	state.Accept(RawSample{CPU: CPUCounters{Total: 200, Busy: 60}, Memory: Memory{Total: 100, Available: 40}})
	state.InvalidateCPUBaseline()
	if got := state.Accept(RawSample{CPU: CPUCounters{Total: 300, Busy: 100}, Memory: Memory{Total: 100, Available: 30}}); got.CPU.Status != WarmingUp {
		t.Fatalf("resumed CPU = %#v, want warming up", got.CPU)
	}
	if got := state.Snapshot(); got.Memory.Status != Available || got.Memory.Value.Used != 70 {
		t.Fatalf("baseline invalidation changed current state: %#v", got.Memory)
	}
}

func TestStateMarksCurrentMetricsUnavailableAndRecovers(t *testing.T) {
	state := NewState(2)
	state.Accept(RawSample{CPU: CPUCounters{Total: 100, Busy: 20}, Memory: Memory{Total: 100, Available: 50}})
	state.Accept(RawSample{CPU: CPUCounters{Total: 200, Busy: 60}, Memory: Memory{Total: 100, Available: 40}})

	got := state.MarkUnavailable()
	if got.CPU.Status != Unavailable || got.Memory.Status != Unavailable {
		t.Fatalf("unavailable snapshot = %#v, want CPU and memory unavailable", got)
	}
	if got := state.Accept(RawSample{CPU: CPUCounters{Total: 300, Busy: 100}, Memory: Memory{Total: 100, Available: 30}}); got.CPU.Status != WarmingUp || got.Memory.Status != Available {
		t.Fatalf("first recovered snapshot = %#v, want CPU warming up and memory available", got)
	}
	if got := state.Accept(RawSample{CPU: CPUCounters{Total: 400, Busy: 160}, Memory: Memory{Total: 100, Available: 20}}); got.CPU.Status != Available || got.CPU.Value != 60 || got.Memory.Status != Available {
		t.Fatalf("second recovered snapshot = %#v, want CPU and memory available", got)
	}
}
