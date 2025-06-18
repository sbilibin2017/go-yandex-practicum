package middlewares

import (
	"context"
	"fmt"
	"net/http"

	"github.com/jmoiron/sqlx"
	"github.com/sbilibin2017/go-yandex-practicum/internal/logger"
	"go.uber.org/zap"
)

func TxMiddleware(db *sqlx.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if db == nil {
				next.ServeHTTP(w, r)
				return
			}

			tx, err := db.BeginTxx(r.Context(), nil)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			ctx := setTx(r.Context(), tx)
			rr := &responseRecorder{ResponseWriter: w}

			err = withTx(ctx, func(ctx context.Context) error {
				next.ServeHTTP(rr, r.WithContext(ctx))
				return nil
			})
			if err != nil {
				logger.Log.Error(zap.Error(err))
			}
		})
	}
}

type responseRecorder struct {
	http.ResponseWriter
	wroteHeader bool
}

func (r *responseRecorder) WriteHeader(code int) {
	if !r.wroteHeader {
		r.wroteHeader = true
		r.ResponseWriter.WriteHeader(code)
	}
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	return r.ResponseWriter.Write(b)
}

// WithTx executes fn with the transaction stored in the context.
// It commits if fn returns nil, rolls back on error or panic.
func withTx(
	ctx context.Context,
	fn func(ctx context.Context) error,
) error {
	tx := getTx(ctx)
	if tx == nil {
		return fmt.Errorf("no transaction in context")
	}

	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
			panic(r) // rethrow after rollback
		}
	}()

	if err := fn(ctx); err != nil {
		_ = tx.Rollback()
		return err
	}

	if err := tx.Commit(); err != nil {
		_ = tx.Rollback()
		return err
	}

	return nil
}

type txContextKey struct{}

func setTx(ctx context.Context, tx *sqlx.Tx) context.Context {
	return context.WithValue(ctx, txContextKey{}, tx)
}

func getTx(ctx context.Context) *sqlx.Tx {
	tx, _ := ctx.Value(txContextKey{}).(*sqlx.Tx)
	return tx
}

// namedPreparer — интерфейс, определяющий возможность подготовки именованных запросов в контексте.
// Используется для унификации работы с транзакцией (sqlx.Tx) и базой данных (sqlx.DB).
type Executor interface {
	// PrepareNamedContext подготавливает именованный запрос в заданном контексте.
	PrepareNamedContext(ctx context.Context, query string) (*sqlx.NamedStmt, error)
}

// getExecutor возвращает объект для выполнения SQL-запросов с именованными параметрами.
// Если в контексте (через txGetter) присутствует транзакция, она будет использована для подготовки запроса;
// в противном случае возвращается экземпляр базы данных *sqlx.DB.
// Параметры:
//   - ctx: контекст выполнения запроса.
//   - db: база данных, используемая как запасной вариант.
//   - txGetter: функция для извлечения транзакции из контекста.
//
// Возвращает объект, реализующий интерфейс namedPreparer.
func GetExecutor(
	ctx context.Context,
	db *sqlx.DB,
) Executor {
	if tx := getTx(ctx); tx != nil {
		return tx
	}
	return db
}
