package main

// TODO: Ask for confirmation on exit to prevent accidental timer stop
// TODO: Erase timer on exit so you don't think you are looking at an active timer when it is stopped
// TODO: Time remaining
// TODO: Custom update interval
// TODO: task/break/task/break loop
// TODO: Bright background when time is up or when on break
// TODO: Timer Pause
// TODO: Option for less distracting colors
// TODO: timer end audible alarm / notification hook

import (
	"fmt"
	"math"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/pflag"
	"golang.org/x/term"
)

const (
	padding  = 2
	maxWidth = 800
)

// Style definitions.
var (
	// General.

	subtle    = lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#383838"}
	highlight = lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#7D56F4"}
	special   = lipgloss.AdaptiveColor{Light: "#43BF6D", Dark: "#73F59F"}

	// Dialog.

	dialogBoxStyle = lipgloss.NewStyle().
		//Background(lipgloss.Color("#3498db")). // Change this to whatever background color you like
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#874BFD")).
		Padding(0, 0).
		BorderTop(true).
		BorderLeft(true).
		BorderRight(true).
		BorderBottom(true)

	dialogBoxStyleGrey = lipgloss.NewStyle().
		//Background(lipgloss.Color("#3498db")). // Change this to whatever background color you like
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#303030")).
		Padding(0, 0).
		BorderTop(true).
		BorderLeft(true).
		BorderRight(true).
		BorderBottom(true)
)

// main entry point
func main() {
	// Process arguments
	var (
		mode     string
		modeInt  Mode
		taskText string
		duration int
	)

	pflag.StringVarP(&mode, "mode", "m", "single", "Mode: single, pomodoro")
	pflag.StringVarP(&taskText, "text", "t", "Task", "Task text")
	pflag.IntVarP(&duration, "duration", "d", 10, "Timer duration(minutes)")

	// Parse arguments
	pflag.Parse()

	switch mode {
	case "single":
		modeInt = SingleMode
	case "pomodoro":
		modeInt = PomodoroMode
	default:
		modeInt = SingleMode
	}
	// Create a model
	m := model{
		//progress: progress.New(progress.WithDefaultGradient()),
		progress:  progress.New(progress.WithGradient("#5A56E0", "#EE6FF8")),
		taskText:  taskText,
		duration:  duration,
		startTime: time.Now(),
		state:     TaskState,
		mode:      modeInt,
	}

	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Println("Oh no!", err)
		os.Exit(1)
	}
}

type tickMsg time.Time

type State int

const (
	TaskState  State = iota // 0
	BreakState              // 1
)

// A type to track what type of timer we want.
type Mode int

const (
	SingleMode   Mode = iota // One timer, then exit
	PomodoroMode             // Alternate between task and break
)

var modeDescriptions = [...]string{
	"Single timer",
	"Pomodoro timer",
}

// model
type model struct {
	progress  progress.Model
	taskText  string
	duration  int
	startTime time.Time
	state     State
	mode      Mode
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
		physicalWidth, _, _ := term.GetSize(int(os.Stdout.Fd()))
		contentWidth := int(math.Floor(float64(physicalWidth-4) * 0.8))
		m.progress.Width = contentWidth
		if m.progress.Width > maxWidth {
			m.progress.Width = maxWidth
		}
		return m, nil

	// Time based update
	case tickMsg:
		if m.progress.Percent() == 1.0 {
			if m.mode == PomodoroMode {
				return m, tea.Quit
			} else {
				m.startTime = time.Now()
			}
		}
		currentTime := time.Now()
		elapsedSeconds := currentTime.Sub(m.startTime).Seconds()
		totalSeconds := 60.0 * m.duration
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
	w, h, _ := term.GetSize(int(os.Stdout.Fd()))

	//	docStyle = lipgloss.NewStyle().Padding(0, 0, 0, 0)
	docStyle := lipgloss.NewStyle().
		Width(w).
		Height(h).
		Align(lipgloss.Center)
		//		Background(lipgloss.Color("#3498db")). // Change this to whatever background color you like
		//		Foreground(lipgloss.Color("#ffffff")) // Change this to whatever text color you like

	physicalWidth, _, _ := term.GetSize(int(os.Stdout.Fd()))

	doc := strings.Builder{}

	// Dialog
	{
		currentTime := time.Now()
		elapsedSeconds := currentTime.Sub(m.startTime)
		elapsedText := "(elapsed: " + elapsedSeconds.Truncate(time.Second).String() + ")"
		taskText := m.taskText + " " + elapsedText
		modeText := modeDescriptions[m.mode]
		//var modeText string
		//modeText = m.mode ? m.mode == SingleMode : "Single timer"

		contentWidth := int(math.Floor(float64(physicalWidth-4) * 0.8)) // Use 80% of the term, leaving 4 characters for dialog frame and padding

		infoText := lipgloss.NewStyle().Width(contentWidth).Align(lipgloss.Center).Render(modeText)
		text := lipgloss.NewStyle().Width(contentWidth).Align(lipgloss.Center).Render(taskText)
		progress := lipgloss.NewStyle().Width(contentWidth).Align(lipgloss.Center).Render(m.progress.View())
		//text = lipgloss.NewStyle().Render(taskText)
		//progress = lipgloss.NewStyle().Render(m.progress.View())

		ui := lipgloss.JoinVertical(lipgloss.Center, infoText, text, progress)

		dialog := lipgloss.Place(w, h,
			lipgloss.Center, lipgloss.Center,
			dialogBoxStyleGrey.Render(ui),
			lipgloss.WithWhitespaceChars(" "),
			lipgloss.WithWhitespaceForeground(subtle),
		)

		doc.WriteString(dialog)
	}
	return docStyle.Render(doc.String())
}

// Tick command.  Called once during init to schedule the first tick and then again when the tick is received in the update function.
func tickCmd() tea.Cmd {
	return tea.Tick(time.Second*2, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}
