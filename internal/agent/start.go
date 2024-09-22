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

// Update обработка событий пользователя.
func (s *StartStageType) Update(a *agent, msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return tea.Quit

		case "ctrl+r":
			a.currenStage = "registration"
			return nil

		case "ctrl+l":
			a.currenStage = "login"
			return nil
		}
	}

	return nil
}

// View отображение экрана в терминале.
func (s *StartStageType) View() string {
	str := "Войдите в систему!\n\n"
	str += "[ Ctrl+r ] - Зарегистрироваться\n"
	str += "[ Ctrl+l ] - Авторизоваться\n"

	return str
}
