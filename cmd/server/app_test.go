package main

import (
	"fmt"
	"github.com/KirillKhitev/goph_keeper/internal/store"
	"net/http"
	"reflect"
	"testing"
)

func Test_app_Close(t *testing.T) {
	type fields struct {
		store  store.Store
		server http.Server
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "Test_app_Close",
			fields: fields{
				store: store.GetTestStore(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &app{
				store:  tt.fields.store,
				server: tt.fields.server,
			}
			if err := a.Close(); (err != nil) != tt.wantErr {
				t.Errorf("Close() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_app_getRouter(t *testing.T) {
	type fields struct {
		store  store.Store
		server http.Server
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "Test_app_getRouter",
			fields: fields{
				store: store.GetTestStore(),
			},
			want: "*chi.Mux",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &app{
				store:  tt.fields.store,
				server: tt.fields.server,
			}

			got := a.getRouter()

			if fmt.Sprintf("%T", got) != tt.want {
				t.Errorf("getRouter() = %T, want %s", got, tt.want)
			}
		})
	}
}

func Test_app_shutdownServer(t *testing.T) {
	type fields struct {
		store  store.Store
		server http.Server
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "Test_app_shutdownServer",
			fields: fields{
				store: store.GetTestStore(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &app{
				store:  tt.fields.store,
				server: tt.fields.server,
			}
			if err := a.shutdownServer(); (err != nil) != tt.wantErr {
				t.Errorf("shutdownServer() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_newApp(t *testing.T) {
	type args struct {
		s store.Store
	}
	tests := []struct {
		name string
		args args
		want *app
	}{
		{
			name: "Test_newApp",
			args: args{
				s: store.GetTestStore(),
			},
			want: &app{
				store: store.GetTestStore(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := newApp(tt.args.s); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newApp() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_printBuildInfo(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "Test_printBuildInfo",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			printBuildInfo()
		})
	}
}

func Test_run(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "Test_run",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			go run()
		})
	}
}
