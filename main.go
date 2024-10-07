package main

import (
	"fmt"
	"os"
	"math"
	"strings"
	"strconv"
	"time"

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

	subtle    = lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#383838"}
	highlight = lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#7D56F4"}
	special   = lipgloss.AdaptiveColor{Light: "#43BF6D", Dark: "#73F59F"}

	divider = lipgloss.NewStyle().
		SetString("•").
		Padding(0, 1).
		Foreground(subtle).
		String()

	url = lipgloss.NewStyle().Foreground(special).Render

	// Title.

	titleStyle = lipgloss.NewStyle().
			MarginLeft(1).
			MarginRight(5).
			Padding(0, 1).
			Italic(true).
			Foreground(lipgloss.Color("#FFF7DB")).
			SetString("Lip Gloss")

	descStyle = lipgloss.NewStyle().MarginTop(1)

	// Dialog.

	dialogBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#874BFD")).
			Padding(1, 0).
			BorderTop(true).
			BorderLeft(true).
			BorderRight(true).
			BorderBottom(true)

	buttonStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFF7DB")).
			Background(lipgloss.Color("#888B7E")).
			Padding(0, 3).
			MarginTop(1)

	activeButtonStyle = buttonStyle.
				Foreground(lipgloss.Color("#FFF7DB")).
				Background(lipgloss.Color("#F25D94")).
				MarginRight(2).
				Underline(true)


	// Status Bar.

	statusNugget = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Padding(0, 1)

	statusBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#343433", Dark: "#C1C6B2"}).
			Background(lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#353533"})

	statusStyle = lipgloss.NewStyle().
			Inherit(statusBarStyle).
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#FF5F87")).
			Padding(0, 1).
			MarginRight(1)

	encodingStyle = statusNugget.
			Background(lipgloss.Color("#A550DF")).
			Align(lipgloss.Right)

	statusText = lipgloss.NewStyle().Inherit(statusBarStyle)

	fishCakeStyle = statusNugget.Background(lipgloss.Color("#6124DF"))

	// Page.

	docStyle = lipgloss.NewStyle().Padding(0, 0, 0, 0)
)
var helpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#626262")).Render

// main entry point
func main() {
	startTime := time.Now()
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
		startTime: startTime,
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
//	return "\n" +
////		pad + m.duration + "\n" +
//		pad + m.taskText + "\n" +
//		pad + m.progress.View() + "\n\n" +
		//okButton := activeButtonStyle.Render("Yes")
		//cancelButton := buttonStyle.Render("Maybe")
        contentWidth := int(math.Floor(float64(physicalWidth-4) * 0.8))
		text := lipgloss.NewStyle().Width(contentWidth).Align(lipgloss.Center).Render(m.taskText)
		progress := lipgloss.NewStyle().Width(contentWidth).Align(lipgloss.Center).Render(m.progress.View())
		//buttons := lipgloss.JoinHorizontal(lipgloss.Top, okButton, cancelButton)
		ui := lipgloss.JoinVertical(lipgloss.Center, text, progress)

		dialog := lipgloss.Place(physicalWidth, 9,
			lipgloss.Center, lipgloss.Center,
			dialogBoxStyle.Render(ui),
//			lipgloss.WithWhitespaceChars("猫咪"),
			lipgloss.WithWhitespaceChars(" "),
			lipgloss.WithWhitespaceForeground(subtle),
		)

		doc.WriteString(dialog )
	}
	return docStyle.Render(doc.String())
//	style := lipgloss.NewStyle().
//		SetString(m.taskText).
//		Width(50).
//		Foreground(lipgloss.Color("63"))
//	return style.Render()



//	//pad := strings.Repeat(" ", padding)
//	return "\n" +
////		pad + m.duration + "\n" +
//		pad + m.taskText + "\n" +
//		pad + m.progress.View() + "\n\n" +
//		pad + helpStyle("Press any key to quit")
}

// Tick command.  Called once during init to schedule the first tick and then again when the tick is received in the update function.
func tickCmd() tea.Cmd {
	return tea.Tick(time.Second/8, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

const (
	width = 96
	columnWidth = 30
)

