package runners

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRunWorker_Success(t *testing.T) {
	ctx := context.Background()

	worker := func(ctx context.Context) error {
		// эмуляция короткой работы
		time.Sleep(10 * time.Millisecond)
		return nil
	}

	err := RunWorker(ctx, worker)
	assert.NoError(t, err, "worker should complete successfully")
}

func TestRunWorker_Error(t *testing.T) {
	ctx := context.Background()

	worker := func(ctx context.Context) error {
		return errors.New("worker failed")
	}

	err := RunWorker(ctx, worker)
	assert.EqualError(t, err, "worker failed")
}

func TestRunWorker_ContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	worker := func(ctx context.Context) error {
		// эмуляция долгой работы
		time.Sleep(100 * time.Millisecond)
		return nil
	}

	// отменяем контекст до завершения воркера
	go func() {
		time.Sleep(10 * time.Millisecond)
		cancel()
	}()

	err := RunWorker(ctx, worker)
	assert.Equal(t, context.Canceled, err)
}
