package middlewares

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

// GzipMiddleware is an HTTP middleware that enables gzip compression and decompression.
//
// If the incoming request has a `Content-Encoding: gzip` header, the middleware
// will automatically decompress the request body before passing it to the next handler.
//
// If the request contains an `Accept-Encoding` header with "gzip", the response
// will be compressed using gzip, and the `Content-Encoding: gzip` header will be added
// to the response.
//
// If decompression fails, the middleware responds with HTTP status 400 (Bad Request).
func GzipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Decompress the request body if it's gzipped
		if r.Header.Get("Content-Encoding") == "gzip" {
			gzReader, err := gzip.NewReader(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			defer gzReader.Close()
			r.Body = io.NopCloser(gzReader)
		}

		// Compress the response if the client supports gzip
		if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			gzw := gzip.NewWriter(w)
			defer gzw.Close()

			w.Header().Set("Content-Encoding", "gzip")
			gzwResponseWriter := &gzipResponseWriter{Writer: gzw, ResponseWriter: w}
			next.ServeHTTP(gzwResponseWriter, r)
		} else {
			next.ServeHTTP(w, r)
		}
	})
}

// gzipResponseWriter wraps an http.ResponseWriter and writes compressed data using gzip.Writer.
type gzipResponseWriter struct {
	io.Writer           // gzip.Writer
	http.ResponseWriter // original HTTP response writer
}

// Write writes the compressed response body using the underlying gzip.Writer.
func (w *gzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}
