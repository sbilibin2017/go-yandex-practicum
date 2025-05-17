package repositories

import (
	"context"
	"os"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

type dummy struct {
	ID    int     `db:"id"`
	Value float64 `db:"value"`
}

func TestCleanQuery(t *testing.T) {
	query := "SELECT *\nFROM\r users\nWHERE id = :id"
	expected := "SELECT * FROM users WHERE id = :id"
	result := cleanQuery(query)
	require.Equal(t, expected, result)
}

func TestWithFileSync(t *testing.T) {
	f, err := os.CreateTemp("", "testfile")
	require.NoError(t, err)
	defer os.Remove(f.Name())
	defer f.Close()

	err = withFileSync(f, func(file *os.File) error {
		_, err := file.Write([]byte("test content"))
		return err
	})

	require.NoError(t, err)
}

func TestNamedExecContext_WithoutTx(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	ctx := context.Background()

	mock.ExpectExec("INSERT INTO test_table").
		WithArgs("value1").
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = namedExecContext(ctx, sqlxDB, func(ctx context.Context) *sqlx.Tx { return nil },
		"INSERT INTO test_table (column1) VALUES (:value1)",
		map[string]any{"value1": "value1"},
	)

	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestNamedQueryOneContext_NoRows(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	ctx := context.Background()

	mock.ExpectQuery("").
		WillReturnRows(sqlmock.NewRows([]string{"id", "value"}))

	result, err := namedQueryOneContext[dummy](ctx, sqlxDB, func(ctx context.Context) *sqlx.Tx { return nil },
		"",
		map[string]any{"id": 1},
	)

	require.NoError(t, err)
	require.Nil(t, result)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestScanRows_ErrorOnScan(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	sqlxDB := sqlx.NewDb(db, "sqlmock")

	mock.ExpectQuery("SELECT .*").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("invalid_int"))

	rows, err := sqlxDB.Queryx("SELECT * FROM test")
	require.NoError(t, err)
	defer rows.Close()

	_, err = scanRows[dummy](rows)
	require.Error(t, err)
}

func TestScanRow_ErrorOnScan(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	sqlxDB := sqlx.NewDb(db, "sqlmock")

	mock.ExpectQuery("SELECT .*").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("invalid"))

	rows, err := sqlxDB.Queryx("SELECT * FROM test")
	require.NoError(t, err)
	defer rows.Close()

	_, err = scanRow[dummy](rows)
	require.Error(t, err)
}
