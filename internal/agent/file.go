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

// clearErrorAfter команда очистки ошибки.
func clearErrorAfter(t time.Duration) tea.Cmd {
	return tea.Tick(t, func(_ time.Time) tea.Msg {
		return clearErrorMsg{}
	})
}

// Init инициализирует модель.
func (m *FileStageType) Init() tea.Cmd {
	return nil
}

// Update обрабатывает события пользователя.
func (m *FileStageType) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.recordID != "" {
		return m, func() tea.Msg {
			return infoMsg{
				message:    "Файл был скачен в папку files",
				back:       "list",
				backButton: "На список",
			}
		}
	}

	var cmd tea.Cmd

	fp, cmd := m.filepicker.Update(msg)

	m.filepicker = fp

	didSelect, path := m.filepicker.DidSelectFile(msg)

	if didSelect {
		m.selectedFile = path
	}

	if didSelect, path := m.filepicker.DidSelectDisabledFile(msg); didSelect {
		m.err = errors.New(path + " не валидный.")
		m.selectedFile = ""
		return m, tea.Batch(cmd, clearErrorAfter(2*time.Second))
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.quitting = true
			return m, tea.Quit
		case "enter":
			return m.save()
		}
	case clearErrorMsg:
		m.err = nil
	}

	return m, cmd
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
func (m *FileStageType) save() (tea.Model, tea.Cmd) {
	body, err := os.ReadFile(m.selectedFile)

	if err != nil {
		return m, func() tea.Msg {
			log.Println("save ", m.selectedFile, err)
			return infoMsg{
				message:    fmt.Sprintf("Ошибка при чтении файла"),
				back:       "file",
				backButton: "Назад",
			}
		}
	}

	body, _ = mycrypto.Encrypt(body, m.userID)

	f, err := os.Open(m.selectedFile)
	if err != nil {
		return m, func() tea.Msg {
			log.Println("save ", m.selectedFile, err)
			return infoMsg{
				message:    fmt.Sprintf("Ошибка при получении статистики файла"),
				back:       "file",
				backButton: "Назад",
			}
		}
	}

	defer f.Close()

	file, _ := f.Stat()

	if file.Size() > 10737418240 {
		return m, func() tea.Msg {
			log.Println("save ", m.selectedFile, err)
			return infoMsg{
				message:    fmt.Sprintf("Файл не может быть больше 10Гб"),
				back:       "file",
				backButton: "Назад",
			}
		}
	}

	data := models.Data{
		ID:          m.recordID,
		UserID:      m.userID,
		Name:        []byte(file.Name()),
		Type:        "file",
		Deleted:     false,
		Description: []byte("Файл"),
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
		return m, func() tea.Msg {
			return infoMsg{
				message:    string(response.Response),
				back:       "file",
				backButton: "Назад",
			}
		}
	}

	return m, func() tea.Msg {
		return openList{}
	}
}
