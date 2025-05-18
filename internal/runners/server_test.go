package runners

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestRunServer_ListenAndServeReturnsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockServer := NewMockServer(ctrl)
	expectedErr := errors.New("listen error")
	mockServer.EXPECT().ListenAndServe().Return(expectedErr)
	ctx := context.Background()
	err := RunServer(ctx, mockServer)
	require.Equal(t, expectedErr, err)
}

func TestRunServer_ContextCancelled_ShutdownSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockServer := NewMockServer(ctrl)
	mockServer.EXPECT().ListenAndServe().DoAndReturn(func() error {
		time.Sleep(50 * time.Millisecond)
		return nil
	})
	mockServer.EXPECT().Shutdown(gomock.Any()).Return(nil)
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(10 * time.Millisecond)
		cancel()
	}()
	err := RunServer(ctx, mockServer)
	require.ErrorIs(t, err, context.Canceled)
}

func TestRunServer_ContextCancelled_ShutdownFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockServer := NewMockServer(ctrl)
	mockServer.EXPECT().ListenAndServe().DoAndReturn(func() error {
		time.Sleep(50 * time.Millisecond)
		return nil
	})
	shutdownErr := errors.New("shutdown failed")
	mockServer.EXPECT().Shutdown(gomock.Any()).Return(shutdownErr)
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(10 * time.Millisecond)
		cancel()
	}()
	err := RunServer(ctx, mockServer)
	require.Equal(t, shutdownErr, err)
}
