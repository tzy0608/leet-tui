package layout

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// Component is the interface all TUI components implement.
type Component interface {
	tea.Model
	Focus() tea.Cmd
	Blur() tea.Cmd
	IsFocused() bool
	SetSize(w, h int) tea.Cmd
	GetSize() (int, int)
	BindingKeys() []key.Binding
}

// BaseComponent provides default implementations for Component.
type BaseComponent struct {
	focused bool
	width   int
	height  int
}

func (b *BaseComponent) Focus() tea.Cmd {
	b.focused = true
	return nil
}

func (b *BaseComponent) Blur() tea.Cmd {
	b.focused = false
	return nil
}

func (b *BaseComponent) IsFocused() bool {
	return b.focused
}

func (b *BaseComponent) SetSize(w, h int) tea.Cmd {
	b.width = w
	b.height = h
	return nil
}

func (b *BaseComponent) GetSize() (int, int) {
	return b.width, b.height
}

func (b *BaseComponent) BindingKeys() []key.Binding {
	return nil
}
