package procfs

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

type CPUCounters struct{ Total, Busy uint64 }

func ParseStat(r io.Reader) (CPUCounters, error) {
	s := bufio.NewScanner(r)
	for s.Scan() {
		fields := strings.Fields(s.Text())
		if len(fields) == 0 || fields[0] != "cpu" {
			continue
		}
		if len(fields) < 9 {
			return CPUCounters{}, fmt.Errorf("cpu: need eight counters")
		}
		var values [8]uint64
		var total uint64
		for i := range values {
			v, err := strconv.ParseUint(fields[i+1], 10, 64)
			if err != nil {
				return CPUCounters{}, fmt.Errorf("cpu counter %d: %w", i, err)
			}
			if ^uint64(0)-total < v {
				return CPUCounters{}, fmt.Errorf("cpu total overflows")
			}
			values[i], total = v, total+v
		}
		idle := values[3] + values[4]
		return CPUCounters{Total: total, Busy: total - idle}, nil
	}
	if err := s.Err(); err != nil {
		return CPUCounters{}, fmt.Errorf("read stat: %w", err)
	}
	return CPUCounters{}, fmt.Errorf("aggregate cpu missing")
}
