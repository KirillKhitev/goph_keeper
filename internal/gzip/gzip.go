// Пакет отвечает за компрессию/декомпрессию запросов и ответов сервера
package gzip

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

// Middleware, если агент поддерживает сжатие gzip, распаковывает данные в запросе и сжимает ответ.
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ow := w

		acceptEncoding := r.Header.Get("Accept-Encoding")
		supportsGzip := strings.Contains(acceptEncoding, "gzip")

		if supportsGzip {
			cw := newCompressWriter(w)

			cw.Header().Set("Content-Encoding", "gzip")
			ow = cw

			defer cw.Close()
		}

		contentEncoding := r.Header.Get("Content-Encoding")
		sendsGzip := strings.Contains(contentEncoding, "gzip")

		if sendsGzip {
			cr, err := newCompressReader(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			r.Body = cr
			defer cr.Close()
		}

		next.ServeHTTP(ow, r)
	})
}

// Writer-обертка.
type compressWriter struct {
	w  http.ResponseWriter
	zw *gzip.Writer
}

// Конструктор Writer-обертки.
func newCompressWriter(w http.ResponseWriter) *compressWriter {
	return &compressWriter{
		w:  w,
		zw: gzip.NewWriter(w),
	}
}

// Header возвращает заголовоки исходного writer-а.
func (c *compressWriter) Header() http.Header {
	return c.w.Header()
}

// Write вызывает метод Write нового Writer-a.
func (c *compressWriter) Write(p []byte) (int, error) {
	return c.zw.Write(p)
}

// Записывает в заголовки код статуса.
func (c *compressWriter) WriteHeader(statusCode int) {
	c.w.WriteHeader(statusCode)
}

// Закрывает новый Writer.
func (c *compressWriter) Close() error {
	return c.zw.Close()
}

// Reader-обертка.
type compressReader struct {
	r  io.ReadCloser
	zr *gzip.Reader
}

// Конструктор Reader-обертки.
func newCompressReader(r io.ReadCloser) (*compressReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	return &compressReader{
		r:  r,
		zr: zr,
	}, nil
}

// Вызывает метод Read нового Reader-a.
func (c compressReader) Read(p []byte) (n int, err error) {
	return c.zr.Read(p)
}

// Закрывает новый Reader.
func (c *compressReader) Close() error {
	if err := c.r.Close(); err != nil {
		return err
	}
	return c.zr.Close()
}
