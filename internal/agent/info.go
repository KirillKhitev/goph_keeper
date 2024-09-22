package agent

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"strings"
)

// InfoStageType модель для информационного сообщения пользователю.
type InfoStageType struct {
	message    string
	back       string
	backButton string
}

// infoMsg сообщение для вывода экрана с информацией.
type infoMsg struct {
	message    string
	back       string
	backButton string
}

// InitInfoModel инициализация модели.
func InitInfoModel(message string, back, backButton string) *InfoStageType {
	return &InfoStageType{
		message:    message,
		back:       back,
		backButton: backButton,
	}
}

// Prepare - заглушка для интерфейса.
func (s *InfoStageType) Prepare(a *agent) {
}

// Update обрабатывает события пользователя.
func (m *InfoStageType) Update(a *agent, msg tea.Msg) tea.Cmd {
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

// View отображает экран в терминале.
func (m *InfoStageType) View() string {
	var b strings.Builder

	b.WriteString(m.message)

	button := focusedStyle.Render(fmt.Sprintf("[ %s ]", m.backButton))

	fmt.Fprintf(&b, "\n%s\n\n", button)

	return b.String()
}
