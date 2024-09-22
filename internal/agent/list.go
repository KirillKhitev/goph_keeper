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

// Prepare подготавливает список записей пользователя.
func (m *ListStageType) Prepare(a *agent) {
	m.userID = a.userID
	m.client = a.client
	m.List = *a.list

	m.List.Title = "Мои записи"

	ctx := context.TODO()
	url := fmt.Sprintf("http://%s/api/data/list", config.ConfigClient.AddrServer)
	response := (*m.client).List(ctx, url, map[string]string{
		"Authorization": a.token,
	})

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
func (m *ListStageType) Update(a *agent, msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return tea.Quit

		case "enter":
			selectedItem := m.List.SelectedItem().(listItem)
			a.currenStage = selectedItem.type_record
			a.recordID = selectedItem.id
			a.Stages[a.currenStage].Prepare(a)
			return nil

		case "ctrl+n":
			a.currenStage = "operation_list"
			return nil
		}

	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.List.SetSize(msg.Width-h, msg.Height-v)
	}

	var cmd tea.Cmd
	m.List, cmd = m.List.Update(msg)

	return cmd
}

// View отображает элементы списка записей пользователя.
func (m *ListStageType) View() string {
	s := m.List.View()
	s += "\n\n[ Ctrl+n ] - Создать новую запись\n"

	return s
}
