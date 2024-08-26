package agent

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"strings"
)

type ErrorStageType struct {
	focusIndex int
	back       string
	error      error
}

func InitErrorModel(error error, back string) *ErrorStageType {
	return &ErrorStageType{
		error: error,
		back:  back,
	}
}

func (s *ErrorStageType) Init() tea.Cmd {
	return nil
}
func (s *ErrorStageType) Prepare(a *agent) {
}

func (m *ErrorStageType) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit
		case "enter":
			return m, func() tea.Msg {
				return openStage(m.back)
			}
		}
	}

	return m, nil
}

func (m *ErrorStageType) View() string {
	var b strings.Builder

	b.WriteString(m.error.Error())

	var focusedBackButton = focusedStyle.Render("[ Назад ]")

	buttonBack := &focusedBackButton

	fmt.Fprintf(&b, "\n%s\n\n", *buttonBack)

	return b.String()
}
