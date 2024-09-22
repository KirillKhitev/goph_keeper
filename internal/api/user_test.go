package api

import (
	"bytes"
	"encoding/json"
	"github.com/KirillKhitev/goph_keeper/internal/auth"
	"github.com/KirillKhitev/goph_keeper/internal/models"
	"github.com/KirillKhitev/goph_keeper/internal/store"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestGetUserFromAuthHeader(t *testing.T) {
	type args struct {
		w          http.ResponseWriter
		user       models.User
		s          store.Store
		needHeader bool
	}
	tests := []struct {
		name   string
		args   args
		wantId string
		want   ResponseType
	}{
		{
			name: "positive test #1",
			args: args{
				w: httptest.NewRecorder(),
				user: models.User{
					ID: "111",
				},
				s:          store.GetTestStore(),
				needHeader: true,
			},
			want:   ResponseType{},
			wantId: "111",
		},
		{
			name: "negative test #2",
			args: args{
				w: httptest.NewRecorder(),
				s: store.GetTestStore(),
				user: models.User{
					ID: "111",
				},
				needHeader: false,
			},
			want: ResponseType{
				Code: http.StatusUnauthorized,
			},
			wantId: "",
		},
		{
			name: "negative test #3",
			args: args{
				w:          httptest.NewRecorder(),
				s:          store.GetTestStore(),
				needHeader: false,
			},
			want: ResponseType{
				Code: http.StatusUnauthorized,
			},
			wantId: "",
		},
		{
			name: "negative test #4",
			args: args{
				w: httptest.NewRecorder(),
				s: store.GetTestStore(),
				user: models.User{
					ID: "wrong",
				},
				needHeader: true,
			},
			want: ResponseType{
				Code: http.StatusInternalServerError,
			},
			wantId: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodPost, "/testing", nil)

			if tt.args.needHeader {
				token, _ := auth.BuildJWTString(tt.args.user)

				r.Header.Set("Authorization", "Bearer "+token)
			}

			userID, resp := GetUserFromAuthHeader(tt.args.w, r, tt.args.s)

			if userID != tt.wantId {
				t.Errorf("GetUserFromAuthHeader() got userID = %v, want %v", userID, tt.wantId)
			}

			if resp.Code != tt.want.Code {
				t.Errorf("GetUserFromAuthHeader() got Code = %v, want %v", resp.Code, tt.want.Code)
			}
		})
	}
}

func TestLogin(t *testing.T) {
	type args struct {
		w    http.ResponseWriter
		s    store.Store
		data auth.AuthorizingData
	}
	tests := []struct {
		name string
		args args
		want ResponseType
	}{
		{
			name: "positive test #1",
			args: args{
				w: httptest.NewRecorder(),
				s: store.GetTestStore(),
				data: auth.AuthorizingData{
					UserName: "Exist User",
					Password: "Пароль",
				},
			},
			want: ResponseType{
				Code: http.StatusOK,
			},
		},
		{
			name: "negative test #2",
			args: args{
				w: httptest.NewRecorder(),
				s: store.GetTestStore(),
				data: auth.AuthorizingData{
					Password: "Пароль",
				},
			},
			want: ResponseType{
				Code: http.StatusBadRequest,
			},
		},
		{
			name: "negative test #3",
			args: args{
				w: httptest.NewRecorder(),
				s: store.GetTestStore(),
				data: auth.AuthorizingData{
					UserName: "Exist User",
				},
			},
			want: ResponseType{
				Code: http.StatusBadRequest,
			},
		},
		{
			name: "negative test #4",
			args: args{
				w: httptest.NewRecorder(),
				s: store.GetTestStore(),
				data: auth.AuthorizingData{
					UserName: "Not Exists User",
					Password: "234234",
				},
			},
			want: ResponseType{
				Code: http.StatusInternalServerError,
			},
		},
		{
			name: "negative test #5",
			args: args{
				w: httptest.NewRecorder(),
				s: store.GetTestStore(),
			},
			want: ResponseType{
				Code: http.StatusBadRequest,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.args.data)

			reader := io.NopCloser(bytes.NewReader(body))
			r := httptest.NewRequest(http.MethodPost, "/login", reader)

			got := Login(tt.args.w, r, tt.args.s)

			if got.Code != tt.want.Code {
				t.Errorf("Login() got code %v, want %v", got.Code, tt.want.Code)
			}
		})
	}
}

func TestRegisterUser(t *testing.T) {
	type args struct {
		w    http.ResponseWriter
		s    store.Store
		data auth.AuthorizingData
	}
	tests := []struct {
		name string
		args args
		want ResponseType
	}{
		{
			name: "positive test #1",
			args: args{
				w: httptest.NewRecorder(),
				s: store.GetTestStore(),
				data: auth.AuthorizingData{
					UserName: "New User",
					Password: "Пароль",
				},
			},
			want: ResponseType{
				Code: http.StatusOK,
			},
		},
		{
			name: "negative test #2",
			args: args{
				w: httptest.NewRecorder(),
				s: store.GetTestStore(),
				data: auth.AuthorizingData{
					UserName: "Exist User",
					Password: "Пароль",
				},
			},
			want: ResponseType{
				Code: http.StatusConflict,
			},
		},
		{
			name: "negative test #3",
			args: args{
				w: httptest.NewRecorder(),
				s: store.GetTestStore(),
				data: auth.AuthorizingData{
					Password: "Пароль",
				},
			},
			want: ResponseType{
				Code: http.StatusBadRequest,
			},
		},
		{
			name: "negative test #4",
			args: args{
				w: httptest.NewRecorder(),
				s: store.GetTestStore(),
				data: auth.AuthorizingData{
					UserName: "New User",
				},
			},
			want: ResponseType{
				Code: http.StatusBadRequest,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.args.data)

			reader := io.NopCloser(bytes.NewReader(body))
			r := httptest.NewRequest(http.MethodPost, "/register", reader)

			got := RegisterUser(tt.args.w, r, tt.args.s)

			if got.Code != tt.want.Code {
				t.Errorf("RegisterUser() got code %v, want %v", got.Code, tt.want.Code)
			}

			if got.Code != http.StatusOK {
				return
			}

			var bodyResp UserAuthBody

			if err := json.Unmarshal(got.Body, &bodyResp); err != nil {
				t.Errorf("RegisterUser() got wrong data")
			}

			if bodyResp.Key == "" {
				t.Errorf("RegisterUser() got key is empty")
			}

			if tt.args.w.Header().Get("Authorization") == "" {
				t.Errorf("RegisterUser() got authorization is empty")
			}
		})
	}
}

func Test_prepareUserAuthBody(t *testing.T) {
	type args struct {
		s  string
		k  string
		id string
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		{
			name: "positive test #1",
			args: args{
				s:  "Message",
				k:  "Key",
				id: "ID user",
			},
			want: []byte(`{"id":"ID user","msg":"Message","key":"Key"}`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := prepareUserAuthBody(tt.args.s, tt.args.k, tt.args.id); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("prepareUserAuthBody() = %v, want %v", got, tt.want)
			}
		})
	}
}
