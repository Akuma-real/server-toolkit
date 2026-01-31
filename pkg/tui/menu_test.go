package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

func TestMenuModel_EnterOnLeafShowsUnimplementedMessage(t *testing.T) {
	m := NewMenu("t", "", []MenuItem{
		{ID: "a", Label: "A"},
	}).SetUnimplementedMessage("TODO")

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m2 := updated.(MenuModel)

	assert.Contains(t, m2.View(), "TODO")
}

func TestMenuModel_StatusClearedOnNavigation(t *testing.T) {
	m := NewMenu("t", "", []MenuItem{
		{ID: "a", Label: "A"},
		{ID: "b", Label: "B"},
	}).SetUnimplementedMessage("TODO")

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m2 := updated.(MenuModel)
	assert.Contains(t, m2.View(), "TODO")

	updated, _ = m2.Update(tea.KeyMsg{Type: tea.KeyDown})
	m3 := updated.(MenuModel)
	assert.NotContains(t, m3.View(), "TODO")
}

func TestMenuModel_ParentMenuMsgReturnsParent(t *testing.T) {
	sub := NewMenu("sub", "", []MenuItem{
		{ID: "back", Label: "Back", Action: func() tea.Cmd { return func() tea.Msg { return ParentMenuMsg{} } }},
	})
	main := NewMenu("main", "", []MenuItem{
		{ID: "sub", Label: "Sub", Submenu: &sub},
	})

	updated, _ := main.Update(tea.KeyMsg{Type: tea.KeyEnter})
	subModel := updated.(MenuModel)
	if assert.NotNil(t, subModel.parent) {
		assert.Equal(t, "sub", subModel.parent.selected)
	}

	updated, _ = subModel.Update(ParentMenuMsg{})
	backToMain := updated.(MenuModel)
	assert.Equal(t, "main", backToMain.title)
}
