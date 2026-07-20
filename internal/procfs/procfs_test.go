package procfs

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/Lotto724/herdr-system-metrics/internal/metrics"
)

func TestParseStat(t *testing.T) {
	valid := "cpu  100 20 30 400 50 10 5 5 7 3\ncpu0 1 2 3 4\n"
	tests := []struct {
		name    string
		input   string
		want    CPUCounters
		wantErr bool
	}{
		{"aggregate counters exclude guests", valid, CPUCounters{Total: 620, Busy: 170}, false},
		{"aggregate line can follow other lines", "intr 9\n" + valid, CPUCounters{Total: 620, Busy: 170}, false},
		{"missing aggregate is unavailable", "cpu0 1 2 3 4\n", CPUCounters{}, true},
		{"malformed counter is unavailable", "cpu 1 2 nope 4 5 6 7 8\n", CPUCounters{}, true},
		{"overflow is unavailable", "cpu 18446744073709551615 1 1 1 1 1 1 1\n", CPUCounters{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseStat(strings.NewReader(tt.input))
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseStat() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Fatalf("ParseStat() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestParseMemInfo(t *testing.T) {
	valid := "MemTotal: 100 kB\nMemAvailable: 40 kB\nSwapTotal: 50 kB\nSwapFree: 10 kB\n"
	tests := []struct {
		name, input string
		want        Memory
		wantErr     bool
	}{
		{"uses available despite reordered fields", "SwapFree: 10 kB\nMemAvailable: 40 kB\nMemTotal: 100 kB\nSwapTotal: 50 kB\n", Memory{Total: 102400, Available: 40960, SwapTotal: 51200, SwapFree: 10240}, false},
		{"missing available is unavailable", "MemTotal: 100 kB\nSwapTotal: 50 kB\nSwapFree: 10 kB\n", Memory{}, true},
		{"malformed value is unavailable", strings.Replace(valid, "40 kB", "nope kB", 1), Memory{}, true},
		{"kilobyte overflow is unavailable", strings.Replace(valid, "100 kB", "18446744073709551615 kB", 1), Memory{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseMemInfo(strings.NewReader(tt.input))
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseMemInfo() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Fatalf("ParseMemInfo() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestParseLoadAvg(t *testing.T) {
	tests := []struct {
		name, input string
		want        Load
		wantErr     bool
	}{
		{"three load windows", "0.50 1.25 2.75 1/2 3\n", Load{One: .5, Five: 1.25, Fifteen: 2.75}, false},
		{"extra fields do not affect load", "0.50 1.25 2.75 9/9 99\n", Load{One: .5, Five: 1.25, Fifteen: 2.75}, false},
		{"missing windows are unavailable", "0.50 1.25\n", Load{}, true},
		{"malformed window is unavailable", "0.50 nope 2.75\n", Load{}, true},
		{"NaN window is unavailable", "NaN 1.25 2.75\n", Load{}, true},
		{"positive infinity window is unavailable", "0.50 +Inf 2.75\n", Load{}, true},
		{"negative infinity window is unavailable", "0.50 1.25 -Inf\n", Load{}, true},
		{"negative one-minute window is unavailable", "-0.50 1.25 2.75\n", Load{}, true},
		{"negative five-minute window is unavailable", "0.50 -1.25 2.75\n", Load{}, true},
		{"negative fifteen-minute window is unavailable", "0.50 1.25 -2.75\n", Load{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseLoadAvg(strings.NewReader(tt.input))
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseLoadAvg() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Fatalf("ParseLoadAvg() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestReaderRejectsOversizedProcFiles(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(root+"/stat", []byte(strings.Repeat("x", maxProcFileBytes+1)), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := NewReader(root).ReadStat()
	if !errors.Is(err, ErrFileTooLarge) {
		t.Fatalf("ReadStat() error = %v, want ErrFileTooLarge", err)
	}
}

func TestReaderProvidesMetricsSourceSample(t *testing.T) {
	root := t.TempDir()
	files := map[string]string{
		"stat":    "cpu 10 0 0 80 0 0 0 0 9 9\n",
		"meminfo": "MemTotal: 100 kB\nMemAvailable: 40 kB\nSwapTotal: 50 kB\nSwapFree: 10 kB\n",
		"loadavg": "1.00 2.00 3.00 1/2 3\n",
	}
	for name, content := range files {
		if err := os.WriteFile(root+"/"+name, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	var source metrics.Source = NewReader(root)
	got, err := source.Read(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	want := metrics.RawSample{CPU: metrics.CPUCounters{Total: 90, Busy: 10}, Memory: metrics.Memory{Total: 102400, Available: 40960, SwapTotal: 51200, SwapFree: 10240}, Load: metrics.Load{One: 1, Five: 2, Fifteen: 3}}
	if got != want {
		t.Fatalf("Read() = %#v, want %#v", got, want)
	}
}
