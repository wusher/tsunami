package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/wusher/tsunami/internal/ports"
)

func TestInit(t *testing.T) {
	m := NewModel()
	cmd := m.Init()
	if cmd == nil {
		t.Error("Init() should return a command")
	}
}

func TestUpdateWindowSize(t *testing.T) {
	m := NewModel()
	msg := tea.WindowSizeMsg{Width: 100, Height: 50}

	newModel, cmd := m.Update(msg)
	updated := newModel.(Model)

	if cmd != nil {
		t.Error("WindowSizeMsg should not return a command")
	}
	if updated.width != 100 {
		t.Errorf("width = %d, expected 100", updated.width)
	}
	if updated.height != 50 {
		t.Errorf("height = %d, expected 50", updated.height)
	}
}

func TestUpdatePortsScanned(t *testing.T) {
	m := NewModel()
	testPorts := []ports.PortInfo{
		{Port: 3000, PID: 100, Process: "node", User: "user", Proto: "tcp"},
	}
	msg := portsScannedMsg{ports: testPorts, err: nil}

	newModel, _ := m.Update(msg)
	updated := newModel.(Model)

	if len(updated.ports) != 1 {
		t.Errorf("len(ports) = %d, expected 1", len(updated.ports))
	}
}

func TestUpdatePortsScannedError(t *testing.T) {
	m := NewModel()
	msg := portsScannedMsg{ports: nil, err: &testError{msg: "scan failed"}}

	newModel, _ := m.Update(msg)
	updated := newModel.(Model)

	if updated.state != StateError {
		t.Errorf("state = %v, expected StateError", updated.state)
	}
}

func TestUpdateCtrlC(t *testing.T) {
	m := NewModel()
	msg := tea.KeyMsg{Type: tea.KeyCtrlC}

	newModel, cmd := m.Update(msg)
	updated := newModel.(Model)

	if updated.state != StateQuit {
		t.Errorf("state = %v, expected StateQuit", updated.state)
	}
	if cmd == nil {
		t.Error("Ctrl+C should return tea.Quit command")
	}
}

func TestHandleListKeyNavigation(t *testing.T) {
	m := NewModel()
	m.SetPorts([]ports.PortInfo{
		{Port: 3000, PID: 100, Process: "node", User: "user", Proto: "tcp"},
		{Port: 8080, PID: 200, Process: "java", User: "user", Proto: "tcp"},
	})

	// Test down arrow
	msg := tea.KeyMsg{Type: tea.KeyDown}
	newModel, _ := m.Update(msg)
	updated := newModel.(Model)
	if updated.cursor != 1 {
		t.Errorf("cursor after KeyDown = %d, expected 1", updated.cursor)
	}

	// Test up arrow
	msg = tea.KeyMsg{Type: tea.KeyUp}
	newModel, _ = updated.Update(msg)
	updated = newModel.(Model)
	if updated.cursor != 0 {
		t.Errorf("cursor after KeyUp = %d, expected 0", updated.cursor)
	}
}

func TestHandleListKeyEscQuit(t *testing.T) {
	m := NewModel()
	m.SetPorts([]ports.PortInfo{
		{Port: 3000, PID: 100, Process: "node", User: "user", Proto: "tcp"},
	})

	// Esc with no filter should quit
	msg := tea.KeyMsg{Type: tea.KeyEsc}
	newModel, cmd := m.Update(msg)
	updated := newModel.(Model)

	if updated.state != StateQuit {
		t.Errorf("state after Esc = %v, expected StateQuit", updated.state)
	}
	if cmd == nil {
		t.Error("Esc should return quit command")
	}
}

func TestHandleListKeyEscClearFilter(t *testing.T) {
	m := NewModel()
	m.SetPorts([]ports.PortInfo{
		{Port: 3000, PID: 100, Process: "node", User: "user", Proto: "tcp"},
	})
	m.filter = "test"

	// Esc with filter should clear filter
	msg := tea.KeyMsg{Type: tea.KeyEsc}
	newModel, cmd := m.Update(msg)
	updated := newModel.(Model)

	if updated.filter != "" {
		t.Errorf("filter after Esc = %q, expected empty", updated.filter)
	}
	if updated.state != StateList {
		t.Errorf("state after Esc with filter = %v, expected StateList", updated.state)
	}
	if cmd != nil {
		t.Error("Esc with filter should not return quit command")
	}
}

func TestHandleListKeyEnter(t *testing.T) {
	m := NewModel()
	m.SetPorts([]ports.PortInfo{
		{Port: 3000, PID: 100, Process: "node", User: "user", Proto: "tcp"},
	})

	msg := tea.KeyMsg{Type: tea.KeyEnter}
	newModel, _ := m.Update(msg)
	updated := newModel.(Model)

	if updated.state != StateConfirm {
		t.Errorf("state after Enter = %v, expected StateConfirm", updated.state)
	}
}

func TestHandleListKeyTyping(t *testing.T) {
	m := NewModel()
	m.SetPorts([]ports.PortInfo{
		{Port: 3000, PID: 100, Process: "node", User: "user", Proto: "tcp"},
	})

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a', 'b'}}
	newModel, _ := m.Update(msg)
	updated := newModel.(Model)

	if updated.filter != "ab" {
		t.Errorf("filter after typing = %q, expected 'ab'", updated.filter)
	}
}

func TestHandleListKeyBackspace(t *testing.T) {
	m := NewModel()
	m.filter = "abc"

	msg := tea.KeyMsg{Type: tea.KeyBackspace}
	newModel, _ := m.Update(msg)
	updated := newModel.(Model)

	if updated.filter != "ab" {
		t.Errorf("filter after backspace = %q, expected 'ab'", updated.filter)
	}
}

func TestHandleConfirmKeyToggle(t *testing.T) {
	m := NewModel()
	m.SetPorts([]ports.PortInfo{
		{Port: 3000, PID: 100, Process: "node", User: "user", Proto: "tcp"},
	})
	m.EnterConfirm()

	// Left/right should toggle
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}}
	newModel, _ := m.Update(msg)
	updated := newModel.(Model)

	if updated.confirmYes {
		t.Error("confirmYes should be false after 'l'")
	}

	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}}
	newModel, _ = updated.Update(msg)
	updated = newModel.(Model)

	if !updated.confirmYes {
		t.Error("confirmYes should be true after 'h'")
	}
}

func TestHandleConfirmKeyEsc(t *testing.T) {
	m := NewModel()
	m.SetPorts([]ports.PortInfo{
		{Port: 3000, PID: 100, Process: "node", User: "user", Proto: "tcp"},
	})
	m.EnterConfirm()

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}}
	newModel, _ := m.Update(msg)
	updated := newModel.(Model)

	if updated.state != StateList {
		t.Errorf("state after 'n' in confirm = %v, expected StateList", updated.state)
	}
}

func TestHandleConfirmKeyY(t *testing.T) {
	m := NewModel()
	m.SetPorts([]ports.PortInfo{
		{Port: 3000, PID: 100, Process: "node", User: "user", Proto: "tcp"},
	})
	m.EnterConfirm()
	m.confirmYes = false // Set to No

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}}
	newModel, cmd := m.Update(msg)
	updated := newModel.(Model)

	if updated.state != StateKilling {
		t.Errorf("state after 'y' = %v, expected StateKilling", updated.state)
	}
	if cmd == nil {
		t.Error("'y' should return kill command")
	}
}

func TestViewLoading(t *testing.T) {
	m := NewModel()
	// Width is 0 initially
	view := m.View()
	if view != "Loading..." {
		t.Errorf("View with width=0: %q, expected 'Loading...'", view)
	}
}

func TestViewList(t *testing.T) {
	m := NewModel()
	m.SetSize(80, 24)
	m.SetPorts([]ports.PortInfo{
		{Port: 3000, PID: 100, Process: "node", User: "user", Proto: "tcp"},
	})

	view := m.View()

	if !strings.Contains(view, "TSUNAMI") {
		t.Error("View should contain 'TSUNAMI'")
	}
	if !strings.Contains(view, "Filter:") {
		t.Error("View should contain 'Filter:'")
	}
	if !strings.Contains(view, "PORT") {
		t.Error("View should contain 'PORT' header")
	}
	if !strings.Contains(view, "3000") {
		t.Error("View should contain port 3000")
	}
	if !strings.Contains(view, "node") {
		t.Error("View should contain process 'node'")
	}
}

func TestViewListEmpty(t *testing.T) {
	m := NewModel()
	m.SetSize(80, 24)
	m.SetPorts([]ports.PortInfo{})

	view := m.View()

	if !strings.Contains(view, "No listening ports found") {
		t.Error("View should show empty message")
	}
}

func TestViewConfirm(t *testing.T) {
	m := NewModel()
	m.SetSize(80, 24)
	m.SetPorts([]ports.PortInfo{
		{Port: 3000, PID: 100, Process: "node", User: "user", Proto: "tcp"},
	})
	m.EnterConfirm()

	// Verify state is set correctly
	if m.state != StateConfirm {
		t.Errorf("state = %v, expected StateConfirm", m.state)
	}
	if m.selected == nil {
		t.Error("selected should not be nil in confirm state")
	}

	view := m.View()

	// Confirm view should show the kill confirmation
	if !strings.Contains(view, "KILL PROCESS?") {
		t.Error("Confirm view should contain 'KILL PROCESS?'")
	}
	if !strings.Contains(view, "node") {
		t.Error("Confirm view should contain process name")
	}
	if !strings.Contains(view, "3000") {
		t.Error("Confirm view should contain port number")
	}
}

func TestViewError(t *testing.T) {
	m := NewModel()
	m.SetSize(80, 24)
	m.SetError(&testError{msg: "test error"})

	view := m.View()

	if !strings.Contains(view, "Error:") {
		t.Error("Error view should contain 'Error:'")
	}
	if !strings.Contains(view, "test error") {
		t.Error("Error view should contain error message")
	}
}

func TestViewKilling(t *testing.T) {
	m := NewModel()
	m.SetSize(80, 24)
	m.SetPorts([]ports.PortInfo{
		{Port: 3000, PID: 100, Process: "node", User: "user", Proto: "tcp"},
	})
	m.EnterConfirm()
	m.state = StateKilling

	view := m.View()

	if !strings.Contains(view, "Killing") {
		t.Error("Killing view should contain 'Killing'")
	}
}

func TestViewQuitWithMessage(t *testing.T) {
	m := NewModel()
	m.SetSize(80, 24)
	m.state = StateQuit
	m.message = "Process killed"

	view := m.View()

	if !strings.Contains(view, "Process killed") {
		t.Error("Quit view should contain message")
	}
}

func TestViewQuitNoMessage(t *testing.T) {
	m := NewModel()
	m.SetSize(80, 24)
	m.state = StateQuit

	view := m.View()

	if view != "" {
		t.Errorf("Quit view with no message should be empty, got %q", view)
	}
}

func TestFormatPortLine(t *testing.T) {
	m := NewModel()
	m.SetSize(80, 24)

	tests := []struct {
		port     ports.PortInfo
		selected bool
	}{
		{ports.PortInfo{Port: 80, PID: 100, Process: "nginx", User: "root", Proto: "tcp"}, false},
		{ports.PortInfo{Port: 3000, PID: 200, Process: "node", User: "user", Proto: "tcp"}, true},
		{ports.PortInfo{Port: 50000, PID: 300, Process: "app", User: "user", Proto: "tcp6"}, false},
	}

	for _, tt := range tests {
		line := m.formatPortLine(tt.port, tt.selected)
		if !strings.Contains(line, tt.port.Process) {
			t.Errorf("formatPortLine should contain process name %q", tt.port.Process)
		}
	}
}

func TestFormatPortLineLongProcess(t *testing.T) {
	m := NewModel()
	m.SetSize(80, 24)

	port := ports.PortInfo{
		Port:    3000,
		PID:     100,
		Process: "very-long-process-name-that-exceeds-limit",
		User:    "user",
		Proto:   "tcp",
	}

	line := m.formatPortLine(port, false)

	if strings.Contains(line, "very-long-process-name-that-exceeds-limit") {
		t.Error("Long process name should be truncated")
	}
	if !strings.Contains(line, "...") {
		t.Error("Truncated process name should end with '...'")
	}
}

func TestUpdateKillResult(t *testing.T) {
	m := NewModel()
	m.SetPorts([]ports.PortInfo{
		{Port: 3000, PID: 100, Process: "node", User: "user", Proto: "tcp"},
	})
	m.EnterConfirm()
	m.state = StateKilling

	// Success
	msg := killResultMsg{success: true, err: nil}
	newModel, cmd := m.Update(msg)
	updated := newModel.(Model)

	if updated.state != StateQuit {
		t.Errorf("state after successful kill = %v, expected StateQuit", updated.state)
	}
	if updated.message == "" {
		t.Error("message should be set after successful kill")
	}
	if cmd == nil {
		t.Error("successful kill should return quit command")
	}
}

func TestUpdateKillResultError(t *testing.T) {
	m := NewModel()
	m.SetPorts([]ports.PortInfo{
		{Port: 3000, PID: 100, Process: "node", User: "user", Proto: "tcp"},
	})
	m.EnterConfirm()
	m.state = StateKilling

	msg := killResultMsg{success: false, err: &testError{msg: "permission denied"}}
	newModel, _ := m.Update(msg)
	updated := newModel.(Model)

	if updated.state != StateError {
		t.Errorf("state after failed kill = %v, expected StateError", updated.state)
	}
}

func TestCenterText(t *testing.T) {
	m := NewModel()
	m.SetSize(80, 24)

	text := "Hello"
	centered := m.centerText(text)

	// Should have padding before the text
	if !strings.HasPrefix(centered, " ") {
		t.Error("centerText should add padding")
	}
	if !strings.Contains(centered, "Hello") {
		t.Error("centerText should contain the text")
	}
}

func TestCenterTextNarrowTerminal(t *testing.T) {
	m := NewModel()
	m.SetSize(10, 24) // Narrow terminal

	text := "Very Long Text Here"
	centered := m.centerText(text)

	// Should not have negative padding
	if !strings.Contains(centered, text) {
		t.Error("centerText should contain the text even in narrow terminal")
	}
}

func TestHandleConfirmEnterWithNo(t *testing.T) {
	m := NewModel()
	m.SetPorts([]ports.PortInfo{
		{Port: 3000, PID: 100, Process: "node", User: "user", Proto: "tcp"},
	})
	m.EnterConfirm()
	m.confirmYes = false

	msg := tea.KeyMsg{Type: tea.KeyEnter}
	newModel, _ := m.Update(msg)
	updated := newModel.(Model)

	if updated.state != StateList {
		t.Errorf("state after Enter with No = %v, expected StateList", updated.state)
	}
}

func TestHandleErrorState(t *testing.T) {
	m := NewModel()
	m.SetError(&testError{msg: "test"})

	msg := tea.KeyMsg{Type: tea.KeyEnter}
	_, cmd := m.Update(msg)

	if cmd == nil {
		t.Error("any key in error state should quit")
	}
}

func TestHandleConfirmTab(t *testing.T) {
	m := NewModel()
	m.SetPorts([]ports.PortInfo{
		{Port: 3000, PID: 100, Process: "node", User: "user", Proto: "tcp"},
	})
	m.EnterConfirm()

	// Tab should toggle
	msg := tea.KeyMsg{Type: tea.KeyTab}
	newModel, _ := m.Update(msg)
	updated := newModel.(Model)

	if updated.confirmYes {
		t.Error("confirmYes should be false after Tab")
	}
}

func TestHandleConfirmLeftRight(t *testing.T) {
	m := NewModel()
	m.SetPorts([]ports.PortInfo{
		{Port: 3000, PID: 100, Process: "node", User: "user", Proto: "tcp"},
	})
	m.EnterConfirm()

	// Left key
	msg := tea.KeyMsg{Type: tea.KeyLeft}
	newModel, _ := m.Update(msg)
	updated := newModel.(Model)

	if updated.confirmYes {
		t.Error("confirmYes should be false after Left")
	}

	// Right key to toggle back
	msg = tea.KeyMsg{Type: tea.KeyRight}
	newModel, _ = updated.Update(msg)
	updated = newModel.(Model)

	if !updated.confirmYes {
		t.Error("confirmYes should be true after Right")
	}
}

func TestHandleConfirmQ(t *testing.T) {
	m := NewModel()
	m.SetPorts([]ports.PortInfo{
		{Port: 3000, PID: 100, Process: "node", User: "user", Proto: "tcp"},
	})
	m.EnterConfirm()

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}
	newModel, _ := m.Update(msg)
	updated := newModel.(Model)

	if updated.state != StateList {
		t.Errorf("state after 'q' in confirm = %v, expected StateList", updated.state)
	}
}

func TestHandleConfirmEsc(t *testing.T) {
	m := NewModel()
	m.SetPorts([]ports.PortInfo{
		{Port: 3000, PID: 100, Process: "node", User: "user", Proto: "tcp"},
	})
	m.EnterConfirm()

	msg := tea.KeyMsg{Type: tea.KeyEsc}
	newModel, _ := m.Update(msg)
	updated := newModel.(Model)

	if updated.state != StateList {
		t.Errorf("state after Esc in confirm = %v, expected StateList", updated.state)
	}
}

func TestViewWithModalNoSelected(t *testing.T) {
	m := NewModel()
	m.SetSize(80, 24)
	m.state = StateConfirm
	m.selected = nil // Force nil selected

	view := m.View()

	// Should just show the list without modal
	if !strings.Contains(view, "TSUNAMI") {
		t.Error("View should contain title")
	}
}

func TestScanPortsMsg(t *testing.T) {
	// Test the scanPorts function indirectly through Init
	m := NewModel()
	cmd := m.Init()

	if cmd == nil {
		t.Error("Init should return scanPorts command")
	}
}

func TestViewScrolling(t *testing.T) {
	m := NewModel()
	m.SetSize(80, 15) // Small height to trigger scrolling

	// Add many ports
	var portList []ports.PortInfo
	for i := 0; i < 20; i++ {
		portList = append(portList, ports.PortInfo{
			Port:    3000 + i,
			PID:     100 + i,
			Process: "proc",
			User:    "user",
			Proto:   "tcp",
		})
	}
	m.SetPorts(portList)

	// Move cursor to bottom
	for i := 0; i < 15; i++ {
		m.MoveDown()
	}

	view := m.View()

	// Should still render without error
	if !strings.Contains(view, "TSUNAMI") {
		t.Error("View should contain title even when scrolled")
	}
}

func TestFormatPortLineSystemPort(t *testing.T) {
	m := NewModel()
	m.SetSize(80, 24)

	// System port (< 1024)
	port := ports.PortInfo{Port: 80, PID: 100, Process: "nginx", User: "root", Proto: "tcp"}
	line := m.formatPortLine(port, false)

	if !strings.Contains(line, "80") {
		t.Error("Line should contain port number")
	}
}

func TestFormatPortLineEphemeralPort(t *testing.T) {
	m := NewModel()
	m.SetSize(80, 24)

	// Ephemeral port (> 49151)
	port := ports.PortInfo{Port: 50000, PID: 100, Process: "app", User: "user", Proto: "tcp"}
	line := m.formatPortLine(port, false)

	if !strings.Contains(line, "50000") {
		t.Error("Line should contain port number")
	}
}
