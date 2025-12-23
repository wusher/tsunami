package tui

import (
	"testing"

	"github.com/wusher/tsunami/internal/ports"
)

func TestNewModel(t *testing.T) {
	m := NewModel()

	if m.state != StateList {
		t.Errorf("NewModel().state = %v, expected StateList", m.state)
	}
	if !m.confirmYes {
		t.Error("NewModel().confirmYes should be true")
	}
	if m.cursor != 0 {
		t.Errorf("NewModel().cursor = %d, expected 0", m.cursor)
	}
	if m.filter != "" {
		t.Errorf("NewModel().filter = %q, expected empty", m.filter)
	}
}

func TestSetPorts(t *testing.T) {
	m := NewModel()
	testPorts := []ports.PortInfo{
		{Port: 3000, PID: 100, Process: "node", User: "user", Proto: "tcp"},
		{Port: 8080, PID: 200, Process: "java", User: "user", Proto: "tcp"},
	}

	m.SetPorts(testPorts)

	if len(m.ports) != 2 {
		t.Errorf("len(ports) = %d, expected 2", len(m.ports))
	}
	if len(m.filtered) != 2 {
		t.Errorf("len(filtered) = %d, expected 2", len(m.filtered))
	}
}

func TestMoveUpDown(t *testing.T) {
	m := NewModel()
	m.SetPorts([]ports.PortInfo{
		{Port: 3000, PID: 100, Process: "node", User: "user", Proto: "tcp"},
		{Port: 8080, PID: 200, Process: "java", User: "user", Proto: "tcp"},
		{Port: 5432, PID: 300, Process: "postgres", User: "user", Proto: "tcp"},
	})

	// Start at 0
	if m.cursor != 0 {
		t.Errorf("cursor = %d, expected 0", m.cursor)
	}

	// Move up at top - should stay at 0
	m.MoveUp()
	if m.cursor != 0 {
		t.Errorf("cursor = %d after MoveUp at top, expected 0", m.cursor)
	}

	// Move down
	m.MoveDown()
	if m.cursor != 1 {
		t.Errorf("cursor = %d after MoveDown, expected 1", m.cursor)
	}

	m.MoveDown()
	if m.cursor != 2 {
		t.Errorf("cursor = %d after second MoveDown, expected 2", m.cursor)
	}

	// Move down at bottom - should stay at 2
	m.MoveDown()
	if m.cursor != 2 {
		t.Errorf("cursor = %d after MoveDown at bottom, expected 2", m.cursor)
	}

	// Move back up
	m.MoveUp()
	if m.cursor != 1 {
		t.Errorf("cursor = %d after MoveUp, expected 1", m.cursor)
	}
}

func TestFilter(t *testing.T) {
	m := NewModel()
	m.SetPorts([]ports.PortInfo{
		{Port: 3000, PID: 100, Process: "node", User: "mike", Proto: "tcp"},
		{Port: 8080, PID: 200, Process: "java", User: "alice", Proto: "tcp"},
		{Port: 5432, PID: 300, Process: "postgres", User: "postgres", Proto: "tcp"},
	})

	// Filter by port number
	m.AddFilterChar('3')
	m.AddFilterChar('0')
	if len(m.filtered) != 1 {
		t.Errorf("filter '30': len(filtered) = %d, expected 1", len(m.filtered))
	}
	if m.filtered[0].Port != 3000 {
		t.Errorf("filter '30': filtered[0].Port = %d, expected 3000", m.filtered[0].Port)
	}

	// Clear filter
	m.ClearFilter()
	if m.filter != "" {
		t.Errorf("after ClearFilter, filter = %q, expected empty", m.filter)
	}
	if len(m.filtered) != 3 {
		t.Errorf("after ClearFilter: len(filtered) = %d, expected 3", len(m.filtered))
	}

	// Filter by process name (case insensitive)
	m.AddFilterChar('N')
	m.AddFilterChar('O')
	m.AddFilterChar('D')
	m.AddFilterChar('E')
	if len(m.filtered) != 1 {
		t.Errorf("filter 'NODE': len(filtered) = %d, expected 1", len(m.filtered))
	}

	// Delete filter chars
	m.DeleteFilterChar()
	if m.filter != "NOD" {
		t.Errorf("after DeleteFilterChar, filter = %q, expected 'NOD'", m.filter)
	}

	// Delete on empty filter
	m.ClearFilter()
	m.DeleteFilterChar() // Should not panic
	if m.filter != "" {
		t.Errorf("DeleteFilterChar on empty: filter = %q, expected empty", m.filter)
	}
}

func TestFilterByUser(t *testing.T) {
	m := NewModel()
	m.SetPorts([]ports.PortInfo{
		{Port: 3000, PID: 100, Process: "node", User: "mike", Proto: "tcp"},
		{Port: 8080, PID: 200, Process: "java", User: "alice", Proto: "tcp"},
	})

	m.AddFilterChar('a')
	m.AddFilterChar('l')
	m.AddFilterChar('i')
	m.AddFilterChar('c')
	m.AddFilterChar('e')

	if len(m.filtered) != 1 {
		t.Errorf("filter 'alice': len(filtered) = %d, expected 1", len(m.filtered))
	}
	if m.filtered[0].User != "alice" {
		t.Errorf("filter 'alice': filtered[0].User = %q, expected 'alice'", m.filtered[0].User)
	}
}

func TestFilterCursorReset(t *testing.T) {
	m := NewModel()
	m.SetPorts([]ports.PortInfo{
		{Port: 3000, PID: 100, Process: "node", User: "user", Proto: "tcp"},
		{Port: 8080, PID: 200, Process: "java", User: "user", Proto: "tcp"},
		{Port: 5432, PID: 300, Process: "postgres", User: "user", Proto: "tcp"},
	})

	// Move cursor to end
	m.cursor = 2

	// Filter to single result - cursor should reset
	m.AddFilterChar('n')
	m.AddFilterChar('o')
	m.AddFilterChar('d')
	m.AddFilterChar('e')

	if m.cursor != 0 {
		t.Errorf("cursor after filtering = %d, expected 0", m.cursor)
	}
}

func TestSelectedPort(t *testing.T) {
	m := NewModel()

	// Empty list
	if m.SelectedPort() != nil {
		t.Error("SelectedPort on empty list should be nil")
	}

	m.SetPorts([]ports.PortInfo{
		{Port: 3000, PID: 100, Process: "node", User: "user", Proto: "tcp"},
		{Port: 8080, PID: 200, Process: "java", User: "user", Proto: "tcp"},
	})

	// First item selected
	selected := m.SelectedPort()
	if selected == nil {
		t.Fatal("SelectedPort should not be nil")
	}
	if selected.Port != 3000 {
		t.Errorf("SelectedPort().Port = %d, expected 3000", selected.Port)
	}

	// Move to second item
	m.MoveDown()
	selected = m.SelectedPort()
	if selected.Port != 8080 {
		t.Errorf("SelectedPort().Port = %d, expected 8080", selected.Port)
	}

	// Cursor out of bounds
	m.cursor = 100
	if m.SelectedPort() != nil {
		t.Error("SelectedPort with out-of-bounds cursor should be nil")
	}
}

func TestConfirmFlow(t *testing.T) {
	m := NewModel()
	m.SetPorts([]ports.PortInfo{
		{Port: 3000, PID: 100, Process: "node", User: "user", Proto: "tcp"},
	})

	// Enter confirm state
	m.EnterConfirm()
	if m.state != StateConfirm {
		t.Errorf("state after EnterConfirm = %v, expected StateConfirm", m.state)
	}
	if m.selected == nil {
		t.Error("selected should not be nil after EnterConfirm")
	}
	if m.selected.Port != 3000 {
		t.Errorf("selected.Port = %d, expected 3000", m.selected.Port)
	}
	if !m.confirmYes {
		t.Error("confirmYes should be true after EnterConfirm")
	}

	// Toggle confirm
	m.ToggleConfirm()
	if m.confirmYes {
		t.Error("confirmYes should be false after ToggleConfirm")
	}

	m.ToggleConfirm()
	if !m.confirmYes {
		t.Error("confirmYes should be true after second ToggleConfirm")
	}

	// Confirm with Yes
	result := m.Confirm()
	if result == nil {
		t.Error("Confirm() with confirmYes should return selected")
	}

	// Confirm with No
	m.confirmYes = false
	result = m.Confirm()
	if result != nil {
		t.Error("Confirm() with confirmYes=false should return nil")
	}

	// Cancel confirm
	m.CancelConfirm()
	if m.state != StateList {
		t.Errorf("state after CancelConfirm = %v, expected StateList", m.state)
	}
	if m.selected != nil {
		t.Error("selected should be nil after CancelConfirm")
	}
}

func TestEnterConfirmEmptyList(t *testing.T) {
	m := NewModel()

	m.EnterConfirm()
	if m.state != StateList {
		t.Error("EnterConfirm on empty list should not change state")
	}
}

func TestSetError(t *testing.T) {
	m := NewModel()
	err := &testError{msg: "test error"}

	m.SetError(err)
	if m.state != StateError {
		t.Errorf("state after SetError = %v, expected StateError", m.state)
	}
	if m.err != err {
		t.Error("err not set correctly")
	}
}

func TestSetMessage(t *testing.T) {
	m := NewModel()
	m.SetMessage("test message")
	if m.message != "test message" {
		t.Errorf("message = %q, expected 'test message'", m.message)
	}
}

func TestQuit(t *testing.T) {
	m := NewModel()
	m.Quit()
	if m.state != StateQuit {
		t.Errorf("state after Quit = %v, expected StateQuit", m.state)
	}
}

func TestSetSize(t *testing.T) {
	m := NewModel()
	m.SetSize(80, 24)
	if m.width != 80 {
		t.Errorf("width = %d, expected 80", m.width)
	}
	if m.height != 24 {
		t.Errorf("height = %d, expected 24", m.height)
	}
}

func TestContains(t *testing.T) {
	tests := []struct {
		s, substr string
		expected  bool
	}{
		{"hello", "ell", true},
		{"hello", "world", false},
		{"hello", "hello", true},
		{"hello", "", true},
		{"", "", true},
		{"", "a", false},
		{"abc", "abcd", false},
	}

	for _, tt := range tests {
		result := contains(tt.s, tt.substr)
		if result != tt.expected {
			t.Errorf("contains(%q, %q) = %v, expected %v", tt.s, tt.substr, result, tt.expected)
		}
	}
}

func TestContainsIgnoreCase(t *testing.T) {
	tests := []struct {
		s, substr string
		expected  bool
	}{
		{"Hello", "ell", true},
		{"HELLO", "ell", true},
		{"hello", "ELL", true},
		{"HeLLo", "eLl", true},
		{"hello", "world", false},
	}

	for _, tt := range tests {
		result := containsIgnoreCase(tt.s, tt.substr)
		if result != tt.expected {
			t.Errorf("containsIgnoreCase(%q, %q) = %v, expected %v", tt.s, tt.substr, result, tt.expected)
		}
	}
}

func TestToLower(t *testing.T) {
	tests := []struct {
		input, expected string
	}{
		{"HELLO", "hello"},
		{"hello", "hello"},
		{"HeLLo", "hello"},
		{"123ABC", "123abc"},
		{"", ""},
	}

	for _, tt := range tests {
		result := toLower(tt.input)
		if result != tt.expected {
			t.Errorf("toLower(%q) = %q, expected %q", tt.input, result, tt.expected)
		}
	}
}

func TestFindSubstring(t *testing.T) {
	tests := []struct {
		s, substr string
		expected  int
	}{
		{"hello", "ell", 1},
		{"hello", "world", -1},
		{"hello", "hello", 0},
		{"hello", "o", 4},
		{"hello", "", 0},
		{"", "", 0},
	}

	for _, tt := range tests {
		result := findSubstring(tt.s, tt.substr)
		if result != tt.expected {
			t.Errorf("findSubstring(%q, %q) = %d, expected %d", tt.s, tt.substr, result, tt.expected)
		}
	}
}

func TestMatchesFilter(t *testing.T) {
	port := ports.PortInfo{
		Port:    3000,
		PID:     100,
		Process: "node",
		User:    "mike",
		Proto:   "tcp",
	}

	tests := []struct {
		filter   string
		expected bool
	}{
		{"3000", true},
		{"300", true},
		{"30", true},
		{"node", true},
		{"NODE", true},
		{"mike", true},
		{"MIKE", true},
		{"java", false},
		{"alice", false},
		{"999", false},
	}

	for _, tt := range tests {
		result := matchesFilter(port, tt.filter)
		if result != tt.expected {
			t.Errorf("matchesFilter(port, %q) = %v, expected %v", tt.filter, result, tt.expected)
		}
	}
}

type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}
