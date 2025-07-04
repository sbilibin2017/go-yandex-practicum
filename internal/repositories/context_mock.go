// Code generated by MockGen. DO NOT EDIT.
// Source: /home/sergey/Go/go-yandex-practicum/internal/repositories/context.go

// Package repositories is a generated GoMock package.
package repositories

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	types "github.com/sbilibin2017/go-yandex-practicum/internal/types"
)

// MockMetricSaver is a mock of MetricSaver interface.
type MockMetricSaver struct {
	ctrl     *gomock.Controller
	recorder *MockMetricSaverMockRecorder
}

// MockMetricSaverMockRecorder is the mock recorder for MockMetricSaver.
type MockMetricSaverMockRecorder struct {
	mock *MockMetricSaver
}

// NewMockMetricSaver creates a new mock instance.
func NewMockMetricSaver(ctrl *gomock.Controller) *MockMetricSaver {
	mock := &MockMetricSaver{ctrl: ctrl}
	mock.recorder = &MockMetricSaverMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockMetricSaver) EXPECT() *MockMetricSaverMockRecorder {
	return m.recorder
}

// Save mocks base method.
func (m *MockMetricSaver) Save(ctx context.Context, metric types.Metrics) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Save", ctx, metric)
	ret0, _ := ret[0].(error)
	return ret0
}

// Save indicates an expected call of Save.
func (mr *MockMetricSaverMockRecorder) Save(ctx, metric interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Save", reflect.TypeOf((*MockMetricSaver)(nil).Save), ctx, metric)
}

// MockMetricGetter is a mock of MetricGetter interface.
type MockMetricGetter struct {
	ctrl     *gomock.Controller
	recorder *MockMetricGetterMockRecorder
}

// MockMetricGetterMockRecorder is the mock recorder for MockMetricGetter.
type MockMetricGetterMockRecorder struct {
	mock *MockMetricGetter
}

// NewMockMetricGetter creates a new mock instance.
func NewMockMetricGetter(ctrl *gomock.Controller) *MockMetricGetter {
	mock := &MockMetricGetter{ctrl: ctrl}
	mock.recorder = &MockMetricGetterMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockMetricGetter) EXPECT() *MockMetricGetterMockRecorder {
	return m.recorder
}

// Get mocks base method.
func (m *MockMetricGetter) Get(ctx context.Context, id types.MetricID) (*types.Metrics, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Get", ctx, id)
	ret0, _ := ret[0].(*types.Metrics)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Get indicates an expected call of Get.
func (mr *MockMetricGetterMockRecorder) Get(ctx, id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockMetricGetter)(nil).Get), ctx, id)
}

// MockMetricLister is a mock of MetricLister interface.
type MockMetricLister struct {
	ctrl     *gomock.Controller
	recorder *MockMetricListerMockRecorder
}

// MockMetricListerMockRecorder is the mock recorder for MockMetricLister.
type MockMetricListerMockRecorder struct {
	mock *MockMetricLister
}

// NewMockMetricLister creates a new mock instance.
func NewMockMetricLister(ctrl *gomock.Controller) *MockMetricLister {
	mock := &MockMetricLister{ctrl: ctrl}
	mock.recorder = &MockMetricListerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockMetricLister) EXPECT() *MockMetricListerMockRecorder {
	return m.recorder
}

// List mocks base method.
func (m *MockMetricLister) List(ctx context.Context) ([]*types.Metrics, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "List", ctx)
	ret0, _ := ret[0].([]*types.Metrics)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// List indicates an expected call of List.
func (mr *MockMetricListerMockRecorder) List(ctx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "List", reflect.TypeOf((*MockMetricLister)(nil).List), ctx)
}
