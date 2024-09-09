package agent

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/KirillKhitev/goph_keeper/internal/api"
	"github.com/KirillKhitev/goph_keeper/internal/auth"
	"github.com/KirillKhitev/goph_keeper/internal/client"
	"github.com/KirillKhitev/goph_keeper/internal/config"
	"github.com/KirillKhitev/goph_keeper/internal/models"
	"github.com/KirillKhitev/goph_keeper/internal/mycrypto"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"net/http"
	"os"
	"reflect"
	"testing"
	"time"
)

type ClientForTests struct {
	userID string
}

func NewClient() client.Client {
	res := NewClientForTests()

	return res
}

func NewClientForTests() *ClientForTests {
	res := &ClientForTests{}

	return res
}

func (c *ClientForTests) SetUserID(userID string) {
	c.userID = userID
}

func (c *ClientForTests) Get(ctx context.Context, headers map[string]string, data []byte) client.APIServiceResult {
	var dataRequest models.Data

	_ = json.Unmarshal(data, &dataRequest)

	if dataRequest.ID != "exist_id" {
		return client.APIServiceResult{
			Code: http.StatusNotFound,
		}
	}

	existdata := models.Data{
		ID:          "exist_id",
		Name:        []byte("Hello World"),
		Type:        dataRequest.Type,
		Description: []byte("Description"),
		Body:        dataRequest.Body,
	}

	if existdata.Type == "file" {
		existdata.ID = ""
		existdata.Name = []byte("test_file.txt")
	} else {
		existdata.Body, _ = json.Marshal(models.LoginBody{
			Login:    "Login_test",
			Password: "Login_pass",
		})
	}

	b, _ := mycrypto.Encrypt(existdata.Body, dataRequest.UserID)
	existdata.Body = b

	bodyR, _ := json.Marshal(existdata)

	return client.APIServiceResult{
		Code:     http.StatusOK,
		Response: bodyR,
	}
}

func (c *ClientForTests) List(ctx context.Context, headers map[string]string) client.APIServiceResult {
	result := []models.Data{
		{
			ID:          "111",
			UserID:      "exist_user",
			Name:        []byte("Первая"),
			Type:        "login_password",
			Description: []byte("Описание"),
		},
		{
			ID:          "222",
			UserID:      "exist_user",
			Name:        []byte("Вторая"),
			Type:        "credit_card",
			Description: []byte("Описание"),
		},
	}

	body, _ := json.Marshal(result)

	return client.APIServiceResult{
		Code:     http.StatusOK,
		Response: body,
	}
}

func (c *ClientForTests) Update(ctx context.Context, headers map[string]string, data []byte) client.APIServiceResult {
	var d models.Data

	_ = json.Unmarshal(data, &d)

	if d.ID == "error" {
		return client.APIServiceResult{
			Code: http.StatusInternalServerError,
		}
	}

	return client.APIServiceResult{
		Code: http.StatusOK,
	}
}

func (c *ClientForTests) Login(ctx context.Context, data []byte) client.APIServiceResult {
	var r auth.AuthorizingData

	_ = json.Unmarshal(data, &r)

	if r.UserName == "error" {
		return client.APIServiceResult{
			Code: http.StatusInternalServerError,
		}
	}

	b := api.UserAuthBody{
		ID: "test_user",
	}

	body, _ := json.Marshal(b)

	return client.APIServiceResult{
		Code:     http.StatusOK,
		Response: body,
	}
}

func (c *ClientForTests) Register(ctx context.Context, data []byte) client.APIServiceResult {
	var r auth.AuthorizingData

	_ = json.Unmarshal(data, &r)

	if r.UserName == "error" {
		return client.APIServiceResult{
			Code: http.StatusInternalServerError,
		}
	}

	b := api.UserAuthBody{
		ID: "test_user",
	}

	body, _ := json.Marshal(b)

	return client.APIServiceResult{
		Code:     http.StatusOK,
		Response: body,
	}
}

func (c *ClientForTests) Close() error {
	return nil
}

func createTestFile() *os.File {
	os.Mkdir("users", 777)
	file, _ := os.Create("users/test_user")

	key := []uint8{168, 49, 93, 155, 71, 68, 144, 227, 208, 250, 186, 149, 211, 116, 214, 100}

	body := base64.StdEncoding.EncodeToString(key)

	file.Write([]byte(body))

	return file
}

func TestCreditCardStageType_Prepare(t *testing.T) {
	f := createTestFile()

	defer func() {
		f.Close()
		os.RemoveAll("users")
	}()

	app, err := NewAgent()
	clientT := NewClient()
	app.client = &clientT

	if err != nil {
		t.Fatal(err)
	}

	type fields struct {
		LoginPasswordStageType LoginPasswordStageType
	}
	type args struct {
		a    *agent
		data models.Data
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "positive test #1",
			fields: fields{
				LoginPasswordStageType: LoginPasswordStageType{},
			},
			args: args{
				a: app,
				data: models.Data{
					ID:     "exist_id",
					UserID: "test_user",
				},
			},
		},
		{
			name: "positive test #2",
			fields: fields{
				LoginPasswordStageType: LoginPasswordStageType{},
			},
			args: args{
				a: app,
				data: models.Data{
					ID:     "wrong",
					UserID: "test_user",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.args.a.recordID = tt.args.data.ID
			tt.args.a.userID = tt.args.data.UserID

			m := &CreditCardStageType{
				LoginPasswordStageType: tt.fields.LoginPasswordStageType,
			}

			m.Prepare(tt.args.a)

			if len(m.inputs) != 5 {
				t.Errorf("Prepare wrong!")
			}
		})
	}
}

func TestCreditCardStageType_Update(t *testing.T) {
	f := createTestFile()

	defer func() {
		f.Close()
		os.RemoveAll("users")
	}()

	clientTests := NewClient()
	app, err := NewAgent()
	app.client = &clientTests

	if err != nil {
		t.Fatal(err)
	}

	type fields struct {
		LoginPasswordStageType LoginPasswordStageType
	}
	type args struct {
		msgs        []tea.Msg
		needPrepare bool
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   tea.Cmd
	}{
		{
			name: "positive test #1",
			fields: fields{
				LoginPasswordStageType: LoginPasswordStageType{},
			},
			args: args{
				msgs: []tea.Msg{
					tea.KeyMsg{Type: tea.KeyCtrlC},
					tea.KeyMsg{Type: tea.KeyEsc},
				},
			},
			want: tea.Quit,
		},
		{
			name: "positive test #2",
			fields: fields{
				LoginPasswordStageType: LoginPasswordStageType{},
			},
			args: args{
				msgs: []tea.Msg{
					tea.KeyMsg{Type: tea.KeyTab},
					tea.KeyMsg{Type: tea.KeyShiftTab},
					tea.KeyMsg{Type: tea.KeyEnter},
					tea.KeyMsg{Type: tea.KeyUp},
					tea.KeyMsg{Type: tea.KeyDown},
				},
			},
			want: nil,
		},
		{
			name: "positive test #3",
			fields: fields{
				LoginPasswordStageType: LoginPasswordStageType{},
			},
			args: args{
				msgs: []tea.Msg{
					tea.KeyMsg{Type: tea.KeyCtrlB},
				},
			},
			want: OpenListMsg,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &CreditCardStageType{
				LoginPasswordStageType: tt.fields.LoginPasswordStageType,
			}

			if tt.args.needPrepare {
				m.Prepare(app)
			}

			for _, msg := range tt.args.msgs {
				_, got := m.Update(msg.(tea.KeyMsg))

				if fmt.Sprintf("%v", got) != fmt.Sprintf("%v", tt.want) {
					t.Errorf("Update(%s) got = %v, want %v", msg, got, tt.want)
				}
			}
		})
	}

	tests2 := []struct {
		name   string
		fields fields
		args   args
		want   tea.Cmd
	}{
		{
			name: "positive test #4",
			fields: fields{
				LoginPasswordStageType: LoginPasswordStageType{},
			},
			args: args{
				msgs: []tea.Msg{
					tea.KeyMsg{Type: tea.KeyCtrlS},
				},
			},
			want: OpenListMsg,
		},
	}

	config.ConfigClient.AddrServer = "localhost:8080"

	for _, tt := range tests2 {
		t.Run(tt.name, func(t *testing.T) {
			m := &CreditCardStageType{
				LoginPasswordStageType: tt.fields.LoginPasswordStageType,
			}

			m.inputs = make([]textinput.Model, 5)
			m.back = "operation_list"
			m.userID = "test_user"
			m.recordID = "12312321"
			m.client = app.client

			var tm textinput.Model

			for i := range m.inputs {
				tm = textinput.New()
				switch i {
				case 0:
					tm.SetValue("Тестовая карта")
				case ccn:
					tm.SetValue("111")
					tm.Validate = ccnValidator
				case exp:
					tm.SetValue("06/30")
					tm.Validate = expValidator
				case cvv:
					tm.SetValue("305")
					tm.Validate = cvvValidator
				case 4:
					tm.SetValue("Описание карты")
				}

				m.inputs[i] = tm
			}

			for _, msg := range tt.args.msgs {
				m.Update(msg.(tea.KeyMsg))
			}
		})
	}
}

func TestCreditCardStageType_save(t *testing.T) {
	f := createTestFile()

	defer func() {
		f.Close()
		os.RemoveAll("users")
	}()

	clientTests := NewClient()
	app, _ := NewAgent()
	app.client = &clientTests

	type fields struct {
		LoginPasswordStageType LoginPasswordStageType
	}

	type args struct {
		recordID string
	}

	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name: "positive test #1",
			fields: fields{
				LoginPasswordStageType: LoginPasswordStageType{},
			},
			args: args{
				recordID: "success",
			},
			want: "agent.openList",
		},
		{
			name: "negative test #2",
			fields: fields{
				LoginPasswordStageType: LoginPasswordStageType{},
			},
			args: args{
				recordID: "error",
			},
			want: "agent.infoMsg",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &CreditCardStageType{
				LoginPasswordStageType: tt.fields.LoginPasswordStageType,
			}

			m.Prepare(app)
			m.recordID = tt.args.recordID

			_, got := m.save()
			if fmt.Sprintf("%T", got()) != fmt.Sprintf("%s", tt.want) {
				t.Errorf("save() got = %T, want %s", got(), tt.want)
			}
		})
	}
}

func TestErrorStageType_Init(t *testing.T) {
	type fields struct {
		focusIndex int
		back       string
		error      error
	}
	tests := []struct {
		name   string
		fields fields
		want   tea.Cmd
	}{
		{
			name: "posititve test #1",
			fields: fields{
				focusIndex: 0,
				back:       "Назад",
				error:      nil,
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &ErrorStageType{
				focusIndex: tt.fields.focusIndex,
				back:       tt.fields.back,
				error:      tt.fields.error,
			}
			if got := s.Init(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Init() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestErrorStageType_Prepare(t *testing.T) {
	app, _ := NewAgent()

	tests := []struct {
		name string
		want interface{}
	}{
		{
			name: "positive test #1",
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &ErrorStageType{}
			s.Prepare(app)
		})
	}
}

func TestErrorStageType_Update(t *testing.T) {
	type fields struct {
		focusIndex int
		back       string
		error      error
	}
	type args struct {
		msgs []tea.Msg
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name: "positive test #1",
			args: args{
				msgs: []tea.Msg{
					tea.KeyMsg{Type: tea.KeyCtrlC},
					tea.KeyMsg{Type: tea.KeyEsc},
				},
			},
			want: "tea.QuitMsg",
		},
		{
			name: "positive test #2",
			args: args{
				msgs: []tea.Msg{
					tea.KeyMsg{Type: tea.KeyEnter},
				},
			},
			want: "agent.openStage",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &ErrorStageType{}

			for _, msg := range tt.args.msgs {
				_, got := m.Update(msg.(tea.KeyMsg))

				if fmt.Sprintf("%T", got()) != fmt.Sprintf("%s", tt.want) {
					t.Errorf("Update(%s) got = %T, want %s", msg, got(), tt.want)
				}
			}
		})
	}
}

func TestErrorStageType_View(t *testing.T) {
	type fields struct {
		focusIndex int
		back       string
		error      error
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "positive test #1",
			fields: fields{
				back:  "Назад",
				error: errors.New("ошибка"),
			},
			want: "ошибка\n[ Назад ]\n\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &ErrorStageType{
				focusIndex: tt.fields.focusIndex,
				back:       tt.fields.back,
				error:      tt.fields.error,
			}
			if got := m.View(); got != tt.want {
				t.Errorf("View() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFileStageType_Init(t *testing.T) {
	type fields struct {
		LoginPasswordStageType LoginPasswordStageType
	}
	tests := []struct {
		name   string
		fields fields
		want   tea.Cmd
	}{
		{
			name: "positive test #1",
			fields: fields{
				LoginPasswordStageType: LoginPasswordStageType{},
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &FileStageType{
				LoginPasswordStageType: tt.fields.LoginPasswordStageType,
			}
			if got := m.Init(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Init() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFileStageType_Prepare(t *testing.T) {
	clientT := NewClient()
	app, _ := NewAgent()
	app.client = &clientT

	type fields struct {
		LoginPasswordStageType LoginPasswordStageType
	}
	type args struct {
		a *agent
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "positive test #1",
			fields: fields{
				LoginPasswordStageType: LoginPasswordStageType{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &FileStageType{
				LoginPasswordStageType: tt.fields.LoginPasswordStageType,
			}
			m.Prepare(app)
		})
	}
}

func TestFileStageType_Update(t *testing.T) {
	clientT := NewClient()
	app, _ := NewAgent()
	app.client = &clientT

	type fields struct {
		LoginPasswordStageType LoginPasswordStageType
	}
	type args struct {
		recordID string
		name     string
		msgs     []tea.Msg
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name: "posisitive test #1",
			fields: fields{
				LoginPasswordStageType: LoginPasswordStageType{},
			},
			args: args{
				recordID: "exist_id",
				msgs: []tea.Msg{
					tea.KeyMsg{Type: tea.KeyCtrlC},
				},
			},
			want: "agent.infoMsg",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Mkdir("files", 777)

			defer func() {
				os.RemoveAll("files")
			}()

			m := &FileStageType{
				LoginPasswordStageType: tt.fields.LoginPasswordStageType,
			}

			app.recordID = tt.args.recordID

			m.Prepare(app)

			for _, msg := range tt.args.msgs {
				_, got := m.Update(msg.(tea.KeyMsg))

				if fmt.Sprintf("%T", got()) != fmt.Sprintf("%s", tt.want) {
					t.Errorf("Update(%s) got = %T, want %s", msg, got(), tt.want)
				}
			}
		})
	}

	tests2 := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name: "posisitive test #2",
			fields: fields{
				LoginPasswordStageType: LoginPasswordStageType{},
			},
			args: args{
				recordID: "exist_id",
				msgs: []tea.Msg{
					tea.KeyMsg{Type: tea.KeyCtrlC},
					tea.KeyMsg{Type: tea.KeyEsc},
				},
			},
			want: "tea.QuitMsg",
		},
		{
			name: "posisitive test #3",
			fields: fields{
				LoginPasswordStageType: LoginPasswordStageType{},
			},
			args: args{
				recordID: "exist_id",
				msgs: []tea.Msg{
					tea.KeyMsg{Type: tea.KeyEnter},
				},
			},
			want: "agent.openList",
		},
	}
	for _, tt := range tests2 {
		t.Run(tt.name, func(t *testing.T) {
			os.Mkdir("files", 777)

			defer func() {
				os.RemoveAll("files")
			}()

			m := &FileStageType{
				LoginPasswordStageType: tt.fields.LoginPasswordStageType,
			}

			app.recordID = tt.args.recordID

			m.Prepare(app)
			m.recordID = ""
			m.selectedFile = "files\\test_file.txt"

			for _, msg := range tt.args.msgs {
				_, got := m.Update(msg.(tea.KeyMsg))

				if fmt.Sprintf("%T", got()) != fmt.Sprintf("%s", tt.want) {
					t.Errorf("Update(%s) got = %T, want %s", msg, got(), tt.want)
				}
			}
		})
	}

}

func TestFileStageType_View(t *testing.T) {
	type fields struct {
		LoginPasswordStageType LoginPasswordStageType
		selectedFile           string
		quitting               bool
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "positive test #1",
			fields: fields{
				LoginPasswordStageType: LoginPasswordStageType{},
			},
			want: "\n  Выберите файл: \n\n\n",
		},
		{
			name: "positive test #2",
			fields: fields{
				LoginPasswordStageType: LoginPasswordStageType{},
				quitting:               true,
			},
			want: "",
		},
		{
			name: "positive test #3",
			fields: fields{
				LoginPasswordStageType: LoginPasswordStageType{},
				selectedFile:           "files\\test_file.txt",
				quitting:               false,
			},
			want: "\n  Выбранный файл: files\\test_file.txt\n\n\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &FileStageType{
				LoginPasswordStageType: tt.fields.LoginPasswordStageType,
				quitting:               tt.fields.quitting,
				selectedFile:           tt.fields.selectedFile,
			}

			if got := m.View(); got != tt.want {
				t.Errorf("View() = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestFileStageType_save(t *testing.T) {
	clientT := NewClient()
	app, _ := NewAgent()
	app.client = &clientT

	type fields struct {
		LoginPasswordStageType LoginPasswordStageType
		selectedFile           string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "positive test #1",
			fields: fields{
				LoginPasswordStageType: LoginPasswordStageType{},
			},
			want: "agent.infoMsg",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &FileStageType{
				LoginPasswordStageType: tt.fields.LoginPasswordStageType,
				selectedFile:           tt.fields.selectedFile,
			}

			os.Mkdir("files", 777)

			defer func() {
				os.RemoveAll("files")
			}()

			m.selectedFile = "files\\test_file.txt"
			m.Prepare(app)
			m.recordID = ""

			_, got := m.save()
			if fmt.Sprintf("%T", got()) != fmt.Sprintf("%s", tt.want) {
				t.Errorf("save() got = %T, want %s", got(), tt.want)
			}
		})
	}
}

func TestInfoStageType_Init(t *testing.T) {
	tests := []struct {
		name string
		want tea.Cmd
	}{
		{
			name: "positive test #1",
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &InfoStageType{}
			if got := s.Init(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Init() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInfoStageType_Prepare(t *testing.T) {
	clientT := NewClient()
	app, _ := NewAgent()
	app.client = &clientT

	type fields struct{}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			name: "positive test #1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &InfoStageType{}
			s.Prepare(app)
		})
	}
}

func TestInfoStageType_Update(t *testing.T) {
	type fields struct {
		message    string
		back       string
		backButton string
	}
	type args struct {
		msgs []tea.Msg
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name: "positive test #1",
			args: args{
				msgs: []tea.Msg{
					tea.KeyMsg{Type: tea.KeyCtrlC},
					tea.KeyMsg{Type: tea.KeyEsc},
				},
			},
			want: "tea.QuitMsg",
		},
		{
			name: "positive test #2",
			args: args{
				msgs: []tea.Msg{
					tea.KeyMsg{Type: tea.KeyEnter},
				},
			},
			want: "agent.openStage",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &InfoStageType{}

			for _, msg := range tt.args.msgs {
				_, got := m.Update(msg.(tea.KeyMsg))

				if fmt.Sprintf("%T", got()) != tt.want {
					t.Errorf("Update(%s) got = %T, want %s", msg, got(), tt.want)
				}
			}
		})
	}
}

func TestInfoStageType_View(t *testing.T) {
	type fields struct {
		message    string
		back       string
		backButton string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "positive test #1",
			fields: fields{
				message:    "Message",
				back:       "list",
				backButton: "Назад",
			},
			want: "Message\n[ Назад ]\n\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &InfoStageType{
				message:    tt.fields.message,
				back:       tt.fields.back,
				backButton: tt.fields.backButton,
			}
			if got := m.View(); got != tt.want {
				t.Errorf("View() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInitErrorModel(t *testing.T) {
	type args struct {
		error error
		back  string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "positive test #1",
			args: args{
				error: errors.New("ошибка"),
				back:  "list",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := InitErrorModel(tt.args.error, tt.args.back); !reflect.DeepEqual(got, &ErrorStageType{
				error: tt.args.error,
				back:  tt.args.back,
			}) {
				t.Error("InitErrorModel() return bad ErrorStage")
			}
		})
	}
}

func TestInitInfoModel(t *testing.T) {
	type args struct {
		message    string
		back       string
		backButton string
	}
	tests := []struct {
		name string
		args args
		want *InfoStageType
	}{
		{
			name: "positive test #1",
			args: args{
				message:    "Message",
				back:       "list",
				backButton: "Назад",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := InitInfoModel(tt.args.message, tt.args.back, tt.args.backButton); !reflect.DeepEqual(got, &InfoStageType{
				message:    tt.args.message,
				back:       tt.args.back,
				backButton: tt.args.backButton,
			}) {
				t.Error("InitInfoModel() get bad  InfoStageType")
			}
		})
	}
}

func TestListStageType_Init(t *testing.T) {
	tests := []struct {
		name string
		want tea.Cmd
	}{
		{
			name: "posisitve test #1",
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &ListStageType{}
			if got := m.Init(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Init() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestListStageType_Prepare(t *testing.T) {
	clientT := NewClient()
	app, _ := NewAgent()
	app.client = &clientT
	app.userID = "exist_user"

	type fields struct {
		userID string
	}
	type args struct {
		a *agent
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "positive test #1",
			fields: fields{
				userID: "exist_user",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &ListStageType{}
			m.Prepare(app)
		})
	}
}

func TestListStageType_Update(t *testing.T) {
	clientT := NewClient()
	app, _ := NewAgent()
	app.client = &clientT
	app.userID = "exist_user"

	type args struct {
		msgs []tea.Msg
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "posisitve test #1",
			args: args{
				msgs: []tea.Msg{
					tea.KeyMsg{Type: tea.KeyCtrlC},
					tea.KeyMsg{Type: tea.KeyEsc},
				},
			},
			want: "tea.QuitMsg",
		},
		{
			name: "positive test #2",
			args: args{
				msgs: []tea.Msg{
					tea.KeyMsg{Type: tea.KeyEnter},
				},
			},
			want: "agent.openForm",
		},
		{
			name: "positive test #3",
			args: args{
				msgs: []tea.Msg{
					tea.KeyMsg{Type: tea.KeyCtrlN},
				},
			},
			want: "agent.openStage",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &ListStageType{}

			m.Prepare(app)

			for _, msg := range tt.args.msgs {
				_, got := m.Update(msg.(tea.KeyMsg))

				if fmt.Sprintf("%T", got()) != tt.want {
					t.Errorf("Update(%s) got = %T, want %s", msg, got(), tt.want)
				}
			}
		})
	}
}

func TestListStageType_View(t *testing.T) {
	clientT := NewClient()
	app, _ := NewAgent()
	app.client = &clientT
	app.userID = "exist_user"

	tests := []struct {
		name string

		want string
	}{
		{
			name: "positive test #1",
			want: "   Мои записи                                   \n                                                \n  2 items                                       \n                                                \n                                                \n  1/2                                           \n                                                \n  ↑/k up • ↓/j down • / filter • q quit • ? more\n\n[ Ctrl+n ] - Создать новую запись\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &ListStageType{}

			m.Prepare(app)

			if got := m.View(); got != tt.want {
				t.Errorf("View() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLoginPasswordStageType_Init(t *testing.T) {
	tests := []struct {
		name string
		want tea.Cmd
	}{
		{
			name: "positive test #1",
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &LoginPasswordStageType{}

			if got := m.Init(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Init() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLoginPasswordStageType_Prepare(t *testing.T) {
	f := createTestFile()

	defer func() {
		f.Close()
		os.RemoveAll("users")
	}()

	clientT := NewClient()
	app, _ := NewAgent()
	app.client = &clientT
	app.userID = "test_user"

	type args struct {
		recordId string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "positive test #1",
			args: args{
				recordId: "exist_id",
			},
		},
		{
			name: "positive test #2",
			args: args{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &LoginPasswordStageType{}

			app.recordID = tt.args.recordId

			m.Prepare(app)
		})
	}
}

func TestLoginPasswordStageType_Update(t *testing.T) {
	f := createTestFile()

	defer func() {
		f.Close()
		os.RemoveAll("users")
	}()

	clientT := NewClient()
	app, _ := NewAgent()
	app.client = &clientT
	app.userID = "test_user"

	type args struct {
		msgs []tea.Msg
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "positive test #1",
			args: args{
				msgs: []tea.Msg{
					tea.KeyMsg{Type: tea.KeyCtrlC},
					tea.KeyMsg{Type: tea.KeyEsc},
				},
			},
			want: "tea.QuitMsg",
		},
		{
			name: "positive test #2",
			args: args{
				msgs: []tea.Msg{
					tea.KeyMsg{Type: tea.KeyCtrlB},
				},
			},
			want: "agent.openList",
		},
		{
			name: "positive test #3",
			args: args{
				msgs: []tea.Msg{
					tea.KeyMsg{Type: tea.KeyTab},
					tea.KeyMsg{Type: tea.KeyShiftTab},
					tea.KeyMsg{Type: tea.KeyEnter},
					tea.KeyMsg{Type: tea.KeyUp},
					tea.KeyMsg{Type: tea.KeyDown},
				},
			},
			want: "cursor.BlinkMsg",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &LoginPasswordStageType{}
			m.Prepare(app)

			for _, msg := range tt.args.msgs {
				_, got := m.Update(msg.(tea.KeyMsg))

				if fmt.Sprintf("%T", got()) != tt.want {
					t.Errorf("Update(%s) got = %T, want %s", msg, got(), tt.want)
				}
			}
		})
	}
}

func TestLoginPasswordStageType_View(t *testing.T) {
	f := createTestFile()

	defer func() {
		f.Close()
		os.RemoveAll("users")
	}()

	clientT := NewClient()
	app, _ := NewAgent()
	app.client = &clientT
	app.userID = "test_user"

	type agrs struct {
		recordId string
	}

	tests := []struct {
		name string
		agrs agrs
		want string
	}{
		{
			name: "positive test #1",
			agrs: agrs{
				recordId: "exist_id",
			},
			want: "> Hello World \n> Login_test \n> Login_pass \n> Description \n[ Ctrl+s ] - Сохранить\n[ Ctrl+b ] - Назад\n",
		},
		{
			name: "positive test #1",
			want: "> Название\n> Логин\n> Пароль\n> Описание\n[ Ctrl+s ] - Сохранить\n[ Ctrl+b ] - Назад\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &LoginPasswordStageType{}
			app.recordID = tt.agrs.recordId
			m.Prepare(app)

			if got := m.View(); got != tt.want {
				t.Errorf("View() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLoginPasswordStageType_getClient(t *testing.T) {
	clientT := NewClient()
	app, _ := NewAgent()
	app.client = &clientT

	tests := []struct {
		name string
		want *client.Client
	}{
		{
			name: "positive test #1",
			want: &clientT,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &LoginPasswordStageType{
				client: app.client,
			}
			if got := m.getClient(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getClient() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLoginPasswordStageType_getRecordID(t *testing.T) {
	clientT := NewClient()
	app, _ := NewAgent()
	app.client = &clientT
	app.recordID = "exist_id"

	tests := []struct {
		name string
		want string
	}{
		{
			name: "positive test #1",
			want: "exist_id",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &LoginPasswordStageType{}
			m.Prepare(app)

			if got := m.getRecordID(); got != tt.want {
				t.Errorf("getRecordID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLoginPasswordStageType_getToken(t *testing.T) {
	clientT := NewClient()
	app, _ := NewAgent()
	app.client = &clientT
	app.token = "111"

	tests := []struct {
		name string
		want string
	}{
		{
			name: "positive test #1",
			want: "111",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &LoginPasswordStageType{}
			m.Prepare(app)

			if got := m.getToken(); got != tt.want {
				t.Errorf("getToken() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLoginPasswordStageType_save(t *testing.T) {
	f := createTestFile()

	defer func() {
		f.Close()
		os.RemoveAll("users")
	}()

	clientT := NewClient()
	app, _ := NewAgent()
	app.client = &clientT
	app.userID = "test_user"

	type args struct {
		recordId string
	}

	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "positive test #1",
			args: args{
				recordId: "exist_id",
			},
			want: "agent.openList",
		},
		{
			name: "positive test #2",
			args: args{
				recordId: "error",
			},
			want: "agent.infoMsg",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &LoginPasswordStageType{}

			app.recordID = tt.args.recordId
			m.Prepare(app)

			_, got := m.save()
			if fmt.Sprintf("%T", got()) != tt.want {
				t.Errorf("save() got = %T, want %s", got(), tt.want)
			}
		})
	}
}

func TestLoginPasswordStageType_updateInputs(t *testing.T) {
	clientT := NewClient()
	app, _ := NewAgent()
	app.client = &clientT

	tests := []struct {
		name string
		want tea.Cmd
	}{
		{
			name: "positive test #1",
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &LoginPasswordStageType{}
			m.Prepare(app)

			got := m.updateInputs(tea.KeyMsg{Type: tea.KeyEnter})

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("save() got = %T, want %T", got, tt.want)
			}
		})
	}
}

func TestLoginStageType_Init(t *testing.T) {
	tests := []struct {
		name string
		want tea.Cmd
	}{
		{
			name: "positive test #1",
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &LoginStageType{}
			if got := s.Init(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Init() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLoginStageType_Prepare(t *testing.T) {
	clientT := NewClient()
	app, _ := NewAgent()
	app.client = &clientT

	tests := []struct {
		name string
	}{
		{
			name: "positive test #1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &LoginStageType{}
			s.Prepare(app)
		})
	}
}

func TestLoginStageType_Update(t *testing.T) {
	clientT := NewClient()
	app, _ := NewAgent()
	app.client = &clientT

	type args struct {
		msgs       []tea.Msg
		focusIndex int
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "positive test #1",
			args: args{
				msgs: []tea.Msg{
					tea.KeyMsg{Type: tea.KeyCtrlC},
					tea.KeyMsg{Type: tea.KeyEsc},
				},
			},
			want: "tea.QuitMsg",
		},
		{
			name: "positive test #2",
			args: args{
				msgs: []tea.Msg{
					tea.KeyMsg{Type: tea.KeyTab},
					tea.KeyMsg{Type: tea.KeyDown},
				},
				focusIndex: 1,
			},
			want: "",
		},
		{
			name: "positive test #2",
			args: args{
				msgs: []tea.Msg{
					tea.KeyMsg{Type: tea.KeyShiftTab},
					tea.KeyMsg{Type: tea.KeyUp},
				},
				focusIndex: 1,
			},
			want: "cursor.BlinkMsg",
		},
		{
			name: "positive test #3",
			args: args{
				msgs: []tea.Msg{
					tea.KeyMsg{Type: tea.KeyEnter},
				},
				focusIndex: 3,
			},
			want: "agent.openStage",
		},
		{
			name: "positive test #4",
			args: args{
				msgs: []tea.Msg{
					tea.KeyMsg{Type: tea.KeyEnter},
				},
				focusIndex: 2,
			},
			want: "agent.authSuccessMsg",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &LoginStageType{}
			m.Prepare(app)

			for _, msg := range tt.args.msgs {
				m.focusIndex = tt.args.focusIndex
				_, got := m.Update(msg.(tea.KeyMsg))

				if got == nil {
					if tt.want != "" {
						t.Errorf("Update(%s) got = nil, want %s", msg, tt.want)
					}
					continue
				}

				if fmt.Sprintf("%T", got()) != tt.want {
					t.Errorf("Update(%s) got = %T, want %s", msg, got(), tt.want)
				}
			}
		})
	}
}

func TestLoginStageType_View(t *testing.T) {
	clientT := NewClient()
	app, _ := NewAgent()
	app.client = &clientT

	tests := []struct {
		name string
		want string
	}{
		{
			name: "positive test #1",
			want: "> Логин\n> Пароль\n[ Авторизоваться ]\n[ Назад ]\n\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &LoginStageType{}
			m.Prepare(app)

			if got := m.View(); got != tt.want {
				t.Errorf("View() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLoginStageType_process(t *testing.T) {
	clientT := NewClient()
	app, _ := NewAgent()
	app.client = &clientT

	type args struct {
		userName string
	}

	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "positive test #1",
			args: args{
				userName: "Пользователь",
			},
			want: "agent.authSuccessMsg",
		},
		{
			name: "negative test #2",
			args: args{
				userName: "error",
			},
			want: "agent.infoMsg",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &LoginStageType{}
			m.Prepare(app)

			m.inputs[0].SetValue(tt.args.userName)
			m.inputs[1].SetValue("Пароль")

			_, got := m.process()

			if fmt.Sprintf("%T", got()) != tt.want {
				t.Errorf("process() got = %T, want %s", got(), tt.want)
			}
		})
	}
}

func TestLoginStageType_updateInputs(t *testing.T) {
	clientT := NewClient()
	app, _ := NewAgent()
	app.client = &clientT

	type args struct {
		msg tea.Msg
	}
	tests := []struct {
		name string
		args args
		want tea.Cmd
	}{
		{
			name: "positive test #1",
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &LoginStageType{}
			m.Prepare(app)
			if got := m.updateInputs(tt.args.msg); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("updateInputs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewAgent(t *testing.T) {
	tests := []struct {
		name          string
		wantStagesLen int
		wantErr       bool
	}{
		{
			name:          "positive test #1",
			wantStagesLen: 11,
			wantErr:       false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewAgent()
			if (err != nil) != tt.wantErr {
				t.Errorf("NewAgent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got.Stages) != tt.wantStagesLen {
				t.Errorf("NewAgent() got Stages = %d, want %v", len(got.Stages), tt.wantStagesLen)
			}
		})
	}
}

func TestOperationListStageType_Init(t *testing.T) {
	tests := []struct {
		name string
		want tea.Cmd
	}{
		{
			name: "positive test #1",
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &OperationListStageType{}
			if got := s.Init(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Init() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOperationListStageType_Prepare(t *testing.T) {
	clientT := NewClient()
	app, _ := NewAgent()
	app.client = &clientT

	tests := []struct {
		name              string
		wantOperationLlen int
	}{
		{
			name:              "positive test #1",
			wantOperationLlen: 4,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &OperationListStageType{}
			s.Prepare(app)

			if len(s.operations) != tt.wantOperationLlen {
				t.Errorf("Prepare() set wrong length of operations")
			}
		})
	}
}

func TestOperationListStageType_Update(t *testing.T) {
	clientT := NewClient()
	app, _ := NewAgent()
	app.client = &clientT

	type args struct {
		msgs []tea.Msg
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "positive test #1",
			args: args{
				msgs: []tea.Msg{
					tea.KeyMsg{Type: tea.KeyCtrlC},
					tea.KeyMsg{Type: tea.KeyEsc},
				},
			},
			want: "tea.QuitMsg",
		},
		{
			name: "positive test #2",
			args: args{
				msgs: []tea.Msg{
					tea.KeyMsg{Type: tea.KeyEnter},
					tea.KeyMsg{Type: tea.KeySpace},
				},
			},
			want: "agent.openForm",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &OperationListStageType{}
			s.Prepare(app)

			for _, msg := range tt.args.msgs {
				_, got := s.Update(msg.(tea.KeyMsg))

				if fmt.Sprintf("%T", got()) != tt.want {
					t.Errorf("Update(%s) got = %T, want %s", msg, got(), tt.want)
				}
			}
		})
	}
}

func TestOperationListStageType_View(t *testing.T) {
	clientT := NewClient()
	app, _ := NewAgent()
	app.client = &clientT

	tests := []struct {
		name string
		want string
	}{
		{
			name: "positive test #1",
			want: "Что будем создавать?\n\n> [x] Логин/пароль\n  [ ] Текст\n  [ ] Файл\n  [ ] Банковская карта\n\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &OperationListStageType{}
			s.Prepare(app)

			if got := s.View(); got != tt.want {
				t.Errorf("View() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRegisterStageType_Init(t *testing.T) {
	tests := []struct {
		name string
		want tea.Cmd
	}{
		{
			name: "positive test #1",
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &RegisterStageType{}
			if got := s.Init(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Init() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRegisterStageType_Prepare(t *testing.T) {
	clientT := NewClient()
	app, _ := NewAgent()
	app.client = &clientT

	tests := []struct {
		name string
	}{
		{
			name: "positive test #1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &RegisterStageType{}
			s.Prepare(app)
		})
	}
}

func TestRegisterStageType_Update(t *testing.T) {
	f := createTestFile()

	defer func() {
		f.Close()
		os.RemoveAll("users")
	}()

	clientT := NewClient()
	app, _ := NewAgent()
	app.client = &clientT

	type args struct {
		msgs       []tea.Msg
		focusIndex int
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "positive test #1",
			args: args{
				msgs: []tea.Msg{
					tea.KeyMsg{Type: tea.KeyCtrlC},
					tea.KeyMsg{Type: tea.KeyEsc},
				},
			},
			want: "tea.QuitMsg",
		},
		{
			name: "positive test #2",
			args: args{
				msgs: []tea.Msg{
					tea.KeyMsg{Type: tea.KeyEnter},
				},
				focusIndex: 3,
			},
			want: "agent.openStage",
		},
		{
			name: "positive test #3",
			args: args{
				msgs: []tea.Msg{
					tea.KeyMsg{Type: tea.KeyEnter},
				},
				focusIndex: 2,
			},
			want: "agent.authSuccessMsg",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &RegisterStageType{}
			m.Prepare(app)

			for _, msg := range tt.args.msgs {
				m.focusIndex = tt.args.focusIndex

				_, got := m.Update(msg.(tea.KeyMsg))

				if fmt.Sprintf("%T", got()) != tt.want {
					t.Errorf("Update(%s) got = %T, want %s", msg, got(), tt.want)
				}
			}
		})
	}
}

func TestRegisterStageType_View(t *testing.T) {
	clientT := NewClient()
	app, _ := NewAgent()
	app.client = &clientT

	tests := []struct {
		name string
		want string
	}{
		{
			name: "positive test #1",
			want: "> Логин\n> Пароль\n[ Зарегистрироваться ]\n[ Назад ]\n\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &RegisterStageType{}
			m.Prepare(app)

			if got := m.View(); got != tt.want {
				t.Errorf("View() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRegisterStageType_process(t *testing.T) {
	os.Mkdir("users", 777)
	defer func() {
		os.RemoveAll("users")
	}()

	clientT := NewClient()
	app, _ := NewAgent()
	app.client = &clientT

	type args struct {
		userName string
	}

	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "positive test #1",
			args: args{
				userName: "Пользователь",
			},
			want: "agent.authSuccessMsg",
		},
		{
			name: "negative test #2",
			args: args{
				userName: "error",
			},
			want: "agent.infoMsg",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &RegisterStageType{}
			m.Prepare(app)

			m.inputs[0].SetValue(tt.args.userName)
			m.inputs[1].SetValue("Пароль")

			_, got := m.process()

			if fmt.Sprintf("%T", got()) != tt.want {
				t.Errorf("process() got = %T, want %s", got(), tt.want)
			}
		})
	}
}

func TestRegisterStageType_updateInputs(t *testing.T) {
	clientT := NewClient()
	app, _ := NewAgent()
	app.client = &clientT

	type args struct {
		msg tea.Msg
	}
	tests := []struct {
		name string
		args args
		want tea.Cmd
	}{
		{
			name: "positive test #1",
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &RegisterStageType{}
			m.Prepare(app)
			if got := m.updateInputs(tt.args.msg); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("updateInputs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStartStageType_Init(t *testing.T) {
	tests := []struct {
		name string
		want tea.Cmd
	}{
		{
			name: "positive test #1",
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &StartStageType{}
			if got := s.Init(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Init() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStartStageType_Prepare(t *testing.T) {
	clientT := NewClient()
	app, _ := NewAgent()
	app.client = &clientT

	tests := []struct {
		name string
	}{
		{
			name: "positive test #1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &StartStageType{}
			s.Prepare(app)
		})
	}
}

func TestStartStageType_Update(t *testing.T) {
	clientT := NewClient()
	app, _ := NewAgent()
	app.client = &clientT

	type args struct {
		msgs []tea.Msg
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "positive test #1",
			args: args{
				msgs: []tea.Msg{
					tea.KeyMsg{Type: tea.KeyCtrlC},
					tea.KeyMsg{Type: tea.KeyEsc},
				},
			},
			want: "tea.QuitMsg",
		},
		{
			name: "positive test #2",
			args: args{
				msgs: []tea.Msg{
					tea.KeyMsg{Type: tea.KeyCtrlR},
					tea.KeyMsg{Type: tea.KeyCtrlL},
				},
			},
			want: "agent.openStage",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &StartStageType{}
			s.Prepare(app)

			for _, msg := range tt.args.msgs {
				_, got := s.Update(msg.(tea.KeyMsg))

				if fmt.Sprintf("%T", got()) != tt.want {
					t.Errorf("Update(%s) got = %T, want %s", msg, got(), tt.want)
				}
			}
		})
	}
}

func TestStartStageType_View(t *testing.T) {
	clientT := NewClient()
	app, _ := NewAgent()
	app.client = &clientT

	tests := []struct {
		name string
		want string
	}{
		{
			name: "positive test #1",
			want: "Войдите в систему!\n\n[ Ctrl+r ] - Зарегистрироваться\n[ Ctrl+l ] - Авторизоваться\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &StartStageType{}
			s.Prepare(app)

			if got := s.View(); got != tt.want {
				t.Errorf("View() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTextStageType_Prepare(t *testing.T) {
	clientT := NewClient()
	app, _ := NewAgent()
	app.client = &clientT

	tests := []struct {
		name string
	}{
		{
			name: "positive test #1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &TextStageType{}
			m.Prepare(app)
		})
	}
}

func TestTextStageType_Update(t *testing.T) {
	clientT := NewClient()
	app, _ := NewAgent()
	app.client = &clientT

	type args struct {
		msgs []tea.Msg
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "positive test #1",
			args: args{
				msgs: []tea.Msg{
					tea.KeyMsg{Type: tea.KeyCtrlC},
				},
			},
			want: "tea.QuitMsg",
		},
		{
			name: "positive test #2",
			args: args{
				msgs: []tea.Msg{
					tea.KeyMsg{Type: tea.KeyCtrlS},
					tea.KeyMsg{Type: tea.KeyCtrlB},
				},
			},
			want: "agent.openList",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &TextStageType{}
			m.Prepare(app)

			for _, msg := range tt.args.msgs {
				_, got := m.Update(msg.(tea.KeyMsg))

				if fmt.Sprintf("%T", got()) != tt.want {
					t.Errorf("Update(%s) got = %T, want %s", msg, got(), tt.want)
				}
			}
		})
	}
}

func TestTextStageType_View(t *testing.T) {
	clientT := NewClient()
	app, _ := NewAgent()
	app.client = &clientT

	tests := []struct {
		name string
		want string
	}{
		{
			name: "positive test 31",
			want: "> Название\n> Описание\n┃   1                                   \n┃   ~                                   \n┃   ~                                   \n┃   ~                                   \n┃   ~                                   \n┃   ~                                   \n\n[ Ctrl+s ] - Сохранить\n[ Ctrl+b ] - Назад\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &TextStageType{}
			m.Prepare(app)

			if got := m.View(); got != tt.want {
				t.Errorf("View() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTextStageType_getRecordDataFromServer(t *testing.T) {
	f := createTestFile()

	defer func() {
		f.Close()
		os.RemoveAll("users")
	}()

	clientT := NewClient()
	app, _ := NewAgent()
	app.client = &clientT

	type args struct {
		data models.Data
		body string
	}
	tests := []struct {
		name     string
		args     args
		wantData models.Data
		wantBody string
		wantErr  bool
	}{
		{
			name: "positive test #1",
			args: args{
				data: models.Data{
					ID:     "exist_id",
					UserID: "test_user",
					Type:   "text",
				},
			},
			wantData: models.Data{
				ID:   "exist_id",
				Body: []byte("Тут текст"),
			},
			wantBody: "{\"Login\":\"Login_test\",\"Password\":\"Login_pass\"}",
			wantErr:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &TextStageType{}
			app.recordID = tt.args.data.ID
			app.userID = tt.args.data.UserID

			m.Prepare(app)

			gotData, gotBody, err := m.getRecordDataFromServer(tt.args.data, "")

			if (err != nil) != tt.wantErr {
				t.Errorf("getRecordDataFromServer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if string(gotData.Body) != tt.wantBody {
				t.Errorf("getRecordDataFromServer() gotData = %v, want %v", gotData.Body, tt.wantBody)
			}
			if gotBody != tt.wantBody {
				t.Errorf("getRecordDataFromServer() gotBody = %v, want %v", gotBody, tt.wantBody)
			}
		})
	}
}

func TestTextStageType_save(t *testing.T) {
	f := createTestFile()

	defer func() {
		f.Close()
		os.RemoveAll("users")
	}()

	clientT := NewClient()
	app, _ := NewAgent()
	app.client = &clientT
	app.userID = "test_user"

	type args struct {
		recordId string
	}

	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "positive test #1",
			args: args{
				recordId: "exist_id",
			},
			want: "agent.openList",
		},
		{
			name: "negetive test #2",
			args: args{
				recordId: "error",
			},
			want: "agent.infoMsg",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &TextStageType{}

			app.recordID = tt.args.recordId
			m.Prepare(app)

			_, got := m.save()
			if fmt.Sprintf("%T", got()) != tt.want {
				t.Errorf("save() got = %T, want %s", got(), tt.want)
			}
		})
	}
}

//func Test_agent_CatchTerminateSignal(t *testing.T) {
//	tests := []struct {
//		name    string
//		wantErr bool
//	}{
//		{
//			name:    "positive test #1",
//			wantErr: false,
//		},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			a := &agent{}
//
//			go func() {
//				timer := time.Timer{}
//			}()
//			if err := a.CatchTerminateSignal(); (err != nil) != tt.wantErr {
//				t.Errorf("CatchTerminateSignal() error = %v, wantErr %v", err, tt.wantErr)
//			}
//		})
//	}
//}

func Test_agent_Close(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "positive test #1",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a, _ := NewAgent()
			if err := a.Close(); (err != nil) != tt.wantErr {
				t.Errorf("Close() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_agent_Init(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{
			name: "positive test #1",
			want: "filepicker.readDirMsg",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a, _ := NewAgent()

			got := a.Init()

			if fmt.Sprintf("%T", got()) != tt.want {
				t.Errorf("Init() = %T, want %s", got(), tt.want)
			}
		})
	}
}

func Test_agent_Prepare(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "positive test #1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a, _ := NewAgent()
			a.Prepare(a)
		})
	}
}

func Test_agent_Update(t *testing.T) {
	clientT := NewClient()
	app, _ := NewAgent()
	app.client = &clientT

	type args struct {
		msg tea.Msg
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "positive test #1",
			args: args{
				msg: tea.KeyMsg{Type: tea.KeyCtrlC},
			},
			want: "tea.QuitMsg",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, got := app.Update(tt.args.msg)

			if fmt.Sprintf("%T", got()) != tt.want {
				t.Errorf("Update(%s) got = %T, want %s", tt.args.msg, got(), tt.want)
			}
		})
	}
}

func Test_agent_View(t *testing.T) {
	clientT := NewClient()
	app, _ := NewAgent()
	app.client = &clientT

	tests := []struct {
		name string
		want string
	}{
		{
			name: "positive test #1",
			want: "Войдите в систему!\n\n[ Ctrl+r ] - Зарегистрироваться\n[ Ctrl+l ] - Авторизоваться\n[ Ctrl+c ] - Выход из программы\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := app.View(); got != tt.want {
				t.Errorf("View() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_ccnValidator(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ccnValidator(tt.args.s); (err != nil) != tt.wantErr {
				t.Errorf("ccnValidator() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_clearErrorAfter(t *testing.T) {
	type args struct {
		t time.Duration
	}
	tests := []struct {
		name string
		args args
		want tea.Cmd
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := clearErrorAfter(tt.args.t); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("clearErrorAfter() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_cvvValidator(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := cvvValidator(tt.args.s); (err != nil) != tt.wantErr {
				t.Errorf("cvvValidator() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_errMsg_Error(t *testing.T) {
	type fields struct {
		error error
		back  string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := errMsg{
				error: tt.fields.error,
				back:  tt.fields.back,
			}
			if got := e.Error(); got != tt.want {
				t.Errorf("Error() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_expValidator(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := expValidator(tt.args.s); (err != nil) != tt.wantErr {
				t.Errorf("expValidator() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

//func Test_getRecordDataFromServer(t *testing.T) {
//	type args[T any] struct {
//		m    FormStageType
//		data models.Data
//		body T
//	}
//	type testCase[T any] struct {
//		name    string
//		args    args[T]
//		want    models.Data
//		want1   T
//		wantErr bool
//	}
//	tests := []testCase[ /* TODO: Insert concrete types here */ ]{
//		// TODO: Add test cases.
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			got, got1, err := getRecordDataFromServer(tt.args.m, tt.args.data, tt.args.body)
//			if (err != nil) != tt.wantErr {
//				t.Errorf("getRecordDataFromServer() error = %v, wantErr %v", err, tt.wantErr)
//				return
//			}
//			if !reflect.DeepEqual(got, tt.want) {
//				t.Errorf("getRecordDataFromServer() got = %v, want %v", got, tt.want)
//			}
//			if !reflect.DeepEqual(got1, tt.want1) {
//				t.Errorf("getRecordDataFromServer() got1 = %v, want %v", got1, tt.want1)
//			}
//		})
//	}
//}

func Test_listItem_Description(t *testing.T) {
	type fields struct {
		id          string
		title       string
		desc        string
		type_record string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "positive test #1",
			fields: fields{
				id:          "111",
				title:       "Название",
				desc:        "Описание",
				type_record: "text",
			},
			want: "Описание",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := listItem{
				id:          tt.fields.id,
				title:       tt.fields.title,
				desc:        tt.fields.desc,
				type_record: tt.fields.type_record,
			}
			if got := i.Description(); got != tt.want {
				t.Errorf("Description() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_listItem_FilterValue(t *testing.T) {
	type fields struct {
		id          string
		title       string
		desc        string
		type_record string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "positive test #1",
			fields: fields{
				id:          "111",
				title:       "Название",
				desc:        "Описание",
				type_record: "text",
			},
			want: "Название",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := listItem{
				id:          tt.fields.id,
				title:       tt.fields.title,
				desc:        tt.fields.desc,
				type_record: tt.fields.type_record,
			}
			if got := i.FilterValue(); got != tt.want {
				t.Errorf("FilterValue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_listItem_Title(t *testing.T) {
	type fields struct {
		id          string
		title       string
		desc        string
		type_record string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "positive test #1",
			fields: fields{
				id:          "111",
				title:       "Название",
				desc:        "Описание",
				type_record: "text",
			},
			want: "Название",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := listItem{
				id:          tt.fields.id,
				title:       tt.fields.title,
				desc:        tt.fields.desc,
				type_record: tt.fields.type_record,
			}
			if got := i.Title(); got != tt.want {
				t.Errorf("Title() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_newClient(t *testing.T) {
	tests := []struct {
		name    string
		want    client.Client
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := newClient()
			if (err != nil) != tt.wantErr {
				t.Errorf("newClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newClient() got = %v, want %v", got, tt.want)
			}
		})
	}
}
