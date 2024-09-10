package handlers

import (
	"bytes"
	"encoding/json"
	"github.com/KirillKhitev/goph_keeper/internal/api"
	"github.com/KirillKhitev/goph_keeper/internal/auth"
	"github.com/KirillKhitev/goph_keeper/internal/models"
	"github.com/KirillKhitev/goph_keeper/internal/store"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestGet_ServeHTTP(t *testing.T) {
	user := models.User{
		ID: "111",
	}

	token, _ := auth.BuildJWTString(user)

	data := models.Data{
		ID:   "11122333",
		Name: []byte("Hello World"),
	}

	body, _ := json.Marshal(data)
	myReader := bytes.NewReader(body)

	r1 := httptest.NewRequest(http.MethodPost, "/api/data/get", myReader)
	r1.Header.Set("Authorization", token)

	r2 := httptest.NewRequest(http.MethodPost, "/api/data/get", myReader)
	r3 := httptest.NewRequest(http.MethodGet, "/api/data/get", myReader)

	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	tests := []struct {
		name string
		ch   Get
		args args
		want int
	}{
		{
			name: "positive test #1",
			ch: Get{
				Store: store.GetTestStore(),
			},
			args: args{
				w: httptest.NewRecorder(),
				r: r1,
			},
			want: 200,
		},
		{
			name: "negative test #2",
			ch: Get{
				Store: store.GetTestStore(),
			},
			args: args{
				w: httptest.NewRecorder(),
				r: r2,
			},
			want: 401,
		},
		{
			name: "negative test #3",
			ch: Get{
				Store: store.GetTestStore(),
			},
			args: args{
				w: httptest.NewRecorder(),
				r: r3,
			},
			want: 405,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.ch.ServeHTTP(tt.args.w, tt.args.r)

			code := tt.args.w.(*httptest.ResponseRecorder).Code

			if code != tt.want {
				t.Errorf("Response code = %d, want %d", code, tt.want)
				return
			}
		})
	}
}

func TestList_ServeHTTP(t *testing.T) {
	user := models.User{
		ID: "111",
	}

	token, _ := auth.BuildJWTString(user)

	r1 := httptest.NewRequest(http.MethodPost, "/api/data/list", nil)
	r1.Header.Set("Authorization", token)

	r2 := httptest.NewRequest(http.MethodPost, "/api/data/list", nil)
	r3 := httptest.NewRequest(http.MethodGet, "/api/data/list", nil)

	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	tests := []struct {
		name string
		ch   List
		args args
		want int
	}{
		{
			name: "positive test #1",
			ch: List{
				Store: store.GetTestStore(),
			},
			args: args{
				w: httptest.NewRecorder(),
				r: r1,
			},
			want: 200,
		},
		{
			name: "negative test #2",
			ch: List{
				Store: store.GetTestStore(),
			},
			args: args{
				w: httptest.NewRecorder(),
				r: r2,
			},
			want: 401,
		},
		{
			name: "negative test #3",
			ch: List{
				Store: store.GetTestStore(),
			},
			args: args{
				w: httptest.NewRecorder(),
				r: r3,
			},
			want: 405,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.ch.ServeHTTP(tt.args.w, tt.args.r)

			code := tt.args.w.(*httptest.ResponseRecorder).Code

			if code != tt.want {
				t.Errorf("Response code = %d, want %d", code, tt.want)
				return
			}
		})
	}
}

func TestLogin_ServeHTTP(t *testing.T) {
	data := auth.AuthorizingData{
		UserName: "Exist User",
		Password: "Пароль",
	}

	body, _ := json.Marshal(data)
	myReader := bytes.NewReader(body)

	r1 := httptest.NewRequest(http.MethodPost, "/api/user/login", myReader)
	r2 := httptest.NewRequest(http.MethodGet, "/api/user/login", myReader)

	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	tests := []struct {
		name string
		ch   Login
		args args
		want int
	}{
		{
			name: "positive test #1",
			ch: Login{
				Store: store.GetTestStore(),
			},
			args: args{
				w: httptest.NewRecorder(),
				r: r1,
			},
			want: 200,
		},
		{
			name: "negative test #2",
			ch: Login{
				Store: store.GetTestStore(),
			},
			args: args{
				w: httptest.NewRecorder(),
				r: r2,
			},
			want: 405,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.ch.ServeHTTP(tt.args.w, tt.args.r)

			code := tt.args.w.(*httptest.ResponseRecorder).Code

			if code != tt.want {
				t.Errorf("Response code = %d, want %d", code, tt.want)
				return
			}
		})
	}
}

func TestRegister_ServeHTTP(t *testing.T) {
	data := auth.AuthorizingData{
		UserName: "Пользователь1",
		Password: "Пароль",
	}

	body, _ := json.Marshal(data)
	myReader := bytes.NewReader(body)

	r1 := httptest.NewRequest(http.MethodPost, "/api/user/register", myReader)
	r2 := httptest.NewRequest(http.MethodGet, "/api/user/register", myReader)

	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	tests := []struct {
		name string
		ch   Register
		args args
		want int
	}{
		{
			name: "positive test #1",
			ch: Register{
				Store: store.GetTestStore(),
			},
			args: args{
				w: httptest.NewRecorder(),
				r: r1,
			},
			want: 200,
		},
		{
			name: "negative test #2",
			ch: Register{
				Store: store.GetTestStore(),
			},
			args: args{
				w: httptest.NewRecorder(),
				r: r2,
			},
			want: 405,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.ch.ServeHTTP(tt.args.w, tt.args.r)

			code := tt.args.w.(*httptest.ResponseRecorder).Code

			if code != tt.want {
				t.Errorf("Response code = %d, want %d", code, tt.want)
				return
			}
		})
	}
}

func TestUpdate_ServeHTTP(t *testing.T) {
	user := models.User{
		ID: "111",
	}

	token, _ := auth.BuildJWTString(user)
	data := models.Data{
		ID:   "11122333",
		Name: []byte("Hello World"),
	}

	body, _ := json.Marshal(data)
	myReader := bytes.NewReader(body)

	r1 := httptest.NewRequest(http.MethodPut, "/api/data/list", myReader)
	r1.Header.Set("Authorization", token)

	r2 := httptest.NewRequest(http.MethodPut, "/api/data/list", myReader)
	r3 := httptest.NewRequest(http.MethodGet, "/api/data/list", myReader)

	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	tests := []struct {
		name string
		ch   Update
		args args
		want int
	}{
		{
			name: "positive test #1",
			ch: Update{
				Store: store.GetTestStore(),
			},
			args: args{
				w: httptest.NewRecorder(),
				r: r1,
			},
			want: 200,
		},
		{
			name: "negative test #2",
			ch: Update{
				Store: store.GetTestStore(),
			},
			args: args{
				w: httptest.NewRecorder(),
				r: r2,
			},
			want: 401,
		},
		{
			name: "negative test #3",
			ch: Update{
				Store: store.GetTestStore(),
			},
			args: args{
				w: httptest.NewRecorder(),
				r: r3,
			},
			want: 405,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.ch.ServeHTTP(tt.args.w, tt.args.r)

			code := tt.args.w.(*httptest.ResponseRecorder).Code

			if code != tt.want {
				t.Errorf("Response code = %d, want %d", code, tt.want)
				return
			}
		})
	}
}

func Test_sendResponse(t *testing.T) {
	type args struct {
		res    api.ResponseType
		writer http.ResponseWriter
	}
	tests := []struct {
		name     string
		args     args
		wantCode int
		wantBody []byte
	}{
		{
			name: "positive test #1",
			args: args{
				res: api.ResponseType{
					Code: 200,
					Body: []byte("Hello World"),
				},
				writer: httptest.NewRecorder(),
			},
			wantCode: 200,
			wantBody: []byte("Hello World"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SendResponse(tt.args.res, tt.args.writer)

			code := tt.args.writer.(*httptest.ResponseRecorder).Code
			body := tt.args.writer.(*httptest.ResponseRecorder).Body.Bytes()

			if code != tt.wantCode {
				t.Errorf("Response code = %d, want %d", code, tt.wantCode)
			}

			if !reflect.DeepEqual(body, tt.wantBody) {
				t.Errorf("Response body = %v, want %v", body, tt.wantBody)
			}
		})
	}
}
