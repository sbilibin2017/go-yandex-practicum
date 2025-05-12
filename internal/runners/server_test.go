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

func TestRunServer_CleanExit(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockServer := runners.NewMockServer(ctrl)
	mockServer.EXPECT().ListenAndServe().Return(nil)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	err := runners.RunServer(ctx, mockServer)
	assert.NoError(t, err)
}

func TestRunServer_ListenAndServeError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockServer := runners.NewMockServer(ctrl)
	expectedErr := errors.New("server failed to start")
	mockServer.EXPECT().ListenAndServe().Return(expectedErr)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	err := runners.RunServer(ctx, mockServer)
	assert.EqualError(t, err, expectedErr.Error())
}

func TestRunServer_ContextCanceled_ShutdownCalled(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockServer := runners.NewMockServer(ctrl)
	listenCalled := make(chan struct{})
	mockServer.EXPECT().ListenAndServe().DoAndReturn(func() error {
		close(listenCalled)
		select {}
	}).AnyTimes()
	mockServer.EXPECT().Shutdown(gomock.Any()).Return(nil).Times(1)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		_ = runners.RunServer(ctx, mockServer)
	}()
	<-listenCalled
	cancel()
	time.Sleep(100 * time.Millisecond)
}
