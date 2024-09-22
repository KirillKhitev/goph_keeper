package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/KirillKhitev/goph_keeper/internal/config"
	"github.com/KirillKhitev/goph_keeper/internal/models"
	"github.com/KirillKhitev/goph_keeper/internal/mycrypto"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// CreditCardStageType модель формы кредитной карты.
type CreditCardStageType struct {
	LoginPasswordStageType
}

// Индексы полей формы.
const (
	ccn = iota + 1
	exp
	cvv
)

// Prepare подготавливает модель.
func (m *CreditCardStageType) Prepare(a *agent) {
	m.inputs = make([]textinput.Model, 5)
	m.back = "operation_list"
	m.client = a.client
	m.userID = a.userID
	m.token = a.token
	m.recordID = a.recordID

	var body models.CreditCardBody
	data := models.Data{
		ID:     m.recordID,
		UserID: m.userID,
		Type:   "credit_card",
	}

	data, body, _ = getRecordDataFromServer[models.CreditCardBody](m, data, body)

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
		case ccn:
			t.Placeholder = "Номер карты"
			t.SetValue(body.Ccn)
			t.Validate = ccnValidator
		case exp:
			t.Placeholder = "Срок действия"
			t.SetValue(body.Exp)
			t.Validate = expValidator
		case cvv:
			t.Placeholder = "cvv"
			t.SetValue(body.CVV)
			t.Validate = cvvValidator
		case 4:
			t.Placeholder = "Описание"
			t.SetValue(string(data.Description))
		}

		m.inputs[i] = t
	}

}

// Update обрабатывает события пользователя.
func (m *CreditCardStageType) Update(a *agent, msg tea.Msg) tea.Cmd {
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
			log.Println("Нажали Сохранить")
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

// save отправляет данные формы на сервер.
func (m *CreditCardStageType) save(a *agent) {
	dataBody := models.CreditCardBody{
		Ccn: m.inputs[ccn].Value(),
		Exp: m.inputs[exp].Value(),
		CVV: m.inputs[cvv].Value(),
	}

	body, _ := json.Marshal(dataBody)
	body, _ = mycrypto.Encrypt(body, m.userID)

	data := models.Data{
		ID:          m.recordID,
		UserID:      m.userID,
		Name:        []byte(m.inputs[0].Value()),
		Type:        "credit_card",
		Deleted:     false,
		Description: []byte(m.inputs[4].Value()),
		Date:        time.Now(),
		Body:        body,
	}

	bytes, _ := json.Marshal(data)

	ctx := context.Background()
	headers := map[string]string{
		"Authorization": m.token,
	}

	url := fmt.Sprintf("http://%s/api/data/update", config.ConfigClient.AddrServer)
	response := (*m.client).Update(ctx, url, headers, bytes)

	if response.Code != 200 {
		a.currenStage = "info"
		a.Stages[a.currenStage] = InitInfoModel(string(response.Response), "credit_card", "Назад")
		return
	}

	a.currenStage = "list"
	a.recordID = ""
	a.Stages[a.currenStage].Prepare(a)
}

// ccnValidator валидирует номер карты.
func ccnValidator(s string) error {
	if len(s) > 16+3 {
		return fmt.Errorf("Номер карты слишком длинный")
	}

	if len(s) == 0 || len(s)%5 != 0 && (s[len(s)-1] < '0' || s[len(s)-1] > '9') {
		return fmt.Errorf("Неверный номер карты")
	}

	if len(s)%5 == 0 && s[len(s)-1] != ' ' {
		return fmt.Errorf("Номер карты должен быть разделен пробелами")
	}

	c := strings.ReplaceAll(s, " ", "")
	_, err := strconv.ParseInt(c, 10, 64)

	return err
}

// expValidator валидирует срок действия карты.
func expValidator(s string) error {
	e := strings.ReplaceAll(s, "/", "")
	_, err := strconv.ParseInt(e, 10, 64)

	if err != nil {
		return fmt.Errorf("EXP is invalid")
	}

	if len(s) >= 3 && (strings.Index(s, "/") != 2 || strings.LastIndex(s, "/") != 2) {
		return fmt.Errorf("EXP is invalid")
	}

	return nil
}

// cvvValidator валидирует ccv-код.
func cvvValidator(s string) error {
	_, err := strconv.ParseInt(s, 10, 64)
	return err
}
