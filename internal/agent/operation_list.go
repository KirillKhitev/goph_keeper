package agent

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
)

// operation структура для операции пользователя.
type operation struct {
	title, type_record string
}

// OperationListStageType модель операции пользователя.
type OperationListStageType struct {
	stageAgent
	operations []operation
}

// Init заглушка для интерфейса.
func (s *OperationListStageType) Init() tea.Cmd {
	return nil
}

// Prepare подготавливает модель.
func (s *OperationListStageType) Prepare(a *agent) {
	s.operations = []operation{
		{
			title:       "Логин/пароль",
			type_record: "login_password",
		},
		{
			title:       "Текст",
			type_record: "text",
		},
		{
			title:       "Файл",
			type_record: "file",
		},
		{
			title:       "Банковская карта",
			type_record: "credit_card",
		},
	}
}

// Update обработка событий пользователя.
func (s *OperationListStageType) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {

		case "ctrl+c", "q":
			return s, tea.Quit

		case "up", "k":
			if s.cursor > 0 {
				s.cursor--
			}
		case "down", "j":
			if s.cursor < len(s.operations)-1 {
				s.cursor++
			}
		case "enter", " ":
			s.selected = s.cursor
			return s, func() tea.Msg {
				return openForm{
					type_record: s.operations[s.cursor].type_record,
				}
			}
		}
	}

	return s, nil
}

// View отображение списка операций.
func (s *OperationListStageType) View() string {
	str := "Что будем создавать?\n\n"

	for i, choice := range s.operations {
		cursor := " "
		if s.cursor == i {
			cursor = ">"
		}

		checked := " "
		if s.selected == i {
			checked = "x"
		}
		str += fmt.Sprintf("%s [%s] %s\n", cursor, checked, choice.title)
	}

	str += "\n"

	return str
}
