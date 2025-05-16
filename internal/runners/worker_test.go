package runners_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/sbilibin2017/go-yandex-practicum/internal/runners"
	"github.com/stretchr/testify/require"
)

func TestRunWorker_StartReturnsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockWorker := runners.NewMockWorker(ctrl)
	expectedErr := errors.New("start error")
	mockWorker.EXPECT().Start(gomock.Any()).Return(expectedErr)
	ctx := context.Background()
	err := runners.RunWorker(ctx, mockWorker)
	require.Equal(t, expectedErr, err)
}

func TestRunWorker_ContextCanceled(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockWorker := runners.NewMockWorker(ctrl)
	mockWorker.EXPECT().Start(gomock.Any()).DoAndReturn(func(ctx context.Context) error {
		<-ctx.Done()
		return ctx.Err()
	})
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(10 * time.Millisecond)
		cancel()
	}()
	err := runners.RunWorker(ctx, mockWorker)
	require.ErrorIs(t, err, context.Canceled)
}
