package apps_test

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/sbilibin2017/go-yandex-practicum/internal/apps"
	"github.com/stretchr/testify/assert"
)

func TestNewAgentAppConfig_Options(t *testing.T) {
	cfg := apps.NewAgentAppConfig(
		apps.WithAgentServerAddress("localhost:8080"),
		apps.WithAgentHeader("X-Test"),
		apps.WithAgentPollInterval(10),
		apps.WithAgentReportInterval(20),
		apps.WithAgentKey("secret"),
		apps.WithAgentRateLimit(100),
		apps.WithAgentCryptoKey("/path/to/key"),
		apps.WithAgentConfigPath("/path/to/config"),
		apps.WithAgentRestore(true),
		apps.WithAgentHashHeader("X-Hash"),
		apps.WithAgentLogLevel("debug"),
		apps.WithAgentBatchSize(50),
	)

	assert.Equal(t, "localhost:8080", cfg.ServerAddress)
	assert.Equal(t, "X-Test", cfg.Header)
	assert.Equal(t, 10, cfg.PollInterval)
	assert.Equal(t, 20, cfg.ReportInterval)
	assert.Equal(t, "secret", cfg.Key)
	assert.Equal(t, 100, cfg.RateLimit)
	assert.Equal(t, "/path/to/key", cfg.CryptoKey)
	assert.Equal(t, "/path/to/config", cfg.ConfigPath)
	assert.True(t, cfg.Restore)
	assert.Equal(t, "X-Hash", cfg.HashHeader)
	assert.Equal(t, "debug", cfg.LogLevel)
	assert.Equal(t, 50, cfg.BatchSize)
}

func TestAgentApp_Run_GracefulShutdown(t *testing.T) {
	workerCalled := false

	worker := func(ctx context.Context) error {
		workerCalled = true
		// Simulate work until context is done
		<-ctx.Done()
		return ctx.Err()
	}

	app := &apps.AgentApp{
		Workers: []func(ctx context.Context) error{worker},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err := app.Run(ctx)

	assert.True(t, workerCalled, "worker should have been called")
	assert.ErrorIs(t, err, context.DeadlineExceeded, "Run should return context error after timeout")
}

func TestAgentApp_Run_WorkerReturnsError(t *testing.T) {
	expectedErr := errors.New("worker error")

	worker := func(ctx context.Context) error {
		return expectedErr
	}

	app := &apps.AgentApp{
		Workers: []func(ctx context.Context) error{worker},
	}

	ctx := context.Background()
	err := app.Run(ctx)

	assert.Equal(t, expectedErr, err)
}

func TestAgentApp_Run_MultipleWorkers(t *testing.T) {
	worker1Called := false
	worker2Called := false

	worker1 := func(ctx context.Context) error {
		worker1Called = true
		<-ctx.Done()
		return ctx.Err()
	}
	worker2 := func(ctx context.Context) error {
		worker2Called = true
		<-ctx.Done()
		return ctx.Err()
	}

	app := &apps.AgentApp{
		Workers: []func(ctx context.Context) error{worker1, worker2},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err := app.Run(ctx)

	assert.True(t, worker1Called)
	assert.True(t, worker2Called)
	assert.ErrorIs(t, err, context.DeadlineExceeded)
}

func createTempPEMFile(t *testing.T) string {
	content := `-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAnuTkq5yfRQzW5ujQl6ML
ZJ7ZaJ3/gIj2Yu1MSZrD7u7WTfvg+YhTL0jL2Xptx7qv4ldREBAzT6pqftBjVhBx
bZFA8jK72E8Ck7kjxxnPzNffRM1NKoUDKweQpxPxHpkXPLtGQ24qGzA5cdyJcd0H
+np3ec7qxsk9ZxIvAtGLq3d/gQX+Q0LyewZTh4nhnm2u3t8CZIUk1QbBdX7x3Aqa
MAm1KxE7+UDBQ+cOYLZfxZ2mV3zJJycmrN6lzWcwxJxqULC7FHTK47ZJZDxO1f94
1YqNKX74T3r/jr9lG9CVp7ypY2NdlzLNpeOQ4mfyQ+cDlMKzLC7TvU9bXmA5Tqrv
hwIDAQAB
-----END PUBLIC KEY-----`

	f, err := os.CreateTemp("", "testkey-*.pem")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatal(err)
	}
	f.Close()
	return f.Name()
}

func TestNewAgentApp_Success(t *testing.T) {
	pemPath := createTempPEMFile(t)
	defer os.Remove(pemPath)

	app, err := apps.NewAgentApp(
		apps.WithAgentServerAddress("localhost:8080"),
		apps.WithAgentHeader("X-Test-Header"),
		apps.WithAgentKey("testkey"),
		apps.WithAgentPollInterval(5),
		apps.WithAgentReportInterval(10),
		apps.WithAgentBatchSize(100),
		apps.WithAgentRateLimit(10),
		apps.WithAgentCryptoKey(pemPath),
		apps.WithAgentLogLevel("debug"),
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if app == nil {
		t.Fatal("expected non-nil app")
	}
}
