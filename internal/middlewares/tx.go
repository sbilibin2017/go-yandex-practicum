package middlewares

import (
	"bytes"
	"context"
	"database/sql"
	"net/http"

	"github.com/jmoiron/sqlx"
)

// TxOption defines a functional option for configuring TxMiddleware.
type TxOption func(*txMiddleware)

type txMiddleware struct {
	db       *sqlx.DB
	txOpts   *sql.TxOptions
	txSetter func(ctx context.Context, tx *sqlx.Tx) context.Context
}

// WithDB sets the database connection to be used by the transaction middleware.
func WithDB(db *sqlx.DB) TxOption {
	return func(mw *txMiddleware) {
		mw.db = db
	}
}

// WithTxOptions sets transaction options such as isolation level or read-only flag.
func WithTxOptions(opts *sql.TxOptions) TxOption {
	return func(mw *txMiddleware) {
		mw.txOpts = opts
	}
}

// WithTxSetter sets a function to inject the started transaction into the request context.
func WithTxSetter(setter func(ctx context.Context, tx *sqlx.Tx) context.Context) TxOption {
	return func(mw *txMiddleware) {
		mw.txSetter = setter
	}
}

// TxMiddleware returns an HTTP middleware that starts a DB transaction before handling the request,
// commits if successful, and rolls back on error.
// If no DB is configured, it passes through without starting a transaction.
func TxMiddleware(opts ...TxOption) (func(http.Handler) http.Handler, error) {
	mw := &txMiddleware{}

	for _, opt := range opts {
		opt(mw)
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if mw.db == nil {
				next.ServeHTTP(w, r)
				return
			}

			tx, err := mw.db.BeginTxx(r.Context(), mw.txOpts)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			ctx := r.Context()
			if mw.txSetter != nil {
				ctx = mw.txSetter(ctx, tx)
				r = r.WithContext(ctx)
			}

			brw := newBufferedTxResponseWriter()
			next.ServeHTTP(brw, r)

			if err := tx.Commit(); err != nil {
				_ = tx.Rollback() // ignore rollback error
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			brw.flushTo(w)
		})
	}, nil
}

// bufferedTxResponseWriter buffers HTTP response headers, status code, and body.
type bufferedTxResponseWriter struct {
	headers     http.Header
	body        bytes.Buffer
	statusCode  int
	wroteHeader bool
}

func newBufferedTxResponseWriter() *bufferedTxResponseWriter {
	return &bufferedTxResponseWriter{
		headers:    make(http.Header),
		statusCode: http.StatusOK,
	}
}

func (b *bufferedTxResponseWriter) Header() http.Header {
	return b.headers
}

func (b *bufferedTxResponseWriter) WriteHeader(statusCode int) {
	if !b.wroteHeader {
		b.statusCode = statusCode
		b.wroteHeader = true
	}
}

func (b *bufferedTxResponseWriter) Write(data []byte) (int, error) {
	return b.body.Write(data)
}

func (b *bufferedTxResponseWriter) flushTo(w http.ResponseWriter) {
	for k, vv := range b.headers {
		for _, v := range vv {
			w.Header().Add(k, v)
		}
	}
	w.WriteHeader(b.statusCode)
	w.Write(b.body.Bytes())
}
