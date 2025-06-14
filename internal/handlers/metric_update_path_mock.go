// Code generated by MockGen. DO NOT EDIT.
// Source: /home/sergey/Go/go-yandex-practicum/internal/handlers/metric_update_path.go

// Package handlers is a generated GoMock package.
package handlers

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	types "github.com/sbilibin2017/go-yandex-practicum/internal/types"
)

// MockMetricUpdatePathService is a mock of MetricUpdatePathService interface.
type MockMetricUpdatePathService struct {
	ctrl     *gomock.Controller
	recorder *MockMetricUpdatePathServiceMockRecorder
}

// MockMetricUpdatePathServiceMockRecorder is the mock recorder for MockMetricUpdatePathService.
type MockMetricUpdatePathServiceMockRecorder struct {
	mock *MockMetricUpdatePathService
}

// NewMockMetricUpdatePathService creates a new mock instance.
func NewMockMetricUpdatePathService(ctrl *gomock.Controller) *MockMetricUpdatePathService {
	mock := &MockMetricUpdatePathService{ctrl: ctrl}
	mock.recorder = &MockMetricUpdatePathServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockMetricUpdatePathService) EXPECT() *MockMetricUpdatePathServiceMockRecorder {
	return m.recorder
}

// Updates mocks base method.
func (m *MockMetricUpdatePathService) Updates(ctx context.Context, metrics []types.Metrics) ([]types.Metrics, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Updates", ctx, metrics)
	ret0, _ := ret[0].([]types.Metrics)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Updates indicates an expected call of Updates.
func (mr *MockMetricUpdatePathServiceMockRecorder) Updates(ctx, metrics interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Updates", reflect.TypeOf((*MockMetricUpdatePathService)(nil).Updates), ctx, metrics)
}
