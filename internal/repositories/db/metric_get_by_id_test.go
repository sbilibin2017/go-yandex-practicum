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

func setupMetricGetByIDPostgresContainer(ctx context.Context, t *testing.T) (testcontainers.Container, *sqlx.DB) {
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

func metricGetByIDDummyExecGetter(ctx context.Context, db *sqlx.DB) middlewares.Executor {
	return db
}

func TestMetricGetByIDDBRepository_GetByID(t *testing.T) {
	ctx := context.Background()
	container, dbConn := setupMetricGetByIDPostgresContainer(ctx, t)

	defer func() {
		_ = dbConn.Close()
		ctxTerm, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := container.Terminate(ctxTerm); err != nil {
			t.Logf("failed to terminate container: %v", err)
		}
	}()

	repo := NewMetricGetByIDDBRepository(dbConn, metricGetByIDDummyExecGetter)

	float64Ptr := func(f float64) *float64 { return &f }
	int64Ptr := func(i int64) *int64 { return &i }

	// Insert some test data first for GetByID tests
	_, err := dbConn.ExecContext(ctx, `
INSERT INTO content.metrics (id, type, delta, value) VALUES
	('metric_gauge_1', 'gauge', NULL, 123.456),
	('metric_counter_1', 'counter', 789, NULL);
`)
	require.NoError(t, err)

	tests := []struct {
		name      string
		inputID   types.MetricID
		want      *types.Metrics
		expectErr bool
	}{
		{
			name: "get existing gauge metric",
			inputID: types.MetricID{
				ID:   "metric_gauge_1",
				Type: types.Gauge,
			},
			want: &types.Metrics{
				ID:    "metric_gauge_1",
				Type:  types.Gauge,
				Value: float64Ptr(123.456),
				Delta: nil,
			},
			expectErr: false,
		},
		{
			name: "get existing counter metric",
			inputID: types.MetricID{
				ID:   "metric_counter_1",
				Type: types.Counter,
			},
			want: &types.Metrics{
				ID:    "metric_counter_1",
				Type:  types.Counter,
				Value: nil,
				Delta: int64Ptr(789),
			},
			expectErr: false,
		},
		{
			name: "metric not found",
			inputID: types.MetricID{
				ID:   "nonexistent_metric",
				Type: types.Gauge,
			},
			want:      nil,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := repo.GetByID(ctx, tt.inputID)
			if tt.expectErr {
				require.Error(t, err)
				assert.Nil(t, got)
				return
			}
			require.NoError(t, err)
			assert.NotNil(t, got)

			assert.Equal(t, tt.want.ID, got.ID)
			assert.Equal(t, tt.want.Type, got.Type)

			if tt.want.Value == nil {
				assert.Nil(t, got.Value)
			} else {
				assert.NotNil(t, got.Value)
				assert.InDelta(t, *tt.want.Value, *got.Value, 0.0001)
			}

			if tt.want.Delta == nil {
				assert.Nil(t, got.Delta)
			} else {
				assert.NotNil(t, got.Delta)
				assert.Equal(t, *tt.want.Delta, *got.Delta)
			}
		})
	}
}
