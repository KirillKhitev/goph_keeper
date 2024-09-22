package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/KirillKhitev/goph_keeper/internal/config"
	"github.com/KirillKhitev/goph_keeper/internal/models"
	"github.com/KirillKhitev/goph_keeper/internal/mycrypto"
	"github.com/charmbracelet/bubbles/filepicker"
	tea "github.com/charmbracelet/bubbletea"
	"log"
	"os"
	"slices"
	"strings"
	"time"
)

// FileStageType модель формы файла пользователя.
type FileStageType struct {
	LoginPasswordStageType
	filepicker   filepicker.Model
	selectedFile string
	quitting     bool
	err          error
}

// Сообщение очистки ошибки.
type clearErrorMsg struct{}

const ChunkSize = 50000000

// clearErrorAfter команда очистки ошибки.
func clearErrorAfter(t time.Duration) tea.Cmd {
	return tea.Tick(t, func(_ time.Time) tea.Msg {
		return clearErrorMsg{}
	})
}

// Update обрабатывает события пользователя.
func (m *FileStageType) Update(a *agent, msg tea.Msg) tea.Cmd {
	if m.recordID != "" {
		a.currenStage = "info"
		a.Stages[a.currenStage] = InitInfoModel("Файл был скачен в папку files", "list", "На список")
		return nil
	}

	var cmd tea.Cmd

	fp, cmd := m.filepicker.Update(msg)

	m.filepicker = fp

	didSelect, path := m.filepicker.DidSelectFile(msg)

	if didSelect {
		m.selectedFile = path
	}

	if didSelect, path := m.filepicker.DidSelectDisabledFile(msg); didSelect {
		m.err = errors.New(path + " не валидный.\r\n")
		m.selectedFile = ""
		return tea.Batch(cmd, clearErrorAfter(2*time.Second))
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.quitting = true
			return tea.Quit
		case "enter":
			m.save(a)
			return nil
		}
	case clearErrorMsg:
		m.err = nil
	}

	return cmd
}

// View отображает форму в терминале.
func (m *FileStageType) View() string {
	if m.quitting {
		return ""
	}

	var s strings.Builder

	s.WriteString("\n  ")

	if m.err != nil {
		s.WriteString(m.filepicker.Styles.DisabledFile.Render(m.err.Error()))
	} else if m.selectedFile == "" {
		s.WriteString("Выберите файл: " + strings.Join(m.filepicker.AllowedTypes[:], ", "))
		s.WriteString("\n\n" + m.filepicker.View() + "\n")
	} else {
		s.WriteString("Выбранный файл: " + m.filepicker.Styles.Selected.Render(m.selectedFile))
		s.WriteString("\n\n" + m.filepicker.View() + "\n")
	}

	return s.String()
}

// Prepare подготавливает модель для отображения.
func (m *FileStageType) Prepare(a *agent) {
	m.back = "operation_list"
	m.client = a.client
	m.userID = a.userID
	m.token = a.token
	m.recordID = a.recordID

	m.filepicker = *a.fp

	var body []byte
	data := models.Data{
		ID:     m.recordID,
		UserID: m.userID,
		Type:   "file",
	}

	data, _, _ = getRecordDataFromServer[[]byte](m, data, body)
	path := "files" + string(os.PathSeparator) + string(data.Name)

	f, err := os.Create(path)
	defer func() {
		f.Close()
	}()

	if err != nil {
		log.Printf("ошибка при создании файла[%s]: %s", data.Name, err)
		return
	}

	f.Write(data.Body)
}

// save отправляет данные формы на сервер.
func (m *FileStageType) save(a *agent) {
	body, err := os.ReadFile(m.selectedFile)

	if err != nil {
		log.Println("save ", m.selectedFile, err)
		a.currenStage = "info"
		a.Stages[a.currenStage] = InitInfoModel("Ошибка при чтении файла", "file", "Назад")
		return
	}

	body, _ = mycrypto.Encrypt(body, m.userID)

	f, _ := os.Open(m.selectedFile)

	defer f.Close()

	file, _ := f.Stat()

	url := fmt.Sprintf("http://%s/api/data/update", config.ConfigClient.AddrServer)
	headers := map[string]string{
		"Authorization": m.token,
	}

	var c int
	for bp := range slices.Chunk(body, ChunkSize) {
		data := models.Data{
			ID:          m.recordID,
			UserID:      m.userID,
			Name:        []byte(file.Name()),
			Type:        "file",
			Deleted:     false,
			Description: []byte("Файл"),
			Date:        time.Now(),
			Body:        bp,
			Part:        c,
		}

		bytes, _ := json.Marshal(data)

		ctx := context.TODO()

		response := (*m.client).Update(ctx, url, headers, bytes)

		if response.Code != 200 {
			a.currenStage = "info"
			a.Stages[a.currenStage] = InitInfoModel(string(response.Response), "file", "Назад")
			return
		}

		m.recordID = string(response.Response)

		c++
	}

	a.currenStage = "list"
	a.recordID = ""
	a.Stages[a.currenStage].Prepare(a)
}
