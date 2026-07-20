package procfs

import (
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"
)

type Load struct{ One, Five, Fifteen float64 }

func ParseLoadAvg(r io.Reader) (Load, error) {
	b, err := io.ReadAll(r)
	if err != nil {
		return Load{}, fmt.Errorf("read loadavg: %w", err)
	}
	fields := strings.Fields(string(b))
	if len(fields) < 3 {
		return Load{}, fmt.Errorf("loadavg: need three values")
	}
	values := [3]float64{}
	for i := range values {
		values[i], err = strconv.ParseFloat(fields[i], 64)
		if err != nil {
			return Load{}, fmt.Errorf("load %d: %w", i, err)
		}
		if math.IsNaN(values[i]) || math.IsInf(values[i], 0) {
			return Load{}, fmt.Errorf("load %d: must be finite", i)
		}
	}
	return Load{values[0], values[1], values[2]}, nil
}
