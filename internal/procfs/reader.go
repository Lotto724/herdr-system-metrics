package procfs

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/Lotto724/herdr-system-metrics/internal/metrics"
)

const maxProcFileBytes = 64 * 1024

var ErrFileTooLarge = errors.New("proc file exceeds size limit")

type Reader struct{ root string }

func NewReader(root string) Reader { return Reader{root: root} }

func (r Reader) ReadStat() (CPUCounters, error) {
	b, err := r.read("stat")
	if err != nil {
		return CPUCounters{}, err
	}
	return ParseStat(bytes.NewReader(b))
}

func (r Reader) ReadMemInfo() (Memory, error) {
	b, err := r.read("meminfo")
	if err != nil {
		return Memory{}, err
	}
	return ParseMemInfo(bytes.NewReader(b))
}

func (r Reader) ReadLoadAvg() (Load, error) {
	b, err := r.read("loadavg")
	if err != nil {
		return Load{}, err
	}
	return ParseLoadAvg(bytes.NewReader(b))
}

func (r Reader) Read(ctx context.Context) (metrics.RawSample, error) {
	if err := ctx.Err(); err != nil {
		return metrics.RawSample{}, err
	}
	cpu, err := r.ReadStat()
	if err != nil {
		return metrics.RawSample{}, err
	}
	memory, err := r.ReadMemInfo()
	if err != nil {
		return metrics.RawSample{}, err
	}
	load, err := r.ReadLoadAvg()
	if err != nil {
		return metrics.RawSample{}, err
	}
	return metrics.RawSample{
		CPU:    metrics.CPUCounters{Total: cpu.Total, Busy: cpu.Busy},
		Memory: metrics.Memory{Total: memory.Total, Available: memory.Available, SwapTotal: memory.SwapTotal, SwapFree: memory.SwapFree},
		Load:   metrics.Load{One: load.One, Five: load.Five, Fifteen: load.Fifteen},
	}, nil
}

func (r Reader) read(name string) ([]byte, error) {
	f, err := os.Open(filepath.Join(r.root, name))
	if err != nil {
		return nil, fmt.Errorf("open /proc/%s: %w", name, err)
	}
	defer f.Close()
	b, err := io.ReadAll(io.LimitReader(f, maxProcFileBytes+1))
	if err != nil {
		return nil, fmt.Errorf("read /proc/%s: %w", name, err)
	}
	if len(b) > maxProcFileBytes {
		return nil, ErrFileTooLarge
	}
	return b, nil
}
