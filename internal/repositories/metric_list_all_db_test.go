package repositories

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func setupMetricDBListAllTestDB(t *testing.T) (*sqlx.DB, func()) {
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

func TestMetricListAllDBRepository_ListAll(t *testing.T) {
	db, cleanup := setupMetricDBListAllTestDB(t)
	defer cleanup()

	ptrFloat64 := func(v float64) *float64 {
		return &v
	}

	repo := NewMetricListAllDBRepository(db, func(ctx context.Context) *sqlx.Tx {
		return nil
	})

	metrics := []types.Metrics{
		{
			MetricID: types.MetricID{
				ID:   "test_metric_1",
				Type: types.GaugeMetricType,
			},
			Delta: nil,
			Value: ptrFloat64(42.5),
		},
		{
			MetricID: types.MetricID{
				ID:   "test_metric_2",
				Type: types.GaugeMetricType,
			},
			Delta: nil,
			Value: ptrFloat64(100.5),
		},
	}

	// Вставка данных в базу с использованием именованных параметров
	for _, metric := range metrics {
		_, err := db.NamedExec(`INSERT INTO content.metrics (id, type, delta, value) VALUES (:id, :type, :delta, :value)`,
			map[string]interface{}{
				"id":    metric.ID,
				"type":  metric.Type,
				"delta": metric.Delta,
				"value": metric.Value,
			})
		require.NoError(t, err)
	}

	ctx := context.Background()
	metricList, err := repo.ListAll(ctx)
	require.NoError(t, err)

	// Проверка, что количество метрик соответствует
	assert.Len(t, metricList, len(metrics))

	// Проверка, что метрики отсортированы по ID
	assert.Equal(t, "test_metric_1", metricList[0].ID)
	assert.Equal(t, "test_metric_2", metricList[1].ID)
}
