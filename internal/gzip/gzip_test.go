package gzip

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func Test_compressReader_Close(t *testing.T) {
	data, _ := base64.StdEncoding.DecodeString("H4sIAAAAAAAA/6pWKi1OLYrPS8xNVbJSylPSUSpILC4uzy9KUbJSMjQyVqoFBAAA//8yJCyTIgAAAA==")

	r := io.NopCloser(bytes.NewReader(data))
	zr, _ := gzip.NewReader(r)

	type fields struct {
		r  io.ReadCloser
		zr *gzip.Reader
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "positive test #1",
			fields: fields{
				r:  r,
				zr: zr,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &compressReader{
				r:  tt.fields.r,
				zr: tt.fields.zr,
			}
			if err := c.Close(); (err != nil) != tt.wantErr {
				t.Errorf("Close() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_compressReader_Read(t *testing.T) {
	data, _ := base64.StdEncoding.DecodeString("H4sIAAAAAAAA/6pWKi1OLYrPS8xNVbJSylPSUSpILC4uzy9KUbJSMjQyVqoFBAAA//8yJCyTIgAAAA==")

	r := io.NopCloser(bytes.NewReader(data))
	zr, _ := gzip.NewReader(r)

	type fields struct {
		r  io.ReadCloser
		zr *gzip.Reader
	}
	type args struct {
		p []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantN   int
		wantErr bool
	}{
		{
			name: "positive test #1",
			fields: fields{
				r:  r,
				zr: zr,
			},
			args: args{
				p: []byte("{\"user_name\":\"n\",\"password\":\"123\"}"),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := compressReader{
				r:  tt.fields.r,
				zr: tt.fields.zr,
			}
			gotN, err := c.Read(tt.args.p)
			if (err != nil && err != io.EOF) != tt.wantErr {
				t.Errorf("Read() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotN != len(tt.args.p) {
				t.Errorf("Read() gotN = %v, want %v", gotN, tt.wantN)
			}
		})
	}
}

func Test_compressWriter_Close(t *testing.T) {
	w := httptest.NewRecorder()

	type fields struct {
		w  http.ResponseWriter
		zw *gzip.Writer
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "positive test #1",
			fields: fields{
				w:  w,
				zw: gzip.NewWriter(w),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &compressWriter{
				w:  tt.fields.w,
				zw: tt.fields.zw,
			}
			if err := c.Close(); (err != nil) != tt.wantErr {
				t.Errorf("Close() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_compressWriter_Header(t *testing.T) {
	w := httptest.NewRecorder()

	type fields struct {
		w  http.ResponseWriter
		zw *gzip.Writer
	}
	tests := []struct {
		name   string
		fields fields
		want   http.Header
	}{
		{
			name: "positive test #1",
			fields: fields{
				w:  w,
				zw: gzip.NewWriter(w),
			},
			want: w.Header(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &compressWriter{
				w:  tt.fields.w,
				zw: tt.fields.zw,
			}
			if got := c.Header(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Header() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_compressWriter_Write(t *testing.T) {
	w := httptest.NewRecorder()

	type fields struct {
		w  http.ResponseWriter
		zw *gzip.Writer
	}
	type args struct {
		p []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    int
		wantErr bool
	}{
		{
			name: "positive test #1",
			fields: fields{
				w:  w,
				zw: gzip.NewWriter(w),
			},
			args: args{
				p: []byte("{\"user_name\":\"n\",\"password\":\"123\"}"),
			},
			wantErr: false,
			want:    34,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &compressWriter{
				w:  tt.fields.w,
				zw: tt.fields.zw,
			}
			got, err := c.Write(tt.args.p)
			if (err != nil) != tt.wantErr {
				t.Errorf("Write() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Write() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_compressWriter_WriteHeader(t *testing.T) {
	w := httptest.NewRecorder()

	type fields struct {
		w  http.ResponseWriter
		zw *gzip.Writer
	}
	type args struct {
		statusCode int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "positive test #1",
			fields: fields{
				w:  w,
				zw: gzip.NewWriter(w),
			},
			args: args{
				statusCode: 200,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &compressWriter{
				w:  tt.fields.w,
				zw: tt.fields.zw,
			}
			c.WriteHeader(tt.args.statusCode)
		})
	}
}

func Test_newCompressReader(t *testing.T) {
	data, _ := base64.StdEncoding.DecodeString("H4sIAAAAAAAA/6pWKi1OLYrPS8xNVbJSylPSUSpILC4uzy9KUbJSMjQyVqoFBAAA//8yJCyTIgAAAA==")
	r := io.NopCloser(bytes.NewReader(data))

	type args struct {
		r io.ReadCloser
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "positive test #1",
			args: args{
				r: r,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := newCompressReader(tt.args.r)
			if (err != nil) != tt.wantErr {
				t.Errorf("newCompressReader() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func Test_newCompressWriter(t *testing.T) {
	w := httptest.NewRecorder()

	type args struct {
		w http.ResponseWriter
	}
	tests := []struct {
		name string
		args args
		want *compressWriter
	}{
		{
			name: "positive test #1",
			args: args{
				w: w,
			},
			want: &compressWriter{
				w:  w,
				zw: gzip.NewWriter(w),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := newCompressWriter(tt.args.w); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newCompressWriter() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMiddleware(t *testing.T) {
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello World"))
	})

	handlerToTest := Middleware(nextHandler)

	data, _ := base64.StdEncoding.DecodeString("H4sIAAAAAAAA/6pWKi1OLYrPS8xNVbJSylPSUSpILC4uzy9KUbJSMjQyVqoFBAAA//8yJCyTIgAAAA==")

	r := io.NopCloser(bytes.NewReader(data))

	req := httptest.NewRequest("GET", "http://testing", r)
	req.Header.Set("Accept-Encoding", "gzip")
	req.Header.Set("Content-Encoding", "gzip")

	w := httptest.NewRecorder()

	t.Run("positive test #1", func(t *testing.T) {
		handlerToTest.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", w.Code, http.StatusOK)
		}
	})

	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest("GET", "http://testing", nil)
	req2.Header.Set("Accept-Encoding", "gzip")
	req2.Header.Set("Content-Encoding", "gzip")

	t.Run("negative test #2", func(t *testing.T) {
		handlerToTest.ServeHTTP(w2, req2)
		if w2.Code != http.StatusInternalServerError {
			t.Errorf("handler returned wrong status code: got %v want %v", w2.Code, http.StatusInternalServerError)
		}
	})
}
