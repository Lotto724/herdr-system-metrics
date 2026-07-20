package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewModelSchedulesCollection(t *testing.T) {
	if command := newModel().Init(); command == nil {
		t.Fatal("newModel Init command = nil, want collection scheduling")
	}
}

func TestRunReturnsProgramError(t *testing.T) {
	for _, test := range []struct {
		name string
		err  error
	}{
		{name: "returns program failure", err: errors.New("terminal unavailable")},
		{name: "returns success", err: nil},
	} {
		t.Run(test.name, func(t *testing.T) {
			if err := run(programStub{err: test.err}); !errors.Is(err, test.err) {
				t.Fatalf("run error = %v, want %v", err, test.err)
			}
		})
	}
}

func TestPluginManifestAndREADMEDescribeTheBoundedPopup(t *testing.T) {
	root := filepath.Join("..", "..")

	t.Run("manifest declares one Linux popup with static argv", func(t *testing.T) {
		manifest, err := os.ReadFile(filepath.Join(root, "herdr-plugin.toml"))
		if err != nil {
			t.Fatal(err)
		}
		for _, want := range []string{
			`id = "herdr.system-metrics"`,
			`platforms = ["linux"]`,
			`placement = "popup"`,
			`command = ["./bin/herdr-system-metrics"]`,
		} {
			if !strings.Contains(string(manifest), want) {
				t.Errorf("manifest does not contain %q", want)
			}
		}
	})

	t.Run("README documents scope controls states and exclusions", func(t *testing.T) {
		readme, err := os.ReadFile(filepath.Join(root, "README.md"))
		if err != nil {
			t.Fatal(err)
		}
		for _, want := range []string{
			"herdr plugin link",
			"herdr plugin pane open",
			"warming up",
			"unavailable",
			"Linux/WSL",
			"VM- or namespace-scoped",
			"temperature",
			"frequency",
			"sidebar",
			"native Metrics",
		} {
			if !strings.Contains(string(readme), want) {
				t.Errorf("README does not contain %q", want)
			}
		}
	})
}

type programStub struct{ err error }

func (stub programStub) Run() (tea.Model, error) { return nil, stub.err }
