package mycrypto

import (
	"encoding/base64"
	"os"
	"reflect"
	"testing"
)

func createTestFile(t *testing.T) *os.File {
	err := os.Mkdir("users", 777)
	if err != nil {
		t.Fatalf("ошибка создания папки users: %w", err)
	}

	path := "users" + string(os.PathSeparator) + "test_user.txt"
	file, err := os.Create(path)
	if err != nil {
		t.Fatalf("ошибка при создании файла %s: %w", path, err)
	}

	key := []uint8{168, 49, 93, 155, 71, 68, 144, 227, 208, 250, 186, 149, 211, 116, 214, 100}

	body := base64.StdEncoding.EncodeToString(key)

	file.Write([]byte(body))

	return file
}

func createWrongFile(t *testing.T) *os.File {
	os.Mkdir("users", 777)
	file, _ := os.Create("users/wrong_file.txt")

	key := []uint8{168, 49, 93, 155, 71, 68, 144, 227, 208, 250, 186, 149, 211, 116, 214, 100, 22}

	body := base64.StdEncoding.EncodeToString(key)

	file.Write([]byte(body))

	return file
}

func TestEncrypt(t *testing.T) {
	f := createTestFile(t)
	fw := createWrongFile(t)

	defer func() {
		f.Close()
		fw.Close()
		os.RemoveAll("users")
	}()

	type args struct {
		src     []byte
		keyFile string
	}
	tests := []struct {
		name    string
		args    args
		want    []uint8
		wantErr bool
	}{
		{
			name: "positive test #1",
			args: args{
				src:     []byte("hello world"),
				keyFile: "test_user.txt",
			},
			want:    []uint8{78, 119, 102, 237, 115, 57, 184, 105, 80, 218, 182, 69, 237, 98, 54, 48, 113, 109, 164, 139, 145, 211, 65, 129, 67, 201, 196},
			wantErr: false,
		},
		{
			name: "negative test #2",
			args: args{
				src:     []byte("hello world"),
				keyFile: "",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "negative test #3",
			args: args{
				src:     []byte("hello world"),
				keyFile: "not_exist_file.txt",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "negative test #4",
			args: args{
				src:     []byte("hello world"),
				keyFile: "wrong_file.txt",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Encrypt(tt.args.src, tt.args.keyFile)
			if (err != nil) != tt.wantErr {
				t.Errorf("Encrypt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.want != nil && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Decrypt() = %v, want %v", got, tt.want)
			}
		})
	}

}

func TestDecrypt(t *testing.T) {
	f := createTestFile(t)
	fw := createWrongFile(t)
	defer func() {
		f.Close()
		fw.Close()
		os.RemoveAll("users")
	}()

	type args struct {
		data    []uint8
		keyFile string
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "positive test #1",
			args: args{
				data:    []uint8{78, 119, 102, 237, 115, 57, 184, 105, 80, 218, 182, 69, 237, 98, 54, 48, 113, 109, 164, 139, 145, 211, 65, 129, 67, 201, 196},
				keyFile: "test_user.txt",
			},
			want:    []byte("hello world"),
			wantErr: false,
		},
		{
			name: "negative test #2",
			args: args{
				data:    []uint8{78, 119, 102, 237, 115, 57, 184, 105, 80, 218, 182, 69, 237, 98, 54, 48, 113, 109, 164, 139, 145, 211, 65, 129, 67, 201, 196},
				keyFile: "",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "negative test #3",
			args: args{
				data:    []byte{},
				keyFile: "not_exist_file.txt",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "negative test #4",
			args: args{
				data:    []byte{},
				keyFile: "wrong_file.txt",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Decrypt(tt.args.data, tt.args.keyFile)
			if (err != nil) != tt.wantErr {
				t.Errorf("Decrypt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.want != nil && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Decrypt() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGenerateRandom(t *testing.T) {
	type args struct {
		size int
	}
	tests := []struct {
		name    string
		args    args
		want    int
		wantErr bool
	}{
		{
			name: "positive test #1",
			args: args{
				size: 5,
			},
			want:    5,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GenerateRandom(tt.args.size)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateRandom() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != tt.want {
				t.Errorf("GenerateRandom() got = %v chars, want %v", got, tt.want)
			}
		})
	}
}
