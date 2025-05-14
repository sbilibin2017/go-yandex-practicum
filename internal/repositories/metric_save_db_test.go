package repositories

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/sbilibin2017/go-yandex-practicum/internal/types"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func setupMetricDBSaveTestDB(t *testing.T) (*sqlx.DB, func()) {
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "postgres:15",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_DB":       "testdb",
			"POSTGRES_USER":     "testuser",
			"POSTGRES_PASSWORD": "testpass",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp").WithStartupTimeout(30 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)

	host, err := container.Host(ctx)
	require.NoError(t, err)

	port, err := container.MappedPort(ctx, "5432")
	require.NoError(t, err)

	dsn := fmt.Sprintf("postgres://testuser:testpass@%s:%s/testdb?sslmode=disable", host, port.Port())

	var db *sqlx.DB
	for i := 0; i < 10; i++ {
		db, err = sqlx.Open("pgx", dsn)
		if err == nil && db.Ping() == nil {
			break
		}
		time.Sleep(time.Second)
	}
	require.NoError(t, err)
	require.NotNil(t, db)

	schema := `
	CREATE SCHEMA IF NOT EXISTS content;

	CREATE TABLE IF NOT EXISTS content.metrics (
		id TEXT NOT NULL,
		type TEXT NOT NULL,
		delta BIGINT,
		value DOUBLE PRECISION,
		PRIMARY KEY (id, type)
	);`
	_, err = db.Exec(schema)
	require.NoError(t, err)

	cleanup := func() {
		if db != nil {
			_ = db.Close()
		}
		if err := container.Terminate(ctx); err != nil {
			log.Printf("Failed to terminate container: %v", err)
		}
	}

	return db, cleanup
}

func TestMetricSaveDBRepository_Save(t *testing.T) {
	db, cleanup := setupMetricDBSaveTestDB(t)
	defer cleanup()

	ptrFloat64 := func(v float64) *float64 {
		return &v
	}

	repo := NewMetricSaveDBRepository(db, func(ctx context.Context) *sqlx.Tx {
		return nil
	})

	metric := types.Metrics{
		MetricID: types.MetricID{
			ID:   "test_metric",
			Type: types.GaugeMetricType,
		},
		Delta: nil,
		Value: ptrFloat64(42.5),
	}

	ctx := context.Background()
	err := repo.Save(ctx, metric)
	require.NoError(t, err)

	// Проверка: SELECT с использованием именованных параметров для sqlx
	var result struct {
		Value float64 `db:"value"`
	}
	query := `SELECT value FROM content.metrics WHERE id=:id AND type=:type`
	namedArgs := map[string]interface{}{
		"id":   metric.ID,
		"type": metric.Type,
	}
	stmt, err := db.PrepareNamed(query)
	require.NoError(t, err)
	err = stmt.GetContext(ctx, &result, namedArgs)
	require.NoError(t, err)
	assert.Equal(t, 42.5, result.Value)
}
