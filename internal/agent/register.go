package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/KirillKhitev/goph_keeper/internal/api"
	"github.com/KirillKhitev/goph_keeper/internal/auth"
	"github.com/KirillKhitev/goph_keeper/internal/client"
	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"log"
	"os"
	"strings"
)

// openStage сообщение для смены модели.
type openStage string

// authSuccessMsg сообщение при успешной авторизации.
type authSuccessMsg struct {
	userID string
	token  string
}

// RegisterStageType модель регистрации нового пользователя.
type RegisterStageType struct {
	stageAgent
	focusIndex int
	inputs     []textinput.Model
	cursorMode cursor.Mode
	back       string
	client     *client.Client
}

// Init - заглушка для интерфейса.
func (s *RegisterStageType) Init() tea.Cmd {
	return nil
}

// Prepare подготавливает модель.
func (s *RegisterStageType) Prepare(a *agent) {
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
func (m *RegisterStageType) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

			//Зарегистрироваться
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

// process отправляет форму регистрации на сервер.
func (m *RegisterStageType) process() (tea.Model, tea.Cmd) {
	data := auth.AuthorizingData{
		UserName: m.inputs[0].Value(),
		Password: m.inputs[1].Value(),
	}

	bytes, _ := json.Marshal(data)

	ctx := context.TODO()
	response := (*m.client).Register(ctx, bytes)

	result := api.UserAuthBody{}

	if err := json.Unmarshal(response.Response, &result); err != nil {
		log.Println("Error unmarshalling response: ", err)
		return m, func() tea.Msg {
			return infoMsg{
				message:    "Не смогли распарсить ответ",
				back:       "registration",
				backButton: "Назад",
			}
		}
	}

	if response.Code != 200 {
		return m, func() tea.Msg {
			return infoMsg{
				message:    result.Msg,
				back:       "registration",
				backButton: "Назад",
			}
		}
	}

	path := "users" + string(os.PathSeparator) + result.ID
	f, err := os.Create(path)
	defer f.Close()

	if err != nil {
		log.Printf("ошибка при сохранении ключа пользователя[%s]: %s", result.ID, err)
		return m, func() tea.Msg {
			return errMsg{
				error: fmt.Errorf("ошибка при сохранении ключа, сохраните файл %s с содержимым: %s", path, result.Key),
				back:  "start",
			}
		}
	}

	f.Write([]byte(result.Key))

	return m, func() tea.Msg {
		return authSuccessMsg{
			userID: result.ID,
			token:  response.Token,
		}
	}
}

// updateInputs обрабатывает ввод текста в поля формы.
func (m *RegisterStageType) updateInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(m.inputs))

	for i := range m.inputs {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	}

	return tea.Batch(cmds...)
}

// View отображает форму в терминале.
func (m *RegisterStageType) View() string {
	var b strings.Builder

	for i := range m.inputs {
		b.WriteString(m.inputs[i].View())

		if i < len(m.inputs)-1 {
			b.WriteRune('\n')
		}
	}

	var (
		focusedSubmitButton = focusedStyle.Render("[ Зарегистрироваться ]")
		blurredSubmitButton = fmt.Sprintf("[ %s ]", blurredStyle.Render("Зарегистрироваться"))
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
