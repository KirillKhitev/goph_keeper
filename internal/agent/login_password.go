package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/KirillKhitev/goph_keeper/internal/config"
	"github.com/KirillKhitev/goph_keeper/internal/models"
	"github.com/KirillKhitev/goph_keeper/internal/mycrypto"
	"log"
	"time"

	"github.com/KirillKhitev/goph_keeper/internal/client"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"strings"
)

// FormStageType интерфейс формы записи пользователя.
type FormStageType interface {
	getClient() *client.Client
	getToken() string
	getRecordID() string
}

// LoginPasswordStageType модель записи логин-пароль.
type LoginPasswordStageType struct {
	stageAgent
	focusIndex int
	inputs     []textinput.Model
	back       string
	client     *client.Client
	recordID   string
	token      string
}

// getClient возвращает http-клиент.
func (m *LoginPasswordStageType) getClient() *client.Client {
	return m.client
}

// getToken возвращает авторизационный токен.
func (m *LoginPasswordStageType) getToken() string {
	return m.token
}

// getRecordID возвращает ID записи пользователя.
func (m *LoginPasswordStageType) getRecordID() string {
	return m.recordID
}

// Prepare подготавливает модель данными.
func (m *LoginPasswordStageType) Prepare(a *agent) {
	m.inputs = make([]textinput.Model, 4)
	m.back = "operation_list"
	m.client = a.client
	m.userID = a.userID
	m.token = a.token
	m.recordID = a.recordID

	var body models.LoginBody
	data := models.Data{
		ID:     m.recordID,
		UserID: m.userID,
		Type:   "login_password",
	}

	data, body, _ = getRecordDataFromServer[models.LoginBody](m, data, body)

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
			t.Placeholder = "Логин"
			t.SetValue(body.Login)
		case 2:
			t.Placeholder = "Пароль"
			t.SetValue(body.Password)
		case 3:
			t.Placeholder = "Описание"
			t.SetValue(string(data.Description))
		}

		m.inputs[i] = t
	}

}

// getRecordDataFromServer метод получения данных записи с сервера.
func getRecordDataFromServer[T any](m FormStageType, data models.Data, body T) (models.Data, T, error) {
	var response client.APIServiceResult

	if m.getRecordID() == "" {
		return data, body, nil
	}

	ctx := context.TODO()
	headers := map[string]string{
		"Authorization": m.getToken(),
	}

	bytes, _ := json.Marshal(data)

	clientHTTP := m.getClient()
	url := fmt.Sprintf("http://%s/api/data/get", config.ConfigClient.AddrServer)
	response = (*clientHTTP).Get(ctx, url, headers, bytes)

	err := json.Unmarshal(response.Response, &data)
	if err != nil {
		log.Printf("не смогли распарсить ответ сервера: %s", err)
		return data, body, err
	}

	data.Body, err = mycrypto.Decrypt(data.Body, data.UserID)
	if err != nil {
		log.Printf("не смогли расшифровать тело записи: %s", err)
	}

	if data.Type == "file" {
		return data, body, err
	}

	err = json.Unmarshal(data.Body, &body)
	if err != nil {
		log.Printf("не смогли распарсить тело записи: %s", err)
	}

	return data, body, err
}

// Update обработка событий пользователя.
func (m *LoginPasswordStageType) Update(a *agent, msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(m.inputs))

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return tea.Quit

		//Назад
		case "ctrl+b":
			a.currenStage = "list"
			a.recordID = ""
			a.Stages[a.currenStage].Prepare(a)
			return nil

		//Сохранить
		case "ctrl+s":
			m.save(a)
			return nil

		case "tab", "shift+tab", "enter", "up", "down":
			s := msg.String()

			if s == "up" || s == "shift+tab" {
				m.focusIndex--
			} else {
				m.focusIndex++
			}

			if m.focusIndex >= len(m.inputs) {
				m.focusIndex = 0
			} else if m.focusIndex < 0 {
				m.focusIndex = len(m.inputs) - 1
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

// save сохранение формы на сервере.
func (m *LoginPasswordStageType) save(a *agent) {
	dataBody := models.LoginBody{
		Login:    m.inputs[1].Value(),
		Password: m.inputs[2].Value(),
	}

	body, _ := json.Marshal(dataBody)
	body, _ = mycrypto.Encrypt(body, m.userID)

	data := models.Data{
		ID:          m.recordID,
		UserID:      m.userID,
		Name:        []byte(m.inputs[0].Value()),
		Type:        "login_password",
		Deleted:     false,
		Description: []byte(m.inputs[3].Value()),
		Date:        time.Now(),
		Body:        body,
	}

	bytes, _ := json.Marshal(data)

	ctx := context.TODO()
	headers := map[string]string{
		"Authorization": m.token,
	}

	url := fmt.Sprintf("http://%s/api/data/update", config.ConfigClient.AddrServer)
	response := (*m.client).Update(ctx, url, headers, bytes)

	if response.Code != 200 {
		a.currenStage = "info"
		a.Stages[a.currenStage] = InitInfoModel(string(response.Response), "login_password", "Назад")
		return
	}

	a.currenStage = "list"
	a.recordID = ""
	a.Stages[a.currenStage].Prepare(a)
}

// updateInputs обработка изменений полей формы.
func (m *LoginPasswordStageType) updateInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(m.inputs))

	for i := range m.inputs {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	}

	return tea.Batch(cmds...)
}

// View отображение формы в терминале.
func (m *LoginPasswordStageType) View() string {
	var b strings.Builder

	for i := range m.inputs {
		b.WriteString(m.inputs[i].View())
		if i < len(m.inputs)-1 {
			b.WriteRune('\n')
		}
	}

	b.WriteRune('\n')
	b.WriteString("[ Ctrl+s ] - Сохранить\n")
	b.WriteString("[ Ctrl+b ] - Назад\n")

	return b.String()
}
