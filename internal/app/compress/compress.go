package compress

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"

	"github.com/AlexTerra21/gophermart/internal/app/logger"
)

const compressContent = "application/json, text/html"

func IsCompress(content string) bool {
	return strings.Contains(compressContent, content)
}

// compressWriter реализует интерфейс http.ResponseWriter и позволяет прозрачно для сервера
// сжимать передаваемые данные и выставлять правильные HTTP-заголовки
type compressWriter struct {
	w  http.ResponseWriter
	zw *gzip.Writer
}

func newCompressWriter(w http.ResponseWriter) *compressWriter {
	return &compressWriter{
		w:  w,
		zw: gzip.NewWriter(w),
	}
}

func (c *compressWriter) Header() http.Header {
	return c.w.Header()
}

func (c *compressWriter) Write(p []byte) (int, error) {
	if IsCompress(c.Header().Get("Content-Type")) {
		return c.zw.Write(p)
	} else {
		return c.w.Write(p)
	}
}

func (c *compressWriter) WriteHeader(statusCode int) {
	if IsCompress(c.Header().Get("Content-Type")) {
		c.w.Header().Set("Content-Encoding", "gzip")
	}
	c.w.WriteHeader(statusCode)
}

// Close закрывает gzip.Writer и досылает все данные из буфера.
func (c *compressWriter) Close() error {
	if IsCompress(c.Header().Get("Content-Type")) {
		return c.zw.Close()
	} else {
		return nil
	}
}

type compressReader struct {
	r  io.ReadCloser
	zr *gzip.Reader
}

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

func (c *compressReader) Read(p []byte) (int, error) {
	return c.zr.Read(p)
}

func (c *compressReader) Close() error {
	if err := c.r.Close(); err != nil {
		return err
	}
	return c.zr.Close()
}

func WithCompress(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ow := w

		acceptEncoding := r.Header.Get("Accept-Encoding")
		supportGzip := strings.Contains(acceptEncoding, "gzip")
		if supportGzip {
			cw := newCompressWriter(w)
			ow = cw
			defer cw.Close()
		}

		contentEncoding := r.Header.Get("Content-Encoding")
		sendsGzip := strings.Contains(contentEncoding, "gzip")
		if sendsGzip {
			cr, err := newCompressReader(r.Body)
			if err != nil {
				logger.Debug("Error", logger.Field{Key: "Error read encoded body", Val: err})
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			r.Body = cr
			defer cr.Close()
		}

		logger.Debug("WithCompress",
			logger.Field{Key: "Content-Type", Val: r.Header.Get("Content-Type")},
			logger.Field{Key: "Accept-Encoding", Val: r.Header.Get("Accept-Encoding")},
			logger.Field{Key: "Content-Encoding", Val: r.Header.Get("Content-Encoding")},
		)

		h.ServeHTTP(ow, r)
	}
}
