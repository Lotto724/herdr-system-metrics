package procfs

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

type Memory struct{ Total, Available, SwapTotal, SwapFree uint64 }

func ParseMemInfo(r io.Reader) (Memory, error) {
	values := map[string]uint64{}
	s := bufio.NewScanner(r)
	for s.Scan() {
		fields := strings.Fields(s.Text())
		if len(fields) < 3 {
			continue
		}
		key := strings.TrimSuffix(fields[0], ":")
		switch key {
		case "MemTotal", "MemAvailable", "SwapTotal", "SwapFree":
		default:
			continue
		}
		if fields[2] != "kB" {
			return Memory{}, fmt.Errorf("%s: invalid unit", key)
		}
		v, err := strconv.ParseUint(fields[1], 10, 64)
		if err != nil || v > ^uint64(0)/1024 {
			return Memory{}, fmt.Errorf("%s: invalid value", key)
		}
		values[key] = v * 1024
	}
	if err := s.Err(); err != nil {
		return Memory{}, fmt.Errorf("read meminfo: %w", err)
	}
	for _, key := range []string{"MemTotal", "MemAvailable", "SwapTotal", "SwapFree"} {
		if _, ok := values[key]; !ok {
			return Memory{}, fmt.Errorf("%s missing", key)
		}
	}
	return Memory{values["MemTotal"], values["MemAvailable"], values["SwapTotal"], values["SwapFree"]}, nil
}
