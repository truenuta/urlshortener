package skill

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

type CompressWriter struct {
	writer     http.ResponseWriter
	gzipWriter *gzip.Writer
	useGzip    bool
}

func newCompressWriter(writer http.ResponseWriter) *CompressWriter {
	return &CompressWriter{
		writer:     writer,
		gzipWriter: gzip.NewWriter(writer),
	}
}

func (cw *CompressWriter) Header() http.Header {
	return cw.writer.Header()
}

func (cw *CompressWriter) Write(p []byte) (int, error) {
	if cw.useGzip {
		return cw.gzipWriter.Write(p)
	}
	return cw.writer.Write(p)
}

func (cw *CompressWriter) WriteHeader(statusCode int) {
	contentType := cw.writer.Header().Get("Content-Type")
	if statusCode < 300 && (strings.Contains(contentType, "application/json") || strings.Contains(contentType, "text/html")) {
		cw.writer.Header().Set("Content-Encoding", "gzip")
		cw.useGzip = true
	}
	cw.writer.WriteHeader(statusCode)
}

func (cw *CompressWriter) Close() error {
	if cw.useGzip {
		return cw.gzipWriter.Close()
	}
	return nil
}

type compressReader struct {
	readCloser io.ReadCloser
	zreader    *gzip.Reader
}

func newCompressReader(readCloser io.ReadCloser) (*compressReader, error) {
	zr, err := gzip.NewReader(readCloser)
	if err != nil {
		return nil, err
	}

	return &compressReader{
		readCloser: readCloser,
		zreader:    zr,
	}, nil
}

func (cr compressReader) Read(p []byte) (n int, err error) {
	return cr.zreader.Read(p)
}

func (c *compressReader) Close() error {
	if err := c.readCloser.Close(); err != nil {
		return err
	}
	return c.zreader.Close()
}

func GzipMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		originalWriter := w

		acceptEncoding := r.Header.Get("Accept-Encoding")
		supportsGzip := strings.Contains(acceptEncoding, "gzip")

		if supportsGzip {
			compressWriter := newCompressWriter(w)
			originalWriter = compressWriter
			defer compressWriter.Close()
		}

		contentEncoding := r.Header.Get("Content-Encoding")
		sendsGzip := strings.Contains(contentEncoding, "gzip")
		if sendsGzip {
			compressReader, err := newCompressReader(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			r.Body = compressReader
			defer compressReader.Close()
		}
		h.ServeHTTP(originalWriter, r)

	})
}
