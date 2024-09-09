package auth

import (
	"github.com/KirillKhitev/goph_keeper/internal/models"
	"strings"
	"testing"
)

func TestAuthorizingData_GenerateHashPassword(t *testing.T) {
	type fields struct {
		UserName string
		Password string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "positive test #1",
			fields: fields{
				UserName: "testuser",
				Password: "testpassword",
			},
			want: "f90cdca6663ad028d2a6a661154bf59e8322effb8c8d70324101b53719f1f9b3",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &AuthorizingData{
				UserName: tt.fields.UserName,
				Password: tt.fields.Password,
			}
			if got := d.GenerateHashPassword(); got != tt.want {
				t.Errorf("GenerateHashPassword() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAuthorizingData_NewUserFromData(t *testing.T) {
	type fields struct {
		UserName string
		Password string
	}
	tests := []struct {
		name   string
		fields fields
		want   models.User
	}{
		{
			name: "positive test #1",
			fields: fields{
				UserName: "testuser",
				Password: "testpassword",
			},
			want: models.User{
				UserName:     "testuser",
				HashPassword: "f90cdca6663ad028d2a6a661154bf59e8322effb8c8d70324101b53719f1f9b3",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &AuthorizingData{
				UserName: tt.fields.UserName,
				Password: tt.fields.Password,
			}

			got := d.NewUserFromData()

			if got.UserName != tt.want.UserName {
				t.Errorf("UserName() = %v, want %v", got.UserName, tt.want.UserName)
			}

			if got.HashPassword != tt.want.HashPassword {
				t.Errorf("HashPassword() = %v, want %v", got.HashPassword, tt.want.HashPassword)
			}
		})
	}
}

func TestBuildJWTString(t *testing.T) {
	type args struct {
		user models.User
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "positive test #1",
			args: args{
				user: models.User{
					ID:       "111",
					UserName: "testuser",
				},
			},
			wantErr: false,
		},
		{
			name: "positive test #2",
			args: args{
				user: models.User{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := BuildJWTString(tt.args.user)
			if (err != nil) != tt.wantErr {
				t.Errorf("BuildJWTString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if strings.Count(got, ".") != 2 {
				t.Error("BuildJWTString() got wrong JWT token")
				return
			}
		})
	}
}

func TestGetHash(t *testing.T) {
	type args struct {
		data string
		key  string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "positive test #1",
			args: args{
				data: "Hello World",
				key:  "testkey",
			},
			want: "d9778588e4c61ddb6985c4d830618b1875499fd6224179e4a9710f2e8e0292e8",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetHash(tt.args.data, tt.args.key); got != tt.want {
				t.Errorf("GetHash() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetUserIDFromAuthHeader(t *testing.T) {
	user := models.User{
		ID:       "111",
		UserName: "testuser",
	}

	token, _ := BuildJWTString(user)

	type args struct {
		header string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "positive test #1",
			args: args{
				header: token,
			},
			want:    "111",
			wantErr: false,
		},
		{
			name: "negative test #2",
			args: args{
				header: "",
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "negative test #3",
			args: args{
				header: "dyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3MjU1NzA4MDksIlVzZXJJRCI6IjExMSJ9.11QS9DOzX4vjLtPtNqsj0c7z6ZC1hnemCOCtK01DSH1",
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetUserIDFromAuthHeader(tt.args.header)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetUserIDFromAuthHeader() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetUserIDFromAuthHeader() got = %v, want %v", got, tt.want)
			}
		})
	}
}
