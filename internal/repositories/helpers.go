package repositories

import (
	"context"
	"os"

	"github.com/jmoiron/sqlx"
	"github.com/sbilibin2017/go-yandex-practicum/internal/middlewares"
)

// withFileSync выполняет синхронизацию файла до и после вызова функции fn.
// Сначала вызывается file.Sync() и file.Seek(0, 0), затем выполняется fn(file).
// После выполнения fn снова вызываются file.Sync() и file.Seek(0, 0).
// Возвращает ошибку, возникшую на любом из этапов синхронизации или выполнения fn.
func withFileSync(
	file *os.File,
	fn func(*os.File) error,
) error {
	if err := file.Sync(); err != nil {
		return err
	}
	if _, err := file.Seek(0, 0); err != nil {
		return err
	}
	if err := fn(file); err != nil {
		return err
	}
	if err := file.Sync(); err != nil {
		return err
	}
	if _, err := file.Seek(0, 0); err != nil {
		return err
	}
	return nil
}

// namedPreparer — интерфейс, определяющий возможность подготовки именованных запросов в контексте.
// Используется для унификации работы с транзакцией (sqlx.Tx) и базой данных (sqlx.DB).
type namedPreparer interface {
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
func getExecutor(
	ctx context.Context,
	db *sqlx.DB,
) namedPreparer {
	if tx := middlewares.GetTx(ctx); tx != nil {
		return tx
	}
	return db
}
