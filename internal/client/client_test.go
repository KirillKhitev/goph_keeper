package client

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"github.com/KirillKhitev/goph_keeper/internal/api"
	"github.com/KirillKhitev/goph_keeper/internal/auth"
	"github.com/KirillKhitev/goph_keeper/internal/gzip"
	"github.com/KirillKhitev/goph_keeper/internal/handlers"
	"github.com/KirillKhitev/goph_keeper/internal/models"
	"github.com/KirillKhitev/goph_keeper/internal/store"
	"github.com/go-resty/resty/v2"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewRestyClient(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "positive test 1",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewRestyClient()
			if (err != nil) != tt.wantErr {
				t.Errorf("NewRestyClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestRestyClient_Close(t *testing.T) {
	type fields struct {
		client *resty.Client
		userID string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "positive test 1",
			fields: fields{
				client: resty.New(),
				userID: "111",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &RestyClient{
				client: tt.fields.client,
				userID: tt.fields.userID,
			}
			if err := c.Close(); (err != nil) != tt.wantErr {
				t.Errorf("Close() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRestyClient_Compress(t *testing.T) {
	type fields struct {
		client *resty.Client
		userID string
	}
	type args struct {
		data []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "positive test 1",
			fields: fields{
				client: resty.New(),
				userID: "111",
			},
			args: args{
				data: []byte("hello world"),
			},
			want:    "H4sIAAAAAAAA/8pIzcnJVyjPL8pJAQQAAP//hRFKDQsAAAA=",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &RestyClient{
				client: tt.fields.client,
				userID: tt.fields.userID,
			}
			got, err := c.Compress(tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("Compress() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			gotStr := base64.StdEncoding.EncodeToString(got)
			if gotStr != tt.want {
				t.Errorf("Compress() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRestyClient_SetUserID(t *testing.T) {
	type fields struct {
		client *resty.Client
		userID string
	}
	type args struct {
		userID string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "positive test 1",
			fields: fields{
				client: resty.New(),
				userID: "222",
			},
			args: args{
				userID: "111",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &RestyClient{
				client: tt.fields.client,
				userID: tt.fields.userID,
			}
			c.SetUserID(tt.args.userID)

			if c.userID != tt.args.userID {
				t.Errorf("UserID = %s, want %s", c.userID, tt.args.userID)
			}
		})
	}
}

func TestRestyClient_prepareDataForSend(t *testing.T) {
	type fields struct {
		client *resty.Client
		userID string
	}
	type args struct {
		data []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "positive test 1",
			fields: fields{
				client: resty.New(),
				userID: "222",
			},
			args: args{
				data: []byte("hello world"),
			},
			want:    "H4sIAAAAAAAA/8pIzcnJVyjPL8pJAQQAAP//hRFKDQsAAAA=",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &RestyClient{
				client: tt.fields.client,
				userID: tt.fields.userID,
			}
			got, err := c.prepareDataForSend(tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("prepareDataForSend() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			gotStr := base64.StdEncoding.EncodeToString(got)
			if gotStr != tt.want {
				t.Errorf("prepareDataForSend() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func createTestServer(t *testing.T) *httptest.Server {
	ts := createServer(func(w http.ResponseWriter, r *http.Request) {
		t.Logf("Method: %v", r.Method)
		t.Logf("Path: %v", r.URL.Path)

		st := store.GetTestStore()

		if r.Method == http.MethodPost {
			switch r.URL.Path {
			case "/api/data/get":
				userID, response := api.GetUserFromAuthHeader(w, r, st)
				if response.Code != 0 {
					handlers.SendResponse(response, w)
					return
				}

				response = api.GetData(w, r, userID, st)

				handlers.SendResponse(response, w)
			case "/api/data/list":
				userID, response := api.GetUserFromAuthHeader(w, r, st)
				if response.Code != 0 {
					handlers.SendResponse(response, w)
					return
				}

				response = api.ListData(w, r, userID, st)

				handlers.SendResponse(response, w)
			case "/api/user/login":
				response := api.Login(w, r, st)

				handlers.SendResponse(response, w)
			}

		}

		if r.Method == http.MethodPut {
			switch r.URL.Path {
			case "/api/data/update":

			}
		}
	})

	return ts
}

func createServer(fn func(w http.ResponseWriter, r *http.Request)) *httptest.Server {
	handler := gzip.Middleware(http.HandlerFunc(fn))
	return httptest.NewServer(handler)
}

func TestRestyClient_Get(t *testing.T) {
	user := models.User{
		ID: "111",
	}

	token, _ := auth.BuildJWTString(user)

	data := models.Data{
		ID:   "11122333",
		Name: []byte("Hello World"),
	}

	body, _ := json.Marshal(data)

	type fields struct {
		client *resty.Client
	}
	type args struct {
		headers map[string]string
		data    []byte
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   APIServiceResult
	}{
		{
			name: "positive test 1",
			fields: fields{
				client: resty.New(),
			},
			args: args{
				headers: map[string]string{
					"Authorization": token,
				},
				data: body,
			},
			want: APIServiceResult{
				Code: http.StatusOK,
			},
		},
		{
			name: "negative test #2",
			fields: fields{
				client: resty.New(),
			},
			args: args{
				data: body,
			},
			want: APIServiceResult{
				Code: http.StatusUnauthorized,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &RestyClient{
				client: tt.fields.client,
				userID: user.ID,
			}

			ts := createTestServer(t)
			defer ts.Close()

			got := c.Get(context.TODO(), ts.URL+"/api/data/get", tt.args.headers, tt.args.data)

			if got.Code != tt.want.Code {
				t.Errorf("Get() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRestyClient_List(t *testing.T) {
	user := models.User{
		ID: "111",
	}

	token, _ := auth.BuildJWTString(user)

	type fields struct {
		client *resty.Client
	}
	type args struct {
		headers map[string]string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   APIServiceResult
	}{
		{
			name: "positive test #1",
			fields: fields{
				client: resty.New(),
			},
			args: args{
				headers: map[string]string{
					"Authorization": token,
				},
			},
			want: APIServiceResult{
				Code: http.StatusOK,
			},
		},
		{
			name: "negative test #2",
			fields: fields{
				client: resty.New(),
			},
			want: APIServiceResult{
				Code: http.StatusUnauthorized,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &RestyClient{
				client: tt.fields.client,
				userID: user.ID,
			}

			ts := createTestServer(t)
			defer ts.Close()

			got := c.List(context.TODO(), ts.URL+"/api/data/list", tt.args.headers)

			if got.Code != tt.want.Code {
				t.Errorf("List() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRestyClient_Login(t *testing.T) {
	data := auth.AuthorizingData{
		UserName: "Exist User",
		Password: "Пароль",
	}

	body, _ := json.Marshal(data)

	type fields struct {
		client *resty.Client
	}
	type args struct {
		data []byte
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   APIServiceResult
	}{
		{
			name: "positive test #1",
			fields: fields{
				client: resty.New(),
			},
			args: args{
				data: body,
			},
			want: APIServiceResult{
				Code: http.StatusOK,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &RestyClient{
				client: tt.fields.client,
			}

			ts := createTestServer(t)
			defer ts.Close()

			got := c.Login(context.TODO(), ts.URL+"/api/user/login", body)

			if got.Code != tt.want.Code {
				t.Errorf("Login() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRestyClient_Register(t *testing.T) {
	data := auth.AuthorizingData{
		UserName: "Пользователь1",
		Password: "Пароль",
	}

	body, _ := json.Marshal(data)

	type fields struct {
		client *resty.Client
	}
	type args struct {
		data []byte
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   APIServiceResult
	}{
		{
			name: "positive test #1",
			fields: fields{
				client: resty.New(),
			},
			args: args{
				data: body,
			},
			want: APIServiceResult{
				Code: http.StatusOK,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &RestyClient{
				client: tt.fields.client,
			}

			ts := createTestServer(t)
			defer ts.Close()

			got := c.Register(context.TODO(), ts.URL+"/api/user/register", body)

			if got.Code != tt.want.Code {
				t.Errorf("Register() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRestyClient_Update(t *testing.T) {
	user := models.User{
		ID: "111",
	}

	token, _ := auth.BuildJWTString(user)
	data := models.Data{
		ID:   "11122333",
		Name: []byte("Hello World"),
	}

	body, _ := json.Marshal(data)

	type fields struct {
		client *resty.Client
	}
	type args struct {
		headers map[string]string
		data    []byte
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   APIServiceResult
	}{
		{
			name: "positive test #1",
			fields: fields{
				client: resty.New(),
			},
			args: args{
				data: body,
				headers: map[string]string{
					"Authorization": token,
				},
			},
			want: APIServiceResult{
				Code: http.StatusOK,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &RestyClient{
				client: tt.fields.client,
				userID: user.ID,
			}

			ts := createTestServer(t)
			defer ts.Close()

			got := c.Update(context.TODO(), ts.URL+"/api/data/update", tt.args.headers, body)

			if got.Code != tt.want.Code {
				t.Errorf("Update() = %v, want %v", got, tt.want)
			}
		})
	}
}
