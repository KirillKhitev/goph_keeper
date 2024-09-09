package api

import (
	"bytes"
	"encoding/json"
	"github.com/KirillKhitev/goph_keeper/internal/models"
	"github.com/KirillKhitev/goph_keeper/internal/store"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"
)

func TestGetData(t *testing.T) {
	type args struct {
		w      http.ResponseWriter
		userID string
		s      store.Store
		data   models.Data
	}
	tests := []struct {
		name string
		args args
		want ResponseType
	}{
		{
			name: "positive test #1",
			args: args{
				w:      httptest.NewRecorder(),
				userID: "1",
				s:      store.GetTestStore(),
				data: models.Data{
					ID:   "11122333",
					Name: []byte("Hello World"),
				},
			},
			want: ResponseType{
				Code: http.StatusOK,
				Body: []byte("{\n    \"id\": \"11122333\",\n    \"name\": \"SGVsbG8gV29ybGQ=\",\n    \"date\": \"0001-01-01T00:00:00Z\"\n}"),
			},
		},
		{
			name: "negative test #2",
			args: args{
				w:      httptest.NewRecorder(),
				userID: "1",
				s:      store.GetTestStore(),
				data:   models.Data{},
			},
			want: ResponseType{
				Code: http.StatusBadRequest,
				Body: []byte("Ошибка в запросе"),
			},
		},
		{
			name: "negative test #3",
			args: args{
				w:      httptest.NewRecorder(),
				userID: "1",
				s:      store.GetTestStore(),
				data: models.Data{
					ID:   "wrong",
					Name: []byte("Hello World"),
				},
			},
			want: ResponseType{
				Code: http.StatusInternalServerError,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body []byte

			if tt.args.data.ID != "" {
				body, _ = json.Marshal(tt.args.data)
			}

			reader := io.NopCloser(bytes.NewReader(body))

			r := httptest.NewRequest(http.MethodGet, "/users/1", reader)

			got := GetData(tt.args.w, r, tt.args.userID, tt.args.s)

			if !reflect.DeepEqual(got.Body, tt.want.Body) {
				t.Errorf("Body = %v, want %v", got.Body, tt.want.Body)
			}

			if !reflect.DeepEqual(got.Code, tt.want.Code) {
				t.Errorf("Code = %v, want %v", got.Code, tt.want.Code)
			}
		})
	}
}

func TestListData(t *testing.T) {
	type args struct {
		w      http.ResponseWriter
		userID string
		s      store.Store
	}
	tests := []struct {
		name string
		args args
		want ResponseType
	}{
		{
			name: "positive test #1",
			args: args{
				w:      httptest.NewRecorder(),
				userID: "1",
				s:      store.GetTestStore(),
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
			},
			want: ResponseType{
				Code: http.StatusInternalServerError,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodPost, "/list", nil)

			got := ListData(tt.args.w, r, tt.args.userID, tt.args.s)

			if got.Code != tt.want.Code {
				t.Errorf("Code = %v, want %v", got.Code, tt.want.Code)
			}
		})
	}
}

func TestUpdateData(t *testing.T) {
	type args struct {
		w      http.ResponseWriter
		userID string
		s      store.Store
		data   models.Data
	}
	tests := []struct {
		name string
		args args
		want ResponseType
	}{
		{
			name: "positive test #1",
			args: args{
				w:      httptest.NewRecorder(),
				userID: "1",
				s:      store.GetTestStore(),
				data: models.Data{
					ID:   "345",
					Name: []byte("Hello World"),
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
				data: models.Data{
					ID:   "345",
					Name: []byte("Hello World"),
				},
			},
			want: ResponseType{
				Code: http.StatusInternalServerError,
			},
		},
		{
			name: "negative test #3",
			args: args{
				w:      httptest.NewRecorder(),
				s:      store.GetTestStore(),
				userID: "123",
			},
			want: ResponseType{
				Code: http.StatusBadRequest,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body []byte

			if tt.args.data.ID != "" {
				body, _ = json.Marshal(tt.args.data)
			}

			reader := io.NopCloser(bytes.NewReader(body))
			r := httptest.NewRequest(http.MethodPost, "/update", reader)

			got := UpdateData(tt.args.w, r, tt.args.userID, tt.args.s)

			if got.Code != tt.want.Code {
				t.Errorf("Code = %v, want %v", got.Code, tt.want.Code)
			}
		})
	}
}

func Test_saveDataFile(t *testing.T) {
	type args struct {
		w    http.ResponseWriter
		data models.Data
		s    store.Store
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
				data: models.Data{
					ID:     "test_id",
					UserID: "111",
					Type:   "file",
					Body:   []byte("Hello World"),
				},
				s: store.GetTestStore(),
			},
			want: ResponseType{
				Code: http.StatusOK,
			},
		},
		{
			name: "negative test #2",
			args: args{
				w: httptest.NewRecorder(),
				data: models.Data{
					UserID: "111",
					Type:   "file",
					Body:   []byte("Hello World"),
				},
				s: store.GetTestStore(),
			},
			want: ResponseType{
				Code: http.StatusInternalServerError,
			},
		},
		{
			name: "negative test #3",
			args: args{
				w: httptest.NewRecorder(),
				data: models.Data{
					ID:   "test_id",
					Type: "file",
					Body: []byte("Hello World"),
				},
				s: store.GetTestStore(),
			},
			want: ResponseType{
				Code: http.StatusInternalServerError,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body []byte

			os.Mkdir(FilesDir, 777)

			if tt.args.data.ID != "" {
				body, _ = json.Marshal(tt.args.data)
			}

			reader := io.NopCloser(bytes.NewReader(body))
			r := httptest.NewRequest(http.MethodPost, "/update", reader)

			got := saveDataFile(tt.args.w, r, tt.args.data, tt.args.s)

			if got.Code == http.StatusOK {
				os.RemoveAll(FilesDir + string(os.PathSeparator) + tt.args.data.ID)
			}

			if got.Code != tt.want.Code {
				t.Errorf("Code = %v, want %v", got.Code, tt.want.Code)
			}
		})
	}
}
