package agent

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"strings"
)

// ErrorStageType - модель для отображения ошибок.
type ErrorStageType struct {
	focusIndex int
	back       string
	error      error
}

// InitErrorModel инициализирует модель.
func InitErrorModel(error error, back string) *ErrorStageType {
	return &ErrorStageType{
		error: error,
		back:  back,
	}
}

// Prepare - заглушка для интерфейса.
func (s *ErrorStageType) Prepare(a *agent) {
}

// Update - обработка событий пользователя.
func (m *ErrorStageType) Update(a *agent, msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return tea.Quit
		case "enter":
			a.currenStage = m.back
			return nil
		}
	}

	return nil
}

// View - отображает текст модели в терминале.
func (m *ErrorStageType) View() string {
	var b strings.Builder

	b.WriteString(m.error.Error())

	var focusedBackButton = focusedStyle.Render("[ Назад ]")

	buttonBack := &focusedBackButton

	fmt.Fprintf(&b, "\n%s\n\n", *buttonBack)

	return b.String()
}
