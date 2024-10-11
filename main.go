package main

// TODO: Time elapsed / remaining
// TODO: Custom update interval
// TODO: task/break/task/break loop
// TODO: Bright background when time is up or when on break

import (
	"fmt"
	"os"
	"math"
	"strings"
	"time"

	"github.com/spf13/pflag"
	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
)

const (
	padding  = 2
	maxWidth = 800
)

// Style definitions.
var (
	// General.

	subtle	= lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#383838"}
	highlight = lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#7D56F4"}
	special   = lipgloss.AdaptiveColor{Light: "#43BF6D", Dark: "#73F59F"}

	// Dialog.

	dialogBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#874BFD")).
			Padding(1, 0).
			BorderTop(true).
			BorderLeft(true).
			BorderRight(true).
			BorderBottom(true)

	// Page.

	docStyle = lipgloss.NewStyle().Padding(0, 0, 0, 0)
)

// main entry point
func main() {
	// Process arguments
	var (
		taskText string
		duration  int
	)

	pflag.StringVarP(&taskText, "text", "t", "Task", "Task text")
	pflag.IntVarP(&duration, "duration", "d", 10, "Timer duration(minutes)")

	// Parse arguments
	pflag.Parse()

	// Create a model
	m := model{
		progress: progress.New(progress.WithDefaultGradient()),
		taskText: taskText,
		duration: duration,
		startTime: time.Now(),
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
	currentTime := time.Now()
	elapsedSeconds := currentTime.Sub(m.startTime).Seconds()
	totalSeconds := 60.0 * m.duration
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m, tea.Quit
	// Window size changed
	case tea.WindowSizeMsg:
		physicalWidth, _, _ := term.GetSize(int(os.Stdout.Fd()))
		contentWidth := int(math.Floor(float64(physicalWidth-4) * 0.8))
		m.progress.Width = msg.Width - padding*2 - 4
		m.progress.Width = contentWidth
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
		cmd := m.progress.SetPercent(elapsedSeconds / float64(totalSeconds))
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
	physicalWidth, _, _ := term.GetSize(int(os.Stdout.Fd()))
	doc := strings.Builder{}

	// Dialog
	{
		currentTime := time.Now()
		elapsedSeconds := currentTime.Sub(m.startTime)
		elapsedText := "(elapsed: " + elapsedSeconds.Truncate(time.Second).String() + ")"
		taskText := m.taskText + " " + elapsedText

		contentWidth := int(math.Floor(float64(physicalWidth-4) * 0.8))
		text := lipgloss.NewStyle().Width(contentWidth).Align(lipgloss.Center).Render(taskText)
		progress := lipgloss.NewStyle().Width(contentWidth).Align(lipgloss.Center).Render(m.progress.View())

		ui := lipgloss.JoinVertical(lipgloss.Center, text, progress)

		dialog := lipgloss.Place(physicalWidth, 9,
			lipgloss.Center, lipgloss.Center,
			dialogBoxStyle.Render(ui),
			lipgloss.WithWhitespaceChars(" "),
			lipgloss.WithWhitespaceForeground(subtle),
		)

		doc.WriteString(dialog )
	}
	return docStyle.Render(doc.String())
}

// Tick command.  Called once during init to schedule the first tick and then again when the tick is received in the update function.
func tickCmd() tea.Cmd {
	return tea.Tick(time.Second/8, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

