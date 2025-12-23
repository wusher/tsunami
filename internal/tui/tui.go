package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/wusher/tsunami/internal/killer"
	"github.com/wusher/tsunami/internal/ports"
)

// Styles
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#00FFFF"))

	subtitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6272A4"))

	selectedStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#1E3A5F")).
			Foreground(lipgloss.Color("#FFFFFF")).
			Bold(true)

	systemPortStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF6B6B"))

	userPortStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#69FF94"))

	ephemeralPortStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFC58"))

	filterStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF79C6"))

	dimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6272A4"))

	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF"))

	modalStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#44475A")).
			Padding(1, 2).
			Align(lipgloss.Center)

	warningStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFC58")).
			Bold(true)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF6B6B")).
			Bold(true)

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#69FF94")).
			Bold(true)

	buttonStyle = lipgloss.NewStyle().
			Padding(0, 2).
			Margin(0, 1)

	activeButtonStyle = buttonStyle.
				Background(lipgloss.Color("#1E3A5F")).
				Foreground(lipgloss.Color("#FFFFFF")).
				Bold(true)

	inactiveButtonStyle = buttonStyle.
				Background(lipgloss.Color("#44475A")).
				Foreground(lipgloss.Color("#6272A4"))
)

// Messages
type portsScannedMsg struct {
	ports []ports.PortInfo
	err   error
}

type killResultMsg struct {
	success bool
	err     error
}

// Init initializes the TUI
func (m Model) Init() tea.Cmd {
	return scanPorts
}

// scanPorts scans for listening ports
func scanPorts() tea.Msg {
	p, err := ports.Scan()
	return portsScannedMsg{ports: p, err: err}
}

// killProcess kills the selected process
func killProcess(pid int) tea.Cmd {
	return func() tea.Msg {
		err := killer.KillWithEscalation(pid)
		return killResultMsg{success: err == nil, err: err}
	}
}

// Update handles events
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKey(msg)

	case tea.WindowSizeMsg:
		m.SetSize(msg.Width, msg.Height)
		return m, nil

	case portsScannedMsg:
		if msg.err != nil {
			m.SetError(msg.err)
			return m, nil
		}
		m.SetPorts(msg.ports)
		return m, nil

	case killResultMsg:
		if msg.err != nil {
			m.SetError(msg.err)
		} else {
			m.SetMessage(fmt.Sprintf("Killed %s (PID %d) on port %d",
				m.selected.Process, m.selected.PID, m.selected.Port))
			m.state = StateQuit
		}
		return m, tea.Quit
	}

	return m, nil
}

// handleKey handles keyboard input
func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Global quit keys
	if msg.String() == "ctrl+c" {
		m.Quit()
		return m, tea.Quit
	}

	switch m.state {
	case StateList:
		return m.handleListKey(msg)
	case StateConfirm:
		return m.handleConfirmKey(msg)
	case StateError:
		return m, tea.Quit
	}

	return m, nil
}

// handleListKey handles keys in list state
func (m Model) handleListKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		if m.filter != "" {
			m.ClearFilter()
		} else {
			m.Quit()
			return m, tea.Quit
		}
	case tea.KeyBackspace:
		m.DeleteFilterChar()
	case tea.KeyEnter:
		m.EnterConfirm()
	case tea.KeyUp:
		m.MoveUp()
	case tea.KeyDown:
		m.MoveDown()
	case tea.KeyRunes:
		// Any character typing adds to filter
		for _, r := range msg.Runes {
			m.AddFilterChar(r)
		}
	}

	return m, nil
}

// handleConfirmKey handles keys in confirm state
func (m Model) handleConfirmKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "left", "right", "h", "l", "tab":
		m.ToggleConfirm()
	case "enter":
		if p := m.Confirm(); p != nil {
			m.state = StateKilling
			return m, killProcess(p.PID)
		}
		m.CancelConfirm()
	case "esc", "n", "q":
		m.CancelConfirm()
	case "y":
		m.confirmYes = true
		if p := m.Confirm(); p != nil {
			m.state = StateKilling
			return m, killProcess(p.PID)
		}
	}

	return m, nil
}

// View renders the TUI
func (m Model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	switch m.state {
	case StateConfirm:
		return m.viewWithModal()
	case StateError:
		return m.viewError()
	case StateKilling:
		return m.viewKilling()
	case StateQuit:
		if m.message != "" {
			return successStyle.Render(m.message) + "\n"
		}
		return ""
	default:
		return m.viewList()
	}
}

// viewList renders the port list
func (m Model) viewList() string {
	var b strings.Builder

	// Header
	b.WriteString(titleStyle.Render("TSUNAMI"))
	b.WriteString("\n")
	b.WriteString(subtitleStyle.Render("Kill processes listening on ports"))
	b.WriteString("\n\n")

	// Filter (always active - just type to filter)
	filterLabel := "Filter: "
	b.WriteString(filterStyle.Render(filterLabel))
	b.WriteString(filterStyle.Render(m.filter + "_"))
	b.WriteString("\n\n")

	// Table header
	header := fmt.Sprintf("  %-8s %-10s %-20s %-15s %s",
		"PORT", "PID", "PROCESS", "USER", "PROTO")
	b.WriteString(headerStyle.Render(header))
	b.WriteString("\n")
	b.WriteString(dimStyle.Render(strings.Repeat("─", min(m.width-4, 70))))
	b.WriteString("\n")

	// Port list
	if len(m.filtered) == 0 {
		b.WriteString(dimStyle.Render("  No listening ports found"))
		b.WriteString("\n")
	} else {
		// Calculate visible range
		maxVisible := m.height - 12
		if maxVisible < 3 {
			maxVisible = 3
		}
		start := 0
		if m.cursor >= maxVisible {
			start = m.cursor - maxVisible + 1
		}
		end := min(start+maxVisible, len(m.filtered))

		for i := start; i < end; i++ {
			p := m.filtered[i]
			line := m.formatPortLine(p, i == m.cursor)
			b.WriteString(line)
			b.WriteString("\n")
		}
	}

	// Footer
	b.WriteString("\n")
	footer := dimStyle.Render("↑/↓ navigate  │  enter select  │  esc clear/quit")
	b.WriteString(footer)

	return b.String()
}

// formatPortLine formats a single port line
func (m Model) formatPortLine(p ports.PortInfo, selected bool) string {
	// Truncate process name if needed
	process := p.Process
	if len(process) > 20 {
		process = process[:17] + "..."
	}

	line := fmt.Sprintf("  %-8d %-10d %-20s %-15s %s",
		p.Port, p.PID, process, p.User, p.Proto)

	if selected {
		return selectedStyle.Render("▸" + line[1:])
	}

	// Color by port range
	portStr := fmt.Sprintf("%-8d", p.Port)
	var styledPort string
	if p.Port < 1024 {
		styledPort = systemPortStyle.Render(portStr)
	} else if p.Port <= 49151 {
		styledPort = userPortStyle.Render(portStr)
	} else {
		styledPort = ephemeralPortStyle.Render(portStr)
	}

	return fmt.Sprintf("  %s %-10d %-20s %-15s %s",
		styledPort, p.PID, process, p.User, p.Proto)
}

// viewWithModal renders the list with confirm modal overlay
func (m Model) viewWithModal() string {
	list := m.viewList()

	if m.selected == nil {
		return list
	}

	// Build modal
	var modal strings.Builder
	modal.WriteString(warningStyle.Render("Kill process?"))
	modal.WriteString("\n\n")
	modal.WriteString(fmt.Sprintf("%s (PID %d) on port %d",
		m.selected.Process, m.selected.PID, m.selected.Port))
	modal.WriteString("\n\n")

	// Buttons
	var yesBtn, noBtn string
	if m.confirmYes {
		yesBtn = activeButtonStyle.Render("[ Yes ]")
		noBtn = inactiveButtonStyle.Render("[ No ]")
	} else {
		yesBtn = inactiveButtonStyle.Render("[ Yes ]")
		noBtn = activeButtonStyle.Render("[ No ]")
	}
	modal.WriteString(yesBtn)
	modal.WriteString("  ")
	modal.WriteString(noBtn)

	styledModal := modalStyle.Render(modal.String())

	// Center modal
	modalWidth := lipgloss.Width(styledModal)
	modalHeight := lipgloss.Height(styledModal)
	x := (m.width - modalWidth) / 2
	y := (m.height - modalHeight) / 2

	return overlayModal(list, styledModal, x, y)
}

// overlayModal places a modal on top of content
func overlayModal(content, modal string, x, y int) string {
	contentLines := strings.Split(content, "\n")
	modalLines := strings.Split(modal, "\n")

	for i, modalLine := range modalLines {
		lineY := y + i
		if lineY >= 0 && lineY < len(contentLines) {
			contentLine := contentLines[lineY]
			// Simple overlay - pad content line if needed
			for len(contentLine) < x {
				contentLine += " "
			}
			if x >= 0 && x < len(contentLine) {
				contentLines[lineY] = contentLine[:x] + modalLine
			} else if x >= 0 {
				contentLines[lineY] = contentLine + strings.Repeat(" ", x-len(contentLine)) + modalLine
			}
		}
	}

	return strings.Join(contentLines, "\n")
}

// viewError renders error state
func (m Model) viewError() string {
	return errorStyle.Render("Error: "+m.err.Error()) + "\n"
}

// viewKilling renders killing state
func (m Model) viewKilling() string {
	return fmt.Sprintf("Killing %s (PID %d)...\n",
		m.selected.Process, m.selected.PID)
}

// Run starts the TUI
func Run() error {
	p := tea.NewProgram(NewModel(), tea.WithAltScreen())
	_, err := p.Run()
	return err
}
