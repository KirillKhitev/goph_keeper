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

type LoginStageType struct {
	RegisterStageType
	inputs []textinput.Model
	back   string
	client *client.Client
}

func (s *LoginStageType) Init() tea.Cmd {
	return nil
}

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

func (m *LoginStageType) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit

		case "tab", "shift+tab", "enter", "up", "down":
			s := msg.String()

			// Назад
			if s == "enter" && m.focusIndex == len(m.inputs)+1 {
				return m, func() tea.Msg {
					return openStage("start")
				}
			}

			//Авторизоваться
			if s == "enter" && m.focusIndex == len(m.inputs) {
				return m.process()
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

			cmds := make([]tea.Cmd, len(m.inputs))
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

			return m, tea.Batch(cmds...)
		}
	}

	cmd := m.updateInputs(msg)

	return m, cmd
}

func (m *LoginStageType) process() (tea.Model, tea.Cmd) {
	data := auth.AuthorizingData{
		UserName: m.inputs[0].Value(),
		Password: m.inputs[1].Value(),
	}

	bytes, _ := json.Marshal(data)

	ctx := context.TODO()
	url := fmt.Sprintf("http://%s/api/user/login", config.ConfigClient.AddrServer)

	response := (*m.client).Send(ctx, url, make(map[string]string), bytes, "POST")
	result := api.UserAuthBody{}

	if err := json.Unmarshal(response.Response, &result); err != nil {
		log.Println("Error unmarshalling response: ", err)
		return m, func() tea.Msg {
			return infoMsg{
				message:    "Не смогли распарсить ответ",
				back:       "login",
				backButton: "Назад",
			}
		}
	}

	if response.Code != 200 {
		return m, func() tea.Msg {
			return infoMsg{
				message:    result.Msg,
				back:       "login",
				backButton: "Назад",
			}
		}
	}

	return m, func() tea.Msg {
		return authSuccessMsg{
			userID: result.ID,
			token:  response.Token,
		}
	}
}

func (m *LoginStageType) updateInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(m.inputs))

	for i := range m.inputs {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	}

	return tea.Batch(cmds...)
}

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
