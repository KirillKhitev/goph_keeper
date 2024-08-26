package agent

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"strings"
)

type InfoStageType struct {
	message    string
	back       string
	backButton string
}

type infoMsg struct {
	message    string
	back       string
	backButton string
}

func InitInfoModel(message string, back, backButton string) *InfoStageType {
	return &InfoStageType{
		message:    message,
		back:       back,
		backButton: backButton,
	}
}

func (s *InfoStageType) Init() tea.Cmd {
	return nil
}

func (s *InfoStageType) Prepare(a *agent) {
}

func (m *InfoStageType) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

func (m *InfoStageType) View() string {
	var b strings.Builder

	b.WriteString(m.message)

	button := focusedStyle.Render(fmt.Sprintf("[ %s ]", m.backButton))

	fmt.Fprintf(&b, "\n%s\n\n", button)

	return b.String()
}
