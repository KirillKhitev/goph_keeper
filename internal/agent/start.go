package agent

import (
	tea "github.com/charmbracelet/bubbletea"
)

// StartStageType стартовая модель агента.
type StartStageType struct {
	stageAgent
}

// Prepare подготовка модели.
func (s *StartStageType) Prepare(a *agent) {
}

// Init инициализация модели.
func (s *StartStageType) Init() tea.Cmd {
	return nil
}

// Update обработка событий пользователя.
func (s *StartStageType) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return s, tea.Quit

		case "ctrl+r":
			return s, func() tea.Msg {
				return openStage("registration")
			}

		case "ctrl+l":
			return s, func() tea.Msg {
				return openStage("login")
			}
		}
	}

	return s, nil
}

// View отображение экрана в терминале.
func (s *StartStageType) View() string {
	str := "Войдите в систему!\n\n"
	str += "[ Ctrl+r ] - Зарегистрироваться\n"
	str += "[ Ctrl+l ] - Авторизоваться\n"

	return str
}
