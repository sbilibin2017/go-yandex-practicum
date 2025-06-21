package middlewares

import (
	"bytes"
	"context"
	"database/sql"
	"net/http"

	"github.com/jmoiron/sqlx"
)

// TxMiddlewareConfig holds configuration options for the transaction middleware.
type TxMiddlewareConfig struct {
	DB        *sqlx.DB                                               // DB is the database connection used to start transactions.
	TxOptions *sql.TxOptions                                         // TxOptions specifies options for starting the transaction.
	TxSetter  func(ctx context.Context, tx *sqlx.Tx) context.Context // TxSetter is a function to set the started transaction in the request context.
}

// TxOption defines a functional option for configuring TxMiddlewareConfig.
type TxOption func(cfg *TxMiddlewareConfig)

// NewTxMiddlewareConfig creates a new TxMiddlewareConfig applying the provided options.
func NewTxMiddlewareConfig(opts ...TxOption) *TxMiddlewareConfig {
	cfg := &TxMiddlewareConfig{}
	for _, opt := range opts {
		opt(cfg)
	}
	return cfg
}

// WithDB sets the database connection to be used by the transaction middleware.
func WithDB(db *sqlx.DB) TxOption {
	return func(cfg *TxMiddlewareConfig) {
		cfg.DB = db
	}
}

// WithTxOptions sets transaction options such as isolation level or read-only flag.
func WithTxOptions(opts *sql.TxOptions) TxOption {
	return func(cfg *TxMiddlewareConfig) {
		cfg.TxOptions = opts
	}
}

// WithTxSetter sets a function to inject the started transaction into the request context.
func WithTxSetter(setter func(ctx context.Context, tx *sqlx.Tx) context.Context) TxOption {
	return func(cfg *TxMiddlewareConfig) {
		cfg.TxSetter = setter
	}
}

// TxMiddleware returns an HTTP middleware that starts a database transaction before
// handling the request, and commits the transaction if the handler completes successfully.
// In case of an error during transaction commit, it attempts to rollback and responds
// with the appropriate HTTP status code.
//
// If no database connection is configured, the middleware simply calls the next handler.
//
// The transaction is injected into the request context using the configured TxSetter function, if provided.
func TxMiddleware(opts ...TxOption) (func(http.Handler) http.Handler, error) {
	cfg := NewTxMiddlewareConfig(opts...)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if cfg.DB == nil {
				// No DB configured, just call next handler.
				next.ServeHTTP(w, r)
				return
			}

			// Begin transaction with given options.
			tx, err := cfg.DB.BeginTxx(r.Context(), cfg.TxOptions)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			ctx := r.Context()
			if cfg.TxSetter != nil {
				// Inject transaction into context if setter provided.
				ctx = cfg.TxSetter(ctx, tx)
				r = r.WithContext(ctx)
			}

			// Use buffered response writer to capture output before committing.
			brw := newBufferedTxResponseWriter()
			next.ServeHTTP(brw, r)

			// Attempt to commit the transaction.
			if err := tx.Commit(); err != nil {
				// Commit failed, attempt rollback.
				_ = tx.Rollback() // ignore rollback error
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			// Flush buffered response to the original ResponseWriter.
			brw.flushTo(w)
		})
	}, nil
}

// bufferedTxResponseWriter buffers HTTP response headers, status code, and body
// to delay writing them until the transaction is committed.
type bufferedTxResponseWriter struct {
	headers     http.Header
	body        bytes.Buffer
	statusCode  int
	wroteHeader bool
}

// newBufferedTxResponseWriter creates a new bufferedTxResponseWriter with default status 200 OK.
func newBufferedTxResponseWriter() *bufferedTxResponseWriter {
	return &bufferedTxResponseWriter{
		headers:    make(http.Header),
		statusCode: http.StatusOK,
	}
}

// Header returns the buffered HTTP headers.
func (b *bufferedTxResponseWriter) Header() http.Header {
	return b.headers
}

// WriteHeader buffers the HTTP status code for the response.
func (b *bufferedTxResponseWriter) WriteHeader(statusCode int) {
	if !b.wroteHeader {
		b.statusCode = statusCode
		b.wroteHeader = true
	}
}

// Write buffers the response body data.
func (b *bufferedTxResponseWriter) Write(data []byte) (int, error) {
	return b.body.Write(data)
}

// flushTo writes the buffered headers, status code, and body to the provided ResponseWriter.
func (b *bufferedTxResponseWriter) flushTo(w http.ResponseWriter) {
	for k, vv := range b.headers {
		for _, v := range vv {
			w.Header().Add(k, v)
		}
	}
	w.WriteHeader(b.statusCode)
	w.Write(b.body.Bytes())
}
