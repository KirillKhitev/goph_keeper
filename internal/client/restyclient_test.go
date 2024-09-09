package client

import (
	"context"
	"encoding/base64"
	"github.com/go-resty/resty/v2"
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

func TestRestyClient_Send(t *testing.T) {
	type fields struct {
		client *resty.Client
		userID string
	}
	type args struct {
		ctx     context.Context
		url     string
		headers map[string]string
		data    []byte
		method  string
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
				userID: "111",
			},
			args: args{
				ctx:     context.Background(),
				url:     "http://localhost:8080/api/user/login",
				headers: map[string]string{},
				data:    []byte("{\"user_name\":\"n\",\"password\":\"123\"}"),
				method:  "POST",
			},
			want: APIServiceResult{
				Code: 200,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &RestyClient{
				client: tt.fields.client,
				userID: tt.fields.userID,
			}
			if got := c.Send(tt.args.ctx, tt.args.url, tt.args.headers, tt.args.data, tt.args.method); got.Code != tt.want.Code {
				t.Errorf("Send() = %v, want %v", got, tt.want)
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
