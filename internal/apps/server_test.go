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
	cfg := NewServerAppConfig()
	assert.NotNil(t, cfg)

	// Test WithServerAddress
	cfg = NewServerAppConfig(WithServerAddress("localhost:8080"))
	assert.Equal(t, "localhost:8080", cfg.ServerAddress)

	// Test WithServerDatabaseDSN
	cfg = NewServerAppConfig(WithServerDatabaseDSN("postgres://user:pass@localhost/db"))
	assert.Equal(t, "postgres://user:pass@localhost/db", cfg.DatabaseDSN)

	// Test WithServerStoreInterval
	cfg = NewServerAppConfig(WithServerStoreInterval(15))
	assert.Equal(t, 15, cfg.StoreInterval)

	// Test WithServerFileStoragePath
	cfg = NewServerAppConfig(WithServerFileStoragePath("/tmp/storage"))
	assert.Equal(t, "/tmp/storage", cfg.FileStoragePath)

	// Test WithServerRestore
	cfg = NewServerAppConfig(WithServerRestore(true))
	assert.True(t, cfg.Restore)

	// Test WithServerKey
	cfg = NewServerAppConfig(WithServerKey("mykey"))
	assert.Equal(t, "mykey", cfg.Key)

	// Test WithServerCryptoKey
	cfg = NewServerAppConfig(WithServerCryptoKey("cryptokey"))
	assert.Equal(t, "cryptokey", cfg.CryptoKey)

	// Test WithServerConfigPath
	cfg = NewServerAppConfig(WithServerConfigPath("/etc/config.yaml"))
	assert.Equal(t, "/etc/config.yaml", cfg.ConfigPath)

	// Test WithServerTrustedSubnet
	cfg = NewServerAppConfig(WithServerTrustedSubnet("192.168.1.0/24"))
	assert.Equal(t, "192.168.1.0/24", cfg.TrustedSubnet)

	// Test WithServerHashHeader
	cfg = NewServerAppConfig(WithServerHashHeader("X-Hash"))
	assert.Equal(t, "X-Hash", cfg.HashHeader)

	// Test WithServerLogLevel
	cfg = NewServerAppConfig(WithServerLogLevel("debug"))
	assert.Equal(t, "debug", cfg.LogLevel)

	// Test WithServerMigrationsDir
	cfg = NewServerAppConfig(WithServerMigrationsDir("/migrations"))
	assert.Equal(t, "/migrations", cfg.MigrationsDir)
}

func TestNewServerApp(t *testing.T) {
	// Test with minimal options (no DB, no file storage)
	_, err := NewServerApp(
		WithServerAddress(":0"),
		WithServerRestore(false),
	)
	assert.NoError(t, err)

}

func TestNewServerApp_FileStoragePath(t *testing.T) {
	// Use a temporary directory and file path
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "metrics.json")

	app, err := NewServerApp(
		WithServerAddress(":0"),
		WithServerFileStoragePath(filePath),
		WithServerRestore(false),
	)

	// Assert no error
	assert.NoError(t, err)
	assert.NotNil(t, app)

}

func TestNewServerApp_WithDatabaseDSN_AndMigrations(t *testing.T) {
	ctx := context.Background()

	// Start PostgreSQL container
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
	assert.NoError(t, err)
	defer pgContainer.Terminate(ctx)

	// Get container's host and port
	host, err := pgContainer.Host(ctx)
	assert.NoError(t, err)

	port, err := pgContainer.MappedPort(ctx, "5432")
	assert.NoError(t, err)

	// Build DSN
	dsn := fmt.Sprintf("postgres://testuser:testpass@%s:%s/testdb?sslmode=disable", host, port.Port())

	// Create temp directory for migrations
	tmpDir := t.TempDir()

	// Create a simple migration file - 0001_create_table.sql
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
	assert.NoError(t, err)

	// Run NewServerApp with migrations dir set
	app, err := NewServerApp(
		WithServerAddress(":0"),
		WithServerDatabaseDSN(dsn),
		WithServerRestore(false),
		WithServerMigrationsDir(tmpDir),
	)
	assert.NoError(t, err)
	assert.NotNil(t, app)

	// Verify DB connected and migration applied by checking if table exists
	var tableName string
	err = app.DB.Get(&tableName, "SELECT to_regclass('public.test_table')")
	assert.NoError(t, err)
	assert.Equal(t, "test_table", tableName)
}

func TestServerApp_Run_WithMemoryStorage(t *testing.T) {
	// Create app with in-memory storage (no DB DSN, no FileStoragePath)
	app, err := NewServerApp(
		WithServerStoreInterval(1),
		WithServerRestore(false),
		WithServerAddress(":0"), // use ephemeral port
	)
	require.NoError(t, err)
	require.NotNil(t, app)

	// Make sure memory repositories are set
	require.NotNil(t, app.MetricMemorySaveRepository)
	require.NotNil(t, app.MetricMemoryGetRepository)
	require.NotNil(t, app.MetricMemoryListRepository)

	// Add worker for memory storage (simulate what NewServerApp should do)
	app.Workers = append(app.Workers, func(ctx context.Context) error {
		// Just wait until context is done
		<-ctx.Done()
		return nil
	})

	// Run the server in a separate goroutine, cancel immediately to trigger shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	doneCh := make(chan error)
	go func() {
		doneCh <- app.Run(ctx)
	}()

	// Cancel immediately, triggering shutdown
	cancel()

	select {
	case err := <-doneCh:
		require.NoError(t, err)
	case <-time.After(5 * time.Second):
		t.Fatal("Server did not shutdown in time")
	}
}
