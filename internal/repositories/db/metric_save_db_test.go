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

func setupMetricSavePostgresContainer(ctx context.Context, t *testing.T) (testcontainers.Container, *sqlx.DB) {
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

// Dummy execGetter that just returns the DB itself as executor.
func metricSaveDBDummyExecGetter(ctx context.Context, db *sqlx.DB) middlewares.Executor {
	return db
}

func TestMetricSaveDBRepository_Save(t *testing.T) {
	ctx := context.Background()
	container, dbConn := setupMetricSavePostgresContainer(ctx, t)

	defer func() {
		// Close DB connection
		_ = dbConn.Close()

		// Use fresh context with timeout for termination
		ctxTerm, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := container.Terminate(ctxTerm); err != nil {
			t.Logf("failed to terminate container: %v", err)
		}
	}()

	repo := NewMetricSaveDBRepository(dbConn, metricSaveDBDummyExecGetter)

	float64Ptr := func(f float64) *float64 { return &f }
	int64Ptr := func(i int64) *int64 { return &i }

	tests := []struct {
		name   string
		metric types.Metrics
	}{
		{
			name: "insert new gauge metric",
			metric: types.Metrics{
				ID:    "metric_gauge_1",
				Type:  types.Gauge,
				Value: float64Ptr(123.456),
				Delta: nil,
			},
		},
		{
			name: "insert new counter metric",
			metric: types.Metrics{
				ID:    "metric_counter_1",
				Type:  types.Counter,
				Value: nil,
				Delta: int64Ptr(789),
			},
		},
		{
			name: "update existing metric",
			metric: types.Metrics{
				ID:    "metric_gauge_1",
				Type:  types.Gauge,
				Value: float64Ptr(654.321),
				Delta: nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.Save(ctx, tt.metric)
			require.NoError(t, err)

			// Verify that metric is stored correctly
			var got types.Metrics
			query := `SELECT id, type, delta, value FROM content.metrics WHERE id=$1 AND type=$2`
			err = dbConn.GetContext(ctx, &got, query, tt.metric.ID, tt.metric.Type)
			require.NoError(t, err)

			assert.Equal(t, tt.metric.ID, got.ID)
			assert.Equal(t, tt.metric.Type, got.Type)

			if tt.metric.Value == nil {
				assert.Nil(t, got.Value)
			} else {
				assert.NotNil(t, got.Value)
				assert.InDelta(t, *tt.metric.Value, *got.Value, 0.0001)
			}

			if tt.metric.Delta == nil {
				assert.Nil(t, got.Delta)
			} else {
				assert.NotNil(t, got.Delta)
				assert.Equal(t, *tt.metric.Delta, *got.Delta)
			}
		})
	}
}
