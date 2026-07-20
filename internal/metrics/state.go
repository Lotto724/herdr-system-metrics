package metrics

type State struct {
	previous    CPUCounters
	hasPrevious bool
	snapshot    Snapshot
	cpu         history[float64]
	memory      history[uint64]
}

func NewState(capacity int) *State {
	if capacity < 0 {
		capacity = 0
	}
	return &State{cpu: newHistory[float64](capacity), memory: newHistory[uint64](capacity)}
}

func (s *State) Accept(raw RawSample) Snapshot {
	s.snapshot.CPU = s.acceptCPU(raw.CPU)
	s.snapshot.Memory = memoryValue(raw.Memory)
	if s.snapshot.CPU.Status == Available {
		s.cpu.add(s.snapshot.CPU.Value)
	}
	if s.snapshot.Memory.Status == Available {
		s.memory.add(s.snapshot.Memory.Value.Used)
	}
	return s.snapshot
}

func (s *State) acceptCPU(current CPUCounters) Value[float64] {
	if current.Total == 0 || current.Busy > current.Total {
		s.hasPrevious = false
		return Value[float64]{Status: Unavailable}
	}
	if !s.hasPrevious {
		s.previous, s.hasPrevious = current, true
		return Value[float64]{Status: WarmingUp}
	}
	if current.Total <= s.previous.Total || current.Busy < s.previous.Busy {
		s.hasPrevious = false
		return Value[float64]{Status: Unavailable}
	}
	total := current.Total - s.previous.Total
	busy := current.Busy - s.previous.Busy
	if busy > total {
		s.hasPrevious = false
		return Value[float64]{Status: Unavailable}
	}
	s.previous = current
	return Value[float64]{Status: Available, Value: float64(busy) * 100 / float64(total)}
}

func memoryValue(memory Memory) Value[MemoryUsage] {
	if memory.Total == 0 || memory.Available > memory.Total || memory.SwapFree > memory.SwapTotal {
		return Value[MemoryUsage]{Status: Unavailable}
	}
	return Value[MemoryUsage]{Status: Available, Value: MemoryUsage{memory.Total - memory.Available, memory.Available, memory.Total, memory.SwapTotal - memory.SwapFree, memory.SwapTotal}}
}

func (s *State) Snapshot() Snapshot      { return s.snapshot }
func (s *State) CPUHistory() []float64   { return s.cpu.all() }
func (s *State) MemoryHistory() []uint64 { return s.memory.all() }
func (s *State) InvalidateCPUBaseline()  { s.hasPrevious = false }

func (s *State) MarkUnavailable() Snapshot {
	s.snapshot.CPU = Value[float64]{Status: Unavailable}
	s.snapshot.Memory = Value[MemoryUsage]{Status: Unavailable}
	s.InvalidateCPUBaseline()
	return s.snapshot
}
