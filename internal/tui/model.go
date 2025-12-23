package tui

import (
	"github.com/wusher/tsunami/internal/ports"
)

// State represents the current TUI state
type State int

const (
	StateList State = iota
	StateConfirm
	StateKilling
	StateError
	StateQuit
)

// Model represents the TUI state
type Model struct {
	ports      []ports.PortInfo
	filtered   []ports.PortInfo
	cursor     int
	filter     string
	state      State
	selected   *ports.PortInfo
	confirmYes bool
	err        error
	width      int
	height     int
	message    string
}

// NewModel creates a new TUI model
func NewModel() Model {
	return Model{
		state:      StateList,
		confirmYes: true, // Default to "Yes" selected
	}
}

// SetPorts sets the port list and initializes filtered view
func (m *Model) SetPorts(p []ports.PortInfo) {
	m.ports = p
	m.applyFilter()
}

// applyFilter filters ports based on current filter string
func (m *Model) applyFilter() {
	if m.filter == "" {
		m.filtered = m.ports
		return
	}

	m.filtered = nil
	for _, p := range m.ports {
		if matchesFilter(p, m.filter) {
			m.filtered = append(m.filtered, p)
		}
	}

	// Reset cursor if out of bounds
	if m.cursor >= len(m.filtered) {
		m.cursor = max(0, len(m.filtered)-1)
	}
}

// matchesFilter checks if a port matches the filter string
func matchesFilter(p ports.PortInfo, filter string) bool {
	// Match against port number, process name, or user
	portStr := string(rune('0' + p.Port%10))
	for n := p.Port / 10; n > 0; n /= 10 {
		portStr = string(rune('0'+n%10)) + portStr
	}

	return contains(portStr, filter) ||
		containsIgnoreCase(p.Process, filter) ||
		containsIgnoreCase(p.User, filter)
}

func contains(s, substr string) bool {
	return len(substr) <= len(s) && findSubstring(s, substr) != -1
}

func containsIgnoreCase(s, substr string) bool {
	return contains(toLower(s), toLower(substr))
}

func toLower(s string) string {
	b := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			b[i] = c + 32
		} else {
			b[i] = c
		}
	}
	return string(b)
}

func findSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// SelectedPort returns the currently selected port
func (m *Model) SelectedPort() *ports.PortInfo {
	if len(m.filtered) == 0 || m.cursor >= len(m.filtered) {
		return nil
	}
	return &m.filtered[m.cursor]
}

// MoveUp moves cursor up
func (m *Model) MoveUp() {
	if m.cursor > 0 {
		m.cursor--
	}
}

// MoveDown moves cursor down
func (m *Model) MoveDown() {
	if m.cursor < len(m.filtered)-1 {
		m.cursor++
	}
}

// AddFilterChar adds a character to the filter
func (m *Model) AddFilterChar(c rune) {
	m.filter += string(c)
	m.applyFilter()
}

// DeleteFilterChar removes the last character from filter
func (m *Model) DeleteFilterChar() {
	if len(m.filter) > 0 {
		m.filter = m.filter[:len(m.filter)-1]
		m.applyFilter()
	}
}

// ClearFilter clears the filter
func (m *Model) ClearFilter() {
	m.filter = ""
	m.applyFilter()
}

// EnterConfirm transitions to confirm state
func (m *Model) EnterConfirm() {
	if p := m.SelectedPort(); p != nil {
		m.selected = p
		m.state = StateConfirm
		m.confirmYes = true
	}
}

// ToggleConfirm toggles between yes and no in confirm dialog
func (m *Model) ToggleConfirm() {
	m.confirmYes = !m.confirmYes
}

// CancelConfirm returns to list state
func (m *Model) CancelConfirm() {
	m.state = StateList
	m.selected = nil
}

// Confirm confirms the action and returns selected port
func (m *Model) Confirm() *ports.PortInfo {
	if m.confirmYes && m.selected != nil {
		return m.selected
	}
	return nil
}

// SetError sets an error state
func (m *Model) SetError(err error) {
	m.err = err
	m.state = StateError
}

// SetMessage sets a status message
func (m *Model) SetMessage(msg string) {
	m.message = msg
}

// Quit transitions to quit state
func (m *Model) Quit() {
	m.state = StateQuit
}

// SetSize sets the terminal size
func (m *Model) SetSize(w, h int) {
	m.width = w
	m.height = h
}
