package repositories

import (
	"context"
	"os"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/sbilibin2017/go-yandex-practicum/internal/logger"
	"go.uber.org/zap"
)

func cleanQuery(query string) string {
	query = strings.ReplaceAll(query, "\n", " ")
	query = strings.ReplaceAll(query, "\r", " ")
	query = strings.Join(strings.Fields(query), " ")
	return query
}

func logQuery(
	tx *sqlx.Tx,
	query string,
	args map[string]any,
) {
	logger.Log.Info(
		"Executing query",
		zap.String("query", cleanQuery(query)),
		zap.Any("args", args),
		zap.Bool("using_tx", tx != nil),
	)
}

func logQueryError(err error) {
	logger.Log.Error("query error", zap.Error(err))
}

func namedExecContext(
	ctx context.Context,
	db *sqlx.DB,
	txProvider func(ctx context.Context) *sqlx.Tx,
	query string,
	args map[string]any,
) error {
	tx := txProvider(ctx)
	logQuery(tx, query, args)

	var err error
	if tx != nil {
		_, err = tx.NamedExecContext(ctx, query, args)
	} else {
		_, err = db.NamedExecContext(ctx, query, args)
	}
	if err != nil {
		logQueryError(err)
		return err
	}
	return nil
}

func namedQueryContext[T any](
	ctx context.Context,
	db *sqlx.DB,
	txProvider func(ctx context.Context) *sqlx.Tx,
	query string,
	args map[string]any,
) ([]T, error) {
	tx := txProvider(ctx)
	logQuery(tx, query, args)

	var rows *sqlx.Rows
	var err error
	if tx != nil {
		rows, err = tx.NamedQuery(query, args)
	} else {
		rows, err = db.NamedQueryContext(ctx, query, args)
	}
	if err != nil {
		logQueryError(err)
		return nil, err
	}
	defer rows.Close()

	return scanRows[T](rows)
}

func namedQueryOneContext[T any](
	ctx context.Context,
	db *sqlx.DB,
	txProvider func(ctx context.Context) *sqlx.Tx,
	query string,
	args map[string]any,
) (*T, error) {
	tx := txProvider(ctx)
	logQuery(tx, query, args)

	var rows *sqlx.Rows
	var err error
	if tx != nil {
		rows, err = tx.NamedQuery(query, args)
	} else {
		rows, err = db.NamedQueryContext(ctx, query, args)
	}
	if err != nil {
		logQueryError(err)
		return nil, err
	}
	defer rows.Close()

	return scanRow[T](rows)
}

func scanRows[T any](rows *sqlx.Rows) ([]T, error) {
	var result []T
	for rows.Next() {
		var item T
		if err := rows.StructScan(&item); err != nil {
			logQueryError(err)
			return nil, err
		}
		result = append(result, item)
	}
	if err := rows.Err(); err != nil {
		logQueryError(err)
		return nil, err
	}
	if len(result) == 0 {
		return nil, nil
	}
	return result, nil
}

func scanRow[T any](rows *sqlx.Rows) (*T, error) {
	var result T
	if rows.Next() {
		if err := rows.StructScan(&result); err != nil {
			logQueryError(err)
			return nil, err
		}
	} else {
		return nil, nil
	}
	if err := rows.Err(); err != nil {
		logQueryError(err)
		return nil, err
	}
	return &result, nil
}

func withFileSync(file *os.File, fn func(*os.File) error) error {
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
