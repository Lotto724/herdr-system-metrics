package metrics

import "context"

type Status string

const (
	WarmingUp   Status = "warming up"
	Available   Status = "available"
	Unavailable Status = "unavailable"
)

type (
	Value[T any] struct {
		Status Status
		Value  T
	}
	CPUCounters struct{ Total, Busy uint64 }
	Memory      struct{ Total, Available, SwapTotal, SwapFree uint64 }
	Load        struct{ One, Five, Fifteen float64 }
	RawSample   struct {
		CPU    CPUCounters
		Memory Memory
		Load   Load
	}
	MemoryUsage struct{ Used, Available, Total, SwapUsed, SwapTotal uint64 }
	Snapshot    struct {
		CPU    Value[float64]
		Memory Value[MemoryUsage]
		Load   Value[Load]
	}
)

type Source interface {
	Read(context.Context) (RawSample, error)
}
