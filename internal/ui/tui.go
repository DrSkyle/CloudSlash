package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/DrSkyle/cloudslash/internal/graph"
	"github.com/DrSkyle/cloudslash/internal/swarm" // Correct import
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Styles using "Future-Glass" palette
var (
	subtle    = lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#383838"}
	highlight = lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#00FF99"} // Neon Green
	text      = lipgloss.AdaptiveColor{Light: "#191919", Dark: "#ECECEC"}
	special   = lipgloss.AdaptiveColor{Light: "#43BF6D", Dark: "#73F59F"}
	warning   = lipgloss.AdaptiveColor{Light: "#F05D5E", Dark: "#F05D5E"}

	titleStyle = lipgloss.NewStyle().
			Foreground(highlight).
			Bold(true).
			Padding(0, 1)

	cardStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(subtle).
			Padding(1, 2).
			Margin(0, 1)
	
	listHeader = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderBottom(true).
			BorderForeground(subtle).
			MarginRight(2).
			Render
)

type tickMsg time.Time

type Model struct {
	spinner     spinner.Model
	scanning    bool
	results     []string
	err         error
	quitting    bool
	isTrial     bool // Changed to bool to match main.go

	// Engines
	Engine *swarm.Engine
	Graph  *graph.Graph

	// State
	wasteItems []string
	cursor     int
	tasksDone  int
}

// NewModel initializes the TUI model.
func NewModel(e *swarm.Engine, g *graph.Graph, isTrial bool) Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(highlight)
	
	return Model{
		spinner:     s,
		scanning:    true,
		isTrial:     isTrial,
		Engine:      e,
		Graph:       g,
		wasteItems:  []string{},
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		tea.Tick(500*time.Millisecond, func(t time.Time) tea.Msg {
			return tickMsg(t)
		}),
	)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "q" || msg.String() == "ctrl+c" {
			m.quitting = true
			return m, tea.Quit
		}
	
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case tickMsg:
		// Refresh stats from Engine
		stats := m.Engine.GetStats()
		m.tasksDone = int(stats.TasksCompleted) // Cast to int

		// Check if done (Mocking completion for UI for now)
		if stats.TasksCompleted > 10 && stats.ActiveWorkers == 0 {
			m.scanning = false
		}
		
		return m, tea.Tick(500*time.Millisecond, func(t time.Time) tea.Msg {
			return tickMsg(t)
		})
	}
	return m, nil
}

func (m Model) View() string {
	if m.err != nil {
		return m.err.Error()
	}

	if m.scanning {
		return fmt.Sprintf("\n %s Scanning AWS Infrastructure... (%d Tasks Done) \n\n %s", 
			m.spinner.View(), 
			m.tasksDone,
			helpStyle("Press q to quit"),
		)
	}

    // Mock Render of results
	s := strings.Builder{}
    
    // Header
	s.WriteString(titleStyle.Render("CLOUDSLASH PROTOCOL"))
    s.WriteString("\n")
    if m.isTrial {
        s.WriteString(lipgloss.NewStyle().Foreground(warning).Render(" [ TRIAL MODE ] "))
    } else {
        s.WriteString(lipgloss.NewStyle().Foreground(special).Render(" [ PRO MODE ACCESS ] "))
    }
    s.WriteString("\n\n")

    // Styles for icons
    warnStyle := lipgloss.NewStyle().Foreground(warning)
    specStyle := lipgloss.NewStyle().Foreground(special)

    // Results Panel (Glass Card)
    resultContent := ""
    if m.isTrial {
        resultContent = fmt.Sprintf(
            "%s %s\n%s %s",
            warnStyle.Render("✖"), "Zombie EBS: [REDACTED]",
            warnStyle.Render("✖"), "Unused NAT: [REDACTED]",
        )
    } else {
         resultContent = fmt.Sprintf(
            "%s %s\n%s %s",
            specStyle.Render("✔"), "Zombie EBS: 12 Volumes (vol-0af3...)",
            specStyle.Render("✔"), "Unused NAT: 5 Gateways (nat-0bc...)",
        )
    }

	s.WriteString(cardStyle.Render(resultContent))
	s.WriteString("\n\n")
    
	s.WriteString(helpStyle("Press q to quit"))
	return s.String()
}

func helpStyle(s string) string {
	return lipgloss.NewStyle().Foreground(subtle).Render(s)
}
