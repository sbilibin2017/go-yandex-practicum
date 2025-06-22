package repositories

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	_ "github.com/jackc/pgx/v5/stdlib"
)

const migrationSQL = `
CREATE SCHEMA IF NOT EXISTS content;

CREATE TABLE IF NOT EXISTS content.metrics (
    id VARCHAR(255),
    type VARCHAR(255),
    delta BIGINT,
    value DOUBLE PRECISION,
    PRIMARY KEY (id, type)  
);
`

func setupPostgresContainer(ctx context.Context, t *testing.T) (*sqlx.DB, func()) {
	req := testcontainers.ContainerRequest{
		Image:        "postgres:15-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "testuser",
			"POSTGRES_PASSWORD": "testpass",
			"POSTGRES_DB":       "testdb",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp"),
	}

	postgresC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)

	host, err := postgresC.Host(ctx)
	require.NoError(t, err)
	port, err := postgresC.MappedPort(ctx, "5432")
	require.NoError(t, err)

	dsn := fmt.Sprintf("postgres://testuser:testpass@%s:%s/testdb?sslmode=disable", host, port.Port())

	db, err := sqlx.ConnectContext(ctx, "pgx", dsn)
	require.NoError(t, err)

	for _, q := range strings.Split(migrationSQL, ";") {
		q = strings.TrimSpace(q)
		if q == "" {
			continue
		}
		_, err := db.ExecContext(ctx, q)
		require.NoError(t, err)
	}

	cleanup := func() {
		db.Close()
		postgresC.Terminate(ctx)
	}

	return db, cleanup
}

func TestMetricDBSaveRepository_Save(t *testing.T) {
	ctx := context.Background()

	float64Ptr := func(f float64) *float64 { return &f }
	int64Ptr := func(i int64) *int64 { return &i }

	db, cleanup := setupPostgresContainer(ctx, t)
	defer cleanup()

	repo := NewMetricDBSaveRepository(
		WithMetricDBSaveRepositoryDB(db),
		WithMetricDBSaveRepositoryTxGetter(func(ctx context.Context) (*sqlx.Tx, bool) {
			return nil, false
		}),
	)

	metric := types.Metrics{
		ID:    "test-metric-id",
		Type:  "gauge",
		Delta: int64Ptr(100),
		Value: float64Ptr(123.45),
	}

	err := repo.Save(ctx, metric)
	require.NoError(t, err)

	var got types.Metrics
	err = db.GetContext(ctx, &got, `SELECT id, type, delta, value FROM content.metrics WHERE id = $1 AND type = $2`, metric.ID, metric.Type)
	require.NoError(t, err)
	require.Equal(t, metric.ID, got.ID)
	require.Equal(t, metric.Type, got.Type)
	require.Equal(t, metric.Delta, got.Delta)
	require.Equal(t, metric.Value, got.Value)

	metric.Delta = int64Ptr(200)
	metric.Value = float64Ptr(543.21)

	err = repo.Save(ctx, metric)
	require.NoError(t, err)

	err = db.GetContext(ctx, &got, `SELECT id, type, delta, value FROM content.metrics WHERE id = $1 AND type = $2`, metric.ID, metric.Type)
	require.NoError(t, err)
	require.Equal(t, metric.Delta, got.Delta)
	require.Equal(t, metric.Value, got.Value)
}

func TestMetricDBGetRepository_Get(t *testing.T) {
	ctx := context.Background()

	float64Ptr := func(f float64) *float64 { return &f }
	int64Ptr := func(i int64) *int64 { return &i }

	db, cleanup := setupPostgresContainer(ctx, t)
	defer cleanup()

	testMetric := types.Metrics{
		ID:    "get-metric-id",
		Type:  "counter",
		Delta: int64Ptr(42),
		Value: float64Ptr(99.9),
	}
	_, err := db.ExecContext(ctx,
		`INSERT INTO content.metrics (id, type, delta, value) VALUES ($1, $2, $3, $4)`,
		testMetric.ID, testMetric.Type, testMetric.Delta, testMetric.Value)
	require.NoError(t, err)

	repo := NewMetricDBGetRepository(
		WithMetricDBGetRepositoryDB(db),
		WithMetricDBGetRepositoryTxGetter(func(ctx context.Context) (*sqlx.Tx, bool) {
			return nil, false
		}),
	)

	got, err := repo.Get(ctx, types.MetricID{ID: testMetric.ID, Type: testMetric.Type})
	require.NoError(t, err)
	require.NotNil(t, got)
	require.Equal(t, testMetric.ID, got.ID)
	require.Equal(t, testMetric.Type, got.Type)
	require.Equal(t, testMetric.Delta, got.Delta)
	require.Equal(t, testMetric.Value, got.Value)

	got, err = repo.Get(ctx, types.MetricID{ID: "non-existent-id", Type: "gauge"})
	require.NoError(t, err)
	require.Nil(t, got)
}

func TestMetricDBListRepository_List(t *testing.T) {
	ctx := context.Background()

	float64Ptr := func(f float64) *float64 { return &f }
	int64Ptr := func(i int64) *int64 { return &i }

	db, cleanup := setupPostgresContainer(ctx, t)
	defer cleanup()

	metricsToInsert := []types.Metrics{
		{ID: "metric1", Type: "gauge", Delta: int64Ptr(10), Value: float64Ptr(1.1)},
		{ID: "metric2", Type: "counter", Delta: int64Ptr(20), Value: float64Ptr(2.2)},
		{ID: "metric3", Type: "gauge", Delta: int64Ptr(30), Value: float64Ptr(3.3)},
	}

	for _, m := range metricsToInsert {
		_, err := db.ExecContext(ctx,
			`INSERT INTO content.metrics (id, type, delta, value) VALUES ($1, $2, $3, $4)`,
			m.ID, m.Type, m.Delta, m.Value)
		require.NoError(t, err)
	}

	repo := NewMetricDBListRepository(
		WithMetricDBListRepositoryDB(db),
		WithMetricDBListRepositoryTxGetter(func(ctx context.Context) (*sqlx.Tx, bool) {
			return nil, false
		}),
	)

	got, err := repo.List(ctx)
	require.NoError(t, err)
	require.Len(t, got, len(metricsToInsert))

	// Verify contents match (order by ID)
	for i, m := range metricsToInsert {
		require.Equal(t, m.ID, got[i].ID)
		require.Equal(t, m.Type, got[i].Type)
		require.Equal(t, m.Delta, got[i].Delta)
		require.Equal(t, m.Value, got[i].Value)
	}
}
