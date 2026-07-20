package tui

import (
	"context"
	"time"

	"github.com/Lotto724/herdr-system-metrics/internal/metrics"
	tea "github.com/charmbracelet/bubbletea"
)

type tickMsg struct{}

type sampleMsg struct {
	generation uint64
	raw        metrics.RawSample
	err        error
}

// Model owns collection state. Rendering is intentionally deferred to task 2.3.
type Model struct {
	source     metrics.Source
	state      *metrics.State
	snapshot   metrics.Snapshot
	paused     bool
	reading    bool
	generation uint64
	err        error
}

func NewModel(source metrics.Source, state *metrics.State) Model {
	return Model{source: source, state: state}
}

func (m Model) Init() tea.Cmd { return m.tick() }

func (m Model) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	switch message := message.(type) {
	case tea.KeyMsg:
		switch message.String() {
		case "q", "esc", "ctrl+c":
			return m, tea.Quit
		case " ":
			m.paused = !m.paused
			m.state.InvalidateCPUBaseline()
			if m.paused {
				m.generation++
			} else {
				return m, m.tick()
			}
		case "r":
			m.state.ResetHistory()
		}
	case tickMsg:
		if m.paused {
			return m, nil
		}
		tick := m.tick()
		if m.reading {
			return m, tick
		}
		m.reading = true
		return m, tea.Batch(tick, m.readCmd())
	case sampleMsg:
		m.reading = false
		if message.generation != m.generation || m.paused {
			return m, nil
		}
		if message.err != nil {
			m.err = message.err
			return m, nil
		}
		m.err = nil
		m.snapshot = m.state.Accept(message.raw)
	}
	return m, nil
}

func (m Model) View() string { return "" }

func (m Model) tick() tea.Cmd {
	return tea.Tick(time.Second, func(time.Time) tea.Msg { return tickMsg{} })
}

func (m Model) readCmd() tea.Cmd {
	generation := m.generation
	return func() tea.Msg {
		raw, err := m.source.Read(context.Background())
		return sampleMsg{generation: generation, raw: raw, err: err}
	}
}
