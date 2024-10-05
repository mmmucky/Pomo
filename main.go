package main

import (
	"fmt"
	"os"
	"strings"
	"strconv"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	padding  = 2
	maxWidth = 800
)

var helpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#626262")).Render

// main entry point
func main() {
    //startTime := time.Now()
	// Process arguments
	taskText := ""
	duration := 0
	var (err error)
	if len(os.Args) > 2 {
		taskText = os.Args[2]
		duration, err = strconv.Atoi(os.Args[1])
		if err != nil {
			fmt.Printf("Error: '%s' is not a valid integer.\n", os.Args[1])
			os.Exit(1)
		}
	} else {
		fmt.Println("Usage: gopomo <minutes> 'Task Text'")
		os.Exit(2)
	}

	// Create a model
	m := model{
		progress: progress.New(progress.WithDefaultGradient()),
		taskText: taskText,
		duration: duration,
	}

	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Println("Oh no!", err)
		os.Exit(1)
	}
}

type tickMsg time.Time

// model
type model struct {
	progress progress.Model
	taskText string
	duration int
    startTime time.Time
}

// Init method
func (m model) Init() tea.Cmd {
	return tickCmd()
}

// Update method
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m, tea.Quit
    // Window size changed
	case tea.WindowSizeMsg:
		m.progress.Width = msg.Width - padding*2 - 4
		if m.progress.Width > maxWidth {
			m.progress.Width = maxWidth
		}
		return m, nil

    // Time based update
	case tickMsg:
		if m.progress.Percent() == 1.0 {
			return m, tea.Quit
		}

		// Note that you can also use progress.Model.SetPercent to set the
		// percentage value explicitly, too.
		cmd := m.progress.IncrPercent(0.25)
		return m, tea.Batch(tickCmd(), cmd)

	// FrameMsg is sent when the progress bar wants to animate itself
	case progress.FrameMsg:
		progressModel, cmd := m.progress.Update(msg)
		m.progress = progressModel.(progress.Model)
		return m, cmd

	default:
		return m, nil
	}
}

// View method
func (m model) View() string {
	pad := strings.Repeat(" ", padding)
	return "\n" +
//		pad + m.duration + "\n" +
		pad + m.taskText + "\n" +
		pad + m.progress.View() + "\n\n" +
		pad + helpStyle("Press any key to quit")
}

// Tick command.  Called once during init to schedule the first tick and then again when the tick is received in the update function.
func tickCmd() tea.Cmd {
	return tea.Tick(time.Second/8, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}
