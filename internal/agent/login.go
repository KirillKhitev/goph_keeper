package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/KirillKhitev/goph_keeper/internal/api"
	"github.com/KirillKhitev/goph_keeper/internal/auth"
	"github.com/KirillKhitev/goph_keeper/internal/client"
	"github.com/KirillKhitev/goph_keeper/internal/config"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"log"
	"strings"
)

// LoginStageType модель авторизации пользователя.
type LoginStageType struct {
	RegisterStageType
	inputs []textinput.Model
	back   string
	client *client.Client
}

// Prepare подготавливает модель.
func (s *LoginStageType) Prepare(a *agent) {
	s.inputs = make([]textinput.Model, 2)
	s.back = "start"
	s.client = a.client

	var t textinput.Model

	for i := range s.inputs {
		t = textinput.New()
		t.Cursor.Style = cursorStyle
		t.CharLimit = 32

		switch i {
		case 0:
			t.Placeholder = "Логин"
			t.Focus()
			t.PromptStyle = focusedStyle
			t.TextStyle = focusedStyle
		case 1:
			t.Placeholder = "Пароль"
			t.EchoMode = textinput.EchoPassword
			t.EchoCharacter = '•'
		}

		s.inputs[i] = t
	}
}

// Update обработка событий пользователя.
func (m *LoginStageType) Update(a *agent, msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(m.inputs))

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return tea.Quit

		case "tab", "shift+tab", "enter", "up", "down":
			s := msg.String()

			// Назад
			if s == "enter" && m.focusIndex == len(m.inputs)+1 {
				a.currenStage = "start"
				return nil
			}

			//Авторизоваться
			if s == "enter" && m.focusIndex == len(m.inputs) {
				m.process(a)
				return nil
			}

			if s == "up" || s == "shift+tab" {
				m.focusIndex--
			} else {
				m.focusIndex++
			}

			if m.focusIndex > len(m.inputs)+1 {
				m.focusIndex = 0
			} else if m.focusIndex < 0 {
				m.focusIndex = len(m.inputs) + 1
			}

			for i := 0; i <= len(m.inputs)-1; i++ {
				if i == m.focusIndex {
					cmds[i] = m.inputs[i].Focus()
					m.inputs[i].PromptStyle = focusedStyle
					m.inputs[i].TextStyle = focusedStyle
					continue
				}

				m.inputs[i].Blur()
				m.inputs[i].PromptStyle = noStyle
				m.inputs[i].TextStyle = noStyle
			}
		}
	}

	cmd := m.updateInputs(msg)
	cmds = append(cmds, cmd)

	return tea.Batch(cmds...)
}

// process отправляет данные формы авторизации на сервер.
func (m *LoginStageType) process(a *agent) {
	data := auth.AuthorizingData{
		UserName: m.inputs[0].Value(),
		Password: m.inputs[1].Value(),
	}

	bytes, _ := json.Marshal(data)

	ctx := context.TODO()
	url := fmt.Sprintf("http://%s/api/user/login", config.ConfigClient.AddrServer)
	response := (*m.client).Login(ctx, url, bytes)
	result := api.UserAuthBody{}

	if err := json.Unmarshal(response.Response, &result); err != nil {
		log.Println("Error unmarshalling response: ", err)
		a.currenStage = "info"
		a.Stages[a.currenStage] = InitInfoModel("Не смогли распарсить ответ", "login", "Назад")
		return
	}

	if response.Code != 200 {
		a.currenStage = "info"
		a.Stages[a.currenStage] = InitInfoModel(result.Msg, "login", "Назад")
		return
	}

	a.currenStage = "list"
	a.userID = result.ID
	a.token = response.Token

	(*a.client).SetUserID(result.ID)

	a.Stages[a.currenStage].Prepare(a)
}

// updateInputs обработка ввода в поля формы.
func (m *LoginStageType) updateInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(m.inputs))

	for i := range m.inputs {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	}

	return tea.Batch(cmds...)
}

// View отображает форму в терминале.
func (m *LoginStageType) View() string {
	var b strings.Builder

	for i := range m.inputs {
		b.WriteString(m.inputs[i].View())
		if i < len(m.inputs)-1 {
			b.WriteRune('\n')
		}
	}

	var (
		focusedSubmitButton = focusedStyle.Render("[ Авторизоваться ]")
		blurredSubmitButton = fmt.Sprintf("[ %s ]", blurredStyle.Render("Авторизоваться"))
		focusedBackButton   = focusedStyle.Render("[ Назад ]")
		blurredBackButton   = fmt.Sprintf("[ %s ]", blurredStyle.Render("Назад"))
	)

	button := &blurredSubmitButton
	if m.focusIndex == len(m.inputs) {
		button = &focusedSubmitButton
	}

	buttonBack := &blurredBackButton

	if m.focusIndex == len(m.inputs)+1 {
		buttonBack = &focusedBackButton
	}

	fmt.Fprintf(&b, "\n%s\n%s\n\n", *button, *buttonBack)

	return b.String()
}
