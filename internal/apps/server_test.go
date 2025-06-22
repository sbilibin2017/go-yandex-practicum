package apps

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestNewServerAppConfig_Options(t *testing.T) {
	// Base config with no options set
	cfg := newServerAppConfig()
	assert.NotNil(t, cfg)

	cfg = newServerAppConfig(WithServerAddress("localhost:8080"))
	assert.Equal(t, "localhost:8080", cfg.ServerAddress)

	cfg = newServerAppConfig(WithServerDatabaseDSN("postgres://user:pass@localhost/db"))
	assert.Equal(t, "postgres://user:pass@localhost/db", cfg.DatabaseDSN)

	cfg = newServerAppConfig(WithServerStoreInterval(15))
	assert.Equal(t, 15, cfg.StoreInterval)

	cfg = newServerAppConfig(WithServerFileStoragePath("/tmp/storage"))
	assert.Equal(t, "/tmp/storage", cfg.FileStoragePath)

	cfg = newServerAppConfig(WithServerRestore(true))
	assert.True(t, cfg.Restore)

	cfg = newServerAppConfig(WithServerKey("mykey"))
	assert.Equal(t, "mykey", cfg.Key)

	cfg = newServerAppConfig(WithServerCryptoKey("cryptokey"))
	assert.Equal(t, "cryptokey", cfg.CryptoKey)

	cfg = newServerAppConfig(WithServerConfigPath("/etc/config.yaml"))
	assert.Equal(t, "/etc/config.yaml", cfg.ConfigPath)

	cfg = newServerAppConfig(WithServerTrustedSubnet("192.168.1.0/24"))
	assert.Equal(t, "192.168.1.0/24", cfg.TrustedSubnet)

	cfg = newServerAppConfig(WithServerHashHeader("X-Hash"))
	assert.Equal(t, "X-Hash", cfg.HashHeader)

	cfg = newServerAppConfig(WithServerLogLevel("debug"))
	assert.Equal(t, "debug", cfg.LogLevel)

	cfg = newServerAppConfig(WithServerMigrationsDir("/migrations"))
	assert.Equal(t, "/migrations", cfg.MigrationsDir)
}

func TestNewServerApp(t *testing.T) {
	app, err := NewServerApp(
		WithServerAddress(":0"),
		WithServerRestore(false),
	)
	require.NoError(t, err)
	require.NotNil(t, app)
}

func TestNewServerApp_FileStoragePath(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "metrics.json")

	app, err := NewServerApp(
		WithServerAddress(":0"),
		WithServerFileStoragePath(filePath),
		WithServerRestore(false),
	)
	require.NoError(t, err)
	require.NotNil(t, app)
}

func TestNewServerApp_WithDatabaseDSN_AndMigrations(t *testing.T) {
	ctx := context.Background()

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

	pgContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)
	defer func() {
		require.NoError(t, pgContainer.Terminate(ctx))
	}()

	host, err := pgContainer.Host(ctx)
	require.NoError(t, err)

	port, err := pgContainer.MappedPort(ctx, "5432")
	require.NoError(t, err)

	dsn := fmt.Sprintf("postgres://testuser:testpass@%s:%s/testdb?sslmode=disable", host, port.Port())

	tmpDir := t.TempDir()

	// Write a valid migration file for goose with Up and Down
	migrationFile := filepath.Join(tmpDir, "0001_create_table.sql")
	err = os.WriteFile(migrationFile, []byte(`
-- +goose Up
CREATE TABLE IF NOT EXISTS test_table (
	id SERIAL PRIMARY KEY,
	name TEXT NOT NULL
);

-- +goose Down
DROP TABLE IF EXISTS test_table;
`), 0644)
	require.NoError(t, err)

	app, err := NewServerApp(
		WithServerAddress(":0"),
		WithServerDatabaseDSN(dsn),
		WithServerRestore(false),
		WithServerMigrationsDir(tmpDir),
	)
	require.NoError(t, err)
	require.NotNil(t, app)
	require.NotNil(t, app.Container)
	require.NotNil(t, app.Container.DB)

	var tableName string
	err = app.Container.DB.Get(&tableName, "SELECT to_regclass('public.test_table')")
	require.NoError(t, err)
	assert.Equal(t, "test_table", tableName)
}

func TestServerApp_Run_WithMemoryStorage(t *testing.T) {
	app, err := NewServerApp(
		WithServerStoreInterval(1),
		WithServerRestore(false),
		WithServerAddress(":0"),
	)
	require.NoError(t, err)
	require.NotNil(t, app)
	require.NotNil(t, app.Container)

	require.NotNil(t, app.Container.MetricMemorySaveRepository)
	require.NotNil(t, app.Container.MetricMemoryGetRepository)
	require.NotNil(t, app.Container.MetricMemoryListRepository)

	// Append dummy worker that exits on context cancellation
	app.Container.Workers = append(app.Container.Workers, func(ctx context.Context) error {
		<-ctx.Done()
		return nil
	})

	// Run app.Run in a separate goroutine
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	errCh := make(chan error)
	go func() {
		errCh <- app.Run(ctx)
	}()

	select {
	case err := <-errCh:
		require.NoError(t, err)
	case <-time.After(5 * time.Second):
		t.Fatal("Timeout waiting for app.Run to finish")
	}
}

func TestNewServerGRPCApp_Success(t *testing.T) {
	app, err := NewServerGRPCApp(
		WithServerLogLevel("info"),
		WithServerAddress("127.0.0.1:0"),
	)
	require.NoError(t, err)
	require.NotNil(t, app)
	require.NotNil(t, app.Container)

	// Append dummy worker to avoid zero workers edge case
	app.Container.Workers = append(app.Container.Workers, func(ctx context.Context) error {
		<-ctx.Done()
		return nil
	})

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	errCh := make(chan error)
	go func() {
		errCh <- app.Run(ctx)
	}()

	select {
	case err := <-errCh:
		require.NoError(t, err)
	case <-time.After(5 * time.Second):
		t.Fatal("Timeout waiting for app.Run to finish")
	}
}
