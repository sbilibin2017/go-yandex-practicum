package db

import (
	"context"
	"fmt"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/sbilibin2017/go-yandex-practicum/internal/middlewares"
	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func setupMetricListAllPostgresContainer(ctx context.Context, t *testing.T) (testcontainers.Container, *sqlx.DB) {
	req := testcontainers.ContainerRequest{
		Image:        "postgres:15",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_PASSWORD": "secret",
			"POSTGRES_USER":     "user",
			"POSTGRES_DB":       "testdb",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp").
			WithStartupTimeout(60 * time.Second),
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

	dsn := fmt.Sprintf("host=%s port=%s user=user password=secret dbname=testdb sslmode=disable", host, port.Port())

	db, err := sqlx.ConnectContext(ctx, "pgx", dsn)
	require.NoError(t, err)

	schema := `
CREATE SCHEMA IF NOT EXISTS content;

CREATE TABLE IF NOT EXISTS content.metrics (
	id TEXT NOT NULL,
	type TEXT NOT NULL,
	delta BIGINT NULL,
	value DOUBLE PRECISION NULL,
	PRIMARY KEY (id, type)
);`

	_, err = db.ExecContext(ctx, schema)
	require.NoError(t, err)

	return container, db
}

func metricListAllDummyExecGetter(ctx context.Context, db *sqlx.DB) middlewares.Executor {
	return db
}

func TestMetricListAllDBRepository_ListAll(t *testing.T) {
	ctx := context.Background()
	container, dbConn := setupMetricListAllPostgresContainer(ctx, t)
	defer func() {
		_ = dbConn.Close()
		ctxTerm, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := container.Terminate(ctxTerm); err != nil {
			t.Logf("failed to terminate container: %v", err)
		}
	}()

	repo := NewMetricListAllDBRepository(dbConn, metricListAllDummyExecGetter)

	float64Ptr := func(f float64) *float64 { return &f }
	int64Ptr := func(i int64) *int64 { return &i }

	// Insert test data
	_, err := dbConn.ExecContext(ctx, `
INSERT INTO content.metrics (id, type, delta, value) VALUES
	('metric_gauge_1', 'gauge', NULL, 123.456),
	('metric_counter_1', 'counter', 789, NULL),
	('metric_gauge_2', 'gauge', NULL, 654.321);
`)
	require.NoError(t, err)

	metrics, err := repo.ListAll(ctx)
	require.NoError(t, err)
	require.Len(t, metrics, 3)

	// Verify returned metrics (ordered by id)
	expected := []types.Metrics{
		{
			ID:    "metric_counter_1",
			Type:  types.Counter,
			Delta: int64Ptr(789),
			Value: nil,
		},
		{
			ID:    "metric_gauge_1",
			Type:  types.Gauge,
			Delta: nil,
			Value: float64Ptr(123.456),
		},
		{
			ID:    "metric_gauge_2",
			Type:  types.Gauge,
			Delta: nil,
			Value: float64Ptr(654.321),
		},
	}

	for i, got := range metrics {
		want := expected[i]

		assert.Equal(t, want.ID, got.ID)
		assert.Equal(t, want.Type, got.Type)

		if want.Value == nil {
			assert.Nil(t, got.Value)
		} else {
			assert.NotNil(t, got.Value)
			assert.InDelta(t, *want.Value, *got.Value, 0.0001)
		}

		if want.Delta == nil {
			assert.Nil(t, got.Delta)
		} else {
			assert.NotNil(t, got.Delta)
			assert.Equal(t, *want.Delta, *got.Delta)
		}
	}
}
