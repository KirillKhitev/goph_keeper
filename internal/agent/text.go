package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/KirillKhitev/goph_keeper/internal/client"
	"github.com/KirillKhitev/goph_keeper/internal/config"
	"github.com/KirillKhitev/goph_keeper/internal/models"
	"github.com/KirillKhitev/goph_keeper/internal/mycrypto"
	"github.com/charmbracelet/bubbles/textarea"
	"log"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// TextStageType модель текстовой записи пользователя.
type TextStageType struct {
	LoginPasswordStageType
	textarea textarea.Model
}

// Prepare подготавливает модель.
func (m *TextStageType) Prepare(a *agent) {
	m.inputs = make([]textinput.Model, 2)
	m.back = "operation_list"
	m.client = a.client
	m.userID = a.userID
	m.token = a.token
	m.recordID = a.recordID

	var body string
	data := models.Data{
		ID:     m.recordID,
		UserID: m.userID,
		Type:   "text",
	}

	data, body, _ = m.getRecordDataFromServer(data, body)

	m.textarea = textarea.New()
	m.textarea.SetValue(body)

	var t textinput.Model

	for i := range m.inputs {
		t = textinput.New()
		t.Cursor.Style = cursorStyle
		t.CharLimit = 32

		switch i {
		case 0:
			t.Placeholder = "Название"
			t.Focus()
			t.PromptStyle = focusedStyle
			t.TextStyle = focusedStyle
			t.SetValue(string(data.Name))
		case 1:
			t.Placeholder = "Описание"
			t.SetValue(string(data.Description))
		}

		m.inputs[i] = t
	}
}

// save отправляет форму на сервер.
func (m *TextStageType) save() (tea.Model, tea.Cmd) {
	body, _ := mycrypto.Encrypt([]byte(m.textarea.Value()), m.userID)

	data := models.Data{
		ID:          m.recordID,
		UserID:      m.userID,
		Name:        []byte(m.inputs[0].Value()),
		Type:        "text",
		Deleted:     false,
		Description: []byte(m.inputs[1].Value()),
		Date:        time.Now(),
		Body:        body,
	}

	bytes, _ := json.Marshal(data)

	ctx := context.TODO()
	url := fmt.Sprintf("http://%s/api/data/update", config.ConfigClient.AddrServer)
	headers := map[string]string{
		"Authorization": m.token,
	}

	response := (*m.client).Send(ctx, url, headers, bytes, "PUT")

	if response.Code != 200 {
		return m, func() tea.Msg {
			return infoMsg{
				message:    string(response.Response),
				back:       "text",
				backButton: "Назад",
			}
		}
	}

	return m, func() tea.Msg {
		return openList{}
	}
}

// Update обработка событий пользователя.
func (m *TextStageType) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit

		case "esc":
			if m.textarea.Focused() {
				m.textarea.Blur()
			}

		case "ctrl+s":
			return m.save()

		case "ctrl+b":
			return m, func() tea.Msg {

				return openList{}
			}

		case "tab", "shift+tab", "up", "down":
			s := msg.String()

			if !m.textarea.Focused() {
				if s == "up" || s == "shift+tab" {
					m.focusIndex--
				} else {
					m.focusIndex++
				}

				if m.focusIndex > len(m.inputs)+2 {
					m.focusIndex = 0
				} else if m.focusIndex < 0 {
					m.focusIndex = len(m.inputs) + 2
				}
			}

			for i := 0; i <= len(m.inputs)-1; i++ {
				if i == m.focusIndex {
					cmds = append(cmds, m.inputs[i].Focus())
					m.inputs[i].PromptStyle = focusedStyle
					m.inputs[i].TextStyle = focusedStyle
					continue
				}

				m.inputs[i].Blur()
				m.inputs[i].PromptStyle = noStyle
				m.inputs[i].TextStyle = noStyle
			}

			if 2 == m.focusIndex {
				cmds = append(cmds, m.textarea.Focus())
			}

			return m, tea.Batch(cmds...)
		}
	}

	cmd := m.updateInputs(msg)
	cmds = append(cmds, cmd)

	var cmdt tea.Cmd
	m.textarea, cmdt = m.textarea.Update(msg)
	cmds = append(cmds, cmdt)

	return m, tea.Batch(cmds...)
}

// View отображение экрана в терминале.
func (m *TextStageType) View() string {
	var b strings.Builder

	for i := range m.inputs {
		b.WriteString(m.inputs[i].View())
		b.WriteRune('\n')
	}

	b.WriteString(m.textarea.View())
	b.WriteRune('\n')
	b.WriteString("\n[ Ctrl+s ] - Сохранить")
	b.WriteString("\n[ Ctrl+b ] - Назад")
	b.WriteRune('\n')

	return b.String()
}

// getRecordDataFromServer получение данных записи с сервера.
func (m *TextStageType) getRecordDataFromServer(data models.Data, body string) (models.Data, string, error) {
	var response client.APIServiceResult

	if m.getRecordID() == "" {
		return data, body, nil
	}

	ctx := context.TODO()
	url := fmt.Sprintf("http://%s/api/data/get", config.ConfigClient.AddrServer)
	headers := map[string]string{
		"Authorization": m.getToken(),
	}

	bytes, _ := json.Marshal(data)

	clientHTTP := m.getClient()
	response = (*clientHTTP).Send(ctx, url, headers, bytes, "POST")

	err := json.Unmarshal(response.Response, &data)
	if err != nil {
		log.Printf("не смогли распарсить ответ сервера: %s", err)
		return data, body, err
	}

	data.Body, err = mycrypto.Decrypt(data.Body, data.UserID)

	body = string(data.Body)

	return data, body, err
}
