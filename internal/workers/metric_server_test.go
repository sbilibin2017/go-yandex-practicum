package workers

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
	"github.com/stretchr/testify/require"
)

func Test_loadMetricsFromFile(t *testing.T) {
	tests := []struct {
		name          string
		mockListAll   func(m *MockMetricListAllFileRepository, ctx context.Context, metric types.Metrics, err error)
		mockSave      func(m *MockMetricSaveRepository, ctx context.Context, metric types.Metrics, err error)
		expectedError error
	}{
		{
			name: "success",
			mockListAll: func(m *MockMetricListAllFileRepository, ctx context.Context, metric types.Metrics, _ error) {
				m.EXPECT().ListAll(ctx).Return([]types.Metrics{metric}, nil)
			},
			mockSave: func(m *MockMetricSaveRepository, ctx context.Context, metric types.Metrics, _ error) {
				m.EXPECT().Save(ctx, metric).Return(nil)
			},
			expectedError: nil,
		},
		{
			name: "listall error",
			mockListAll: func(m *MockMetricListAllFileRepository, ctx context.Context, _ types.Metrics, err error) {
				m.EXPECT().ListAll(ctx).Return(nil, err)
			},
			mockSave:      func(m *MockMetricSaveRepository, ctx context.Context, metric types.Metrics, err error) {},
			expectedError: errors.New("listall error"),
		},
		{
			name: "save error",
			mockListAll: func(m *MockMetricListAllFileRepository, ctx context.Context, metric types.Metrics, _ error) {
				m.EXPECT().ListAll(ctx).Return([]types.Metrics{metric}, nil)
			},
			mockSave: func(m *MockMetricSaveRepository, ctx context.Context, metric types.Metrics, err error) {
				m.EXPECT().Save(ctx, metric).Return(err)
			},
			expectedError: errors.New("save error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			ctx := context.Background()

			var delta int64 = 123
			testMetric := types.Metrics{
				ID:    "metric1",
				Type:  types.Counter,
				Delta: &delta,
			}

			mockList := NewMockMetricListAllFileRepository(ctrl)
			mockSave := NewMockMetricSaveRepository(ctrl)

			tt.mockListAll(mockList, ctx, testMetric, tt.expectedError)
			tt.mockSave(mockSave, ctx, testMetric, tt.expectedError)

			err := loadMetricsFromFile(ctx, mockList, mockSave)
			if tt.expectedError != nil {
				require.Error(t, err)
				require.EqualError(t, err, tt.expectedError.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func Test_saveMetricsToFile(t *testing.T) {
	tests := []struct {
		name          string
		mockListAll   func(m *MockMetricListAllRepository, ctx context.Context, metric types.Metrics, err error)
		mockSave      func(m *MockMetricSaveFileRepository, ctx context.Context, metric types.Metrics, err error)
		expectedError error
	}{
		{
			name: "success",
			mockListAll: func(m *MockMetricListAllRepository, ctx context.Context, metric types.Metrics, _ error) {
				m.EXPECT().ListAll(ctx).Return([]types.Metrics{metric}, nil)
			},
			mockSave: func(m *MockMetricSaveFileRepository, ctx context.Context, metric types.Metrics, _ error) {
				m.EXPECT().Save(ctx, metric).Return(nil)
			},
			expectedError: nil,
		},
		{
			name: "listall error",
			mockListAll: func(m *MockMetricListAllRepository, ctx context.Context, _ types.Metrics, err error) {
				m.EXPECT().ListAll(ctx).Return(nil, err)
			},
			mockSave:      func(m *MockMetricSaveFileRepository, ctx context.Context, metric types.Metrics, err error) {},
			expectedError: errors.New("listall error"),
		},
		{
			name: "save error",
			mockListAll: func(m *MockMetricListAllRepository, ctx context.Context, metric types.Metrics, _ error) {
				m.EXPECT().ListAll(ctx).Return([]types.Metrics{metric}, nil)
			},
			mockSave: func(m *MockMetricSaveFileRepository, ctx context.Context, metric types.Metrics, err error) {
				m.EXPECT().Save(ctx, metric).Return(err)
			},
			expectedError: errors.New("save error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			ctx := context.Background()

			value := 3.14 // type float64 omitted here

			testMetric := types.Metrics{
				ID:    "metric2",
				Type:  types.Gauge,
				Value: &value,
			}

			mockList := NewMockMetricListAllRepository(ctrl)
			mockSave := NewMockMetricSaveFileRepository(ctrl)

			tt.mockListAll(mockList, ctx, testMetric, tt.expectedError)
			tt.mockSave(mockSave, ctx, testMetric, tt.expectedError)

			err := saveMetricsToFile(ctx, mockList, mockSave)
			if tt.expectedError != nil {
				require.Error(t, err)
				require.EqualError(t, err, tt.expectedError.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestStartMetricServerWorker(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	metric := types.Metrics{
		ID:   "metric1",
		Type: types.Counter,
		Delta: func() *int64 {
			v := int64(10)
			return &v
		}(),
	}

	tests := []struct {
		name          string
		restore       bool
		storeInterval int
		mockSetup     func(
			mListAll *MockMetricListAllRepository,
			mSave *MockMetricSaveRepository,
			fListAll *MockMetricListAllFileRepository,
			fSave *MockMetricSaveFileRepository,
		)
		expectErr bool
	}{
		{
			name:          "Restore success, store interval zero (wait shutdown save success)",
			restore:       true,
			storeInterval: 0,
			mockSetup: func(mListAll *MockMetricListAllRepository, mSave *MockMetricSaveRepository, fListAll *MockMetricListAllFileRepository, fSave *MockMetricSaveFileRepository) {
				fListAll.EXPECT().ListAll(gomock.Any()).Return([]types.Metrics{metric}, nil)
				mSave.EXPECT().Save(gomock.Any(), metric).Return(nil)

				mListAll.EXPECT().ListAll(gomock.Any()).Return([]types.Metrics{metric}, nil)
				fSave.EXPECT().Save(gomock.Any(), metric).Return(nil)
			},
			expectErr: false,
		},
		{
			name:          "Restore fails",
			restore:       true,
			storeInterval: 1,
			mockSetup: func(mListAll *MockMetricListAllRepository, mSave *MockMetricSaveRepository, fListAll *MockMetricListAllFileRepository, fSave *MockMetricSaveFileRepository) {
				fListAll.EXPECT().ListAll(gomock.Any()).Return(nil, errors.New("restore error"))
			},
			expectErr: true,
		},
		{
			name:          "Store interval zero - save on shutdown error",
			restore:       false,
			storeInterval: 0,
			mockSetup: func(mListAll *MockMetricListAllRepository, mSave *MockMetricSaveRepository, fListAll *MockMetricListAllFileRepository, fSave *MockMetricSaveFileRepository) {
				mListAll.EXPECT().ListAll(gomock.Any()).Return([]types.Metrics{metric}, nil)
				fSave.EXPECT().Save(gomock.Any(), metric).Return(errors.New("save error"))
			},
			expectErr: true,
		},
		{
			name:          "Store interval non-zero - periodic save success then shutdown",
			restore:       false,
			storeInterval: 1,
			mockSetup: func(mListAll *MockMetricListAllRepository, mSave *MockMetricSaveRepository, fListAll *MockMetricListAllFileRepository, fSave *MockMetricSaveFileRepository) {
				mListAll.EXPECT().ListAll(gomock.Any()).Return([]types.Metrics{metric}, nil).MinTimes(1)
				fSave.EXPECT().Save(gomock.Any(), metric).Return(nil).MinTimes(1)
			},
			expectErr: false,
		},
		{
			name:          "Store interval non-zero - save returns error",
			restore:       false,
			storeInterval: 1,
			mockSetup: func(mListAll *MockMetricListAllRepository, mSave *MockMetricSaveRepository, fListAll *MockMetricListAllFileRepository, fSave *MockMetricSaveFileRepository) {
				mListAll.EXPECT().ListAll(gomock.Any()).Return([]types.Metrics{metric}, nil).Times(1)
				fSave.EXPECT().Save(gomock.Any(), metric).Return(errors.New("save failure")).Times(1)
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			mockListAllMem := NewMockMetricListAllRepository(ctrl)
			mockSaveMem := NewMockMetricSaveRepository(ctrl)
			mockListAllFile := NewMockMetricListAllFileRepository(ctrl)
			mockSaveFile := NewMockMetricSaveFileRepository(ctrl)

			tt.mockSetup(mockListAllMem, mockSaveMem, mockListAllFile, mockSaveFile)

			ctx, cancel := context.WithCancel(context.Background())
			if tt.storeInterval > 0 {
				go func() {
					time.Sleep(time.Duration(tt.storeInterval) * 1500 * time.Millisecond)
					cancel()
				}()
			} else {
				go func() {
					time.Sleep(50 * time.Millisecond)
					cancel()
				}()
			}

			err := StartMetricServerWorker(ctx, mockListAllMem, mockSaveMem, mockListAllFile, mockSaveFile, tt.restore, tt.storeInterval)
			if tt.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
