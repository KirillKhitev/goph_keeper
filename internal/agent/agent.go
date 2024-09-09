// Пакет моделей экранов агента.
package agent

import (
	"github.com/KirillKhitev/goph_keeper/internal/client"
	"github.com/charmbracelet/bubbles/filepicker"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"log"
	"os"
	"os/signal"
	"syscall"
)

// Базовый интерфейс модели экрана агента.
type stage interface {
	Init() tea.Cmd
	Update(msg tea.Msg) (tea.Model, tea.Cmd)
	View() string
	Prepare(a *agent)
}

// Модель экрана агента.
type stageAgent struct {
	stage
	choices    []string
	selected   int
	cursor     int
	back       string
	backButton string
	userID     string
	token      string
	recordID   string
}

// Главная модель агента.
type agent struct {
	client      *client.Client
	userID      string
	token       string
	currenStage string
	Stages      map[string]stage
	list        *list.Model
	fp          *filepicker.Model
	recordID    string
}

// Стили отображения текста
var (
	focusedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	blurredStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	cursorStyle  = focusedStyle
	noStyle      = lipgloss.NewStyle()
)

// Команда вывода ошибок
type errMsg struct {
	error
	back string
}

// Команда вывода списка записей пользователя, с запросом обновления с сервера.
type openList struct{}

// Вывод ошибки в текстовом виде.
func (e errMsg) Error() string { return e.error.Error() }

// Init инициализирует модель агента.
func (a agent) Init() tea.Cmd {
	cmds := make([]tea.Cmd, 2)
	cmds = append(cmds, a.fp.Init(), a.Stages[a.currenStage].Init())

	return tea.Batch(cmds...)
}

var OpenListMsg = func() tea.Msg {
	return openList{}
}

// Prepare подготавливает модель.
func (a agent) Prepare(agent *agent) {

}

// Update отвечает за выполение команд на дейстсвия пользователя.
// Обрабатывает стартовые события и осуществляет роутинг между моделями агента.
func (a agent) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		a.list.SetSize(msg.Width-h, msg.Height-v)
	}

	cmds := make([]tea.Cmd, 0)

	// Инициализируем считываение списка файлов в домашней директории
	if a.currenStage == "start" {
		var cmdFile tea.Cmd
		*a.fp, cmdFile = a.fp.Update(msg)
		cmds = append(cmds, cmdFile)
	}

	_, cmd := a.Stages[a.currenStage].Update(msg)

	if cmd != nil {
		switch command := cmd().(type) {
		case openStage:
			a.currenStage = string(command)
		case errMsg:
			a.currenStage = "error"
			a.Stages[a.currenStage] = InitErrorModel(command.error, command.back)
		case infoMsg:
			a.currenStage = "info"
			a.Stages[a.currenStage] = InitInfoModel(command.message, command.back, command.backButton)
		case openList:
			a.currenStage = "list"
			a.recordID = ""
			a.Stages[a.currenStage].Prepare(&a)
		case authSuccessMsg:
			a.currenStage = "list"
			a.userID = command.userID
			a.token = command.token

			(*a.client).SetUserID(command.userID)

			a.Stages[a.currenStage].Prepare(&a)
		case openForm:
			a.currenStage = command.type_record
			a.recordID = command.id
			a.Stages[a.currenStage].Prepare(&a)
		}
	}

	cmds = append(cmds, cmd)

	return a, tea.Batch(cmds...)
}

// View отображает экран текущей модели агента.
func (a agent) View() string {
	var s string

	s = a.Stages[a.currenStage].View()
	s += "[ Ctrl+c ] - Выход из программы\n"
	return s
}

// NewAgent конструктор главной структуры приложения Агента.
func NewAgent() (*agent, error) {
	client, err := newClient()
	if err != nil {
		return nil, err
	}

	var items []list.Item
	listModel := list.New(items, list.NewDefaultDelegate(), 0, 0)

	fp := filepicker.New()
	fp.AllowedTypes = []string{".txt", ".png", ".jpg"}
	fp.CurrentDirectory, err = os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	stages := map[string]stage{}
	stages["start"] = &StartStageType{}
	stages["registration"] = &RegisterStageType{}
	stages["login"] = &LoginStageType{}
	stages["error"] = &ErrorStageType{}
	stages["info"] = &InfoStageType{}
	stages["list"] = &ListStageType{}
	stages["operation_list"] = &OperationListStageType{}
	stages["login_password"] = &LoginPasswordStageType{}
	stages["credit_card"] = &CreditCardStageType{}
	stages["text"] = &TextStageType{}
	stages["file"] = &FileStageType{}

	a := &agent{
		client:      &client,
		Stages:      stages,
		currenStage: "start",
		list:        &listModel,
		fp:          &fp,
	}

	a.Stages["registration"].Prepare(a)
	a.Stages["login"].Prepare(a)
	a.Stages["operation_list"].Prepare(a)
	a.Stages[a.currenStage].Prepare(a)

	return a, nil
}

// Конструктор Http-клиента.
func newClient() (client.Client, error) {
	return client.NewRestyClient()
}

// CatchTerminateSignal ловит сигналы ОС для корректной остановки агента.
func (a *agent) CatchTerminateSignal() error {
	terminateSignals := make(chan os.Signal, 1)

	signal.Notify(terminateSignals, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	<-terminateSignals

	if err := a.Close(); err != nil {
		return err
	}

	return nil
}

// Close отвечает за корректную остановку агента.
func (a *agent) Close() error {
	(*a.client).Close()

	log.Println("Successful stop agent")

	return nil
}
