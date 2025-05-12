package runners_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/sbilibin2017/go-yandex-practicum/internal/runners"
	"github.com/stretchr/testify/assert"
)

func TestRunWorker_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockWorker := runners.NewMockWorker(ctrl)
	mockWorker.EXPECT().Start(gomock.Any()).Return(nil)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	err := runners.RunWorker(ctx, mockWorker)
	assert.NoError(t, err)
}

func TestRunWorker_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockWorker := runners.NewMockWorker(ctrl)
	expectedErr := errors.New("worker failed")
	mockWorker.EXPECT().Start(gomock.Any()).Return(expectedErr)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	err := runners.RunWorker(ctx, mockWorker)
	assert.EqualError(t, err, expectedErr.Error())
}

func TestRunWorker_ContextCanceled(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockWorker := runners.NewMockWorker(ctrl)
	workerStarted := make(chan struct{})
	mockWorker.EXPECT().Start(gomock.Any()).DoAndReturn(func(ctx context.Context) error {
		close(workerStarted)
		<-ctx.Done()
		return nil
	})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		_ = runners.RunWorker(ctx, mockWorker)
	}()
	<-workerStarted
	cancel()
	time.Sleep(100 * time.Millisecond)
}
