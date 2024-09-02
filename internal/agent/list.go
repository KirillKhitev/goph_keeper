package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/KirillKhitev/goph_keeper/internal/client"
	"github.com/KirillKhitev/goph_keeper/internal/config"
	"github.com/KirillKhitev/goph_keeper/internal/models"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"log"
)

// Стиль документа.
var docStyle = lipgloss.NewStyle().Padding(2, 2)

// openForm команда для открытия формы пользовательской записи.
type openForm struct {
	id          string
	type_record string
}

// listItem элемент списка записей.
type listItem struct {
	id, title, desc, type_record string
}

// Title возвращает заголовок элемента списка.
func (i listItem) Title() string { return i.title }

// Description возвращает описание элемента списка.
func (i listItem) Description() string { return i.desc }

// FilterValue метод для фильтрации элементов списка.
func (i listItem) FilterValue() string { return i.title }

// ListStageType модель списка записей пользователя.
type ListStageType struct {
	userID string
	List   list.Model
	client *client.Client
}

// Init - заглушка для интерфейса.
func (m *ListStageType) Init() tea.Cmd {
	return nil
}

// Prepare подготавливает список записей пользователя.
func (m *ListStageType) Prepare(a *agent) {
	m.userID = a.userID
	m.client = a.client
	m.List = *a.list

	m.List.Title = "Мои записи"

	ctx := context.TODO()
	url := fmt.Sprintf("http://%s/api/data/list", config.ConfigClient.AddrServer)

	response := (*m.client).Send(ctx, url, map[string]string{
		"Authorization": a.token,
	}, []byte{}, "POST")

	var result []models.Data

	if err := json.Unmarshal(response.Response, &result); err != nil {
		log.Println("Error unmarshalling response: ", err)
		return
	}

	var items []list.Item

	for _, row := range result {
		items = append(items, listItem{
			id:          row.ID,
			title:       string(row.Name),
			type_record: row.Type,
			desc:        string(row.Description),
		})
	}

	m.List.SetItems(items)
}

// Update обрабатывает события пользователя.
func (m *ListStageType) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit

		case "enter":
			selectedItem := m.List.SelectedItem().(listItem)
			return m, func() tea.Msg {
				return openForm{
					id:          selectedItem.id,
					type_record: selectedItem.type_record,
				}
			}

		case "ctrl+n":
			return m, func() tea.Msg {
				return openStage("operation_list")
			}
		}

	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.List.SetSize(msg.Width-h, msg.Height-v)
	}

	var cmd tea.Cmd
	m.List, cmd = m.List.Update(msg)

	return m, cmd
}

// View отображает элементы списка записей пользователя.
func (m *ListStageType) View() string {
	s := m.List.View()
	s += "\n\n[ Ctrl+n ] - Создать новую запись\n"

	return s
}
