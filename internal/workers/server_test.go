package workers

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
)

func TestSaveMetrics(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	type args struct {
		listReturn []*types.Metrics
		listErr    error
		saveErr    error
	}

	tests := []struct {
		name    string
		args    args
		wantErr string
	}{
		{
			name: "success",
			args: args{
				listReturn: []*types.Metrics{{ID: "metric1"}, {ID: "metric2"}},
				listErr:    nil,
				saveErr:    nil,
			},
			wantErr: "",
		},
		{
			name: "list error",
			args: args{
				listReturn: nil,
				listErr:    errors.New("list error"),
				saveErr:    nil,
			},
			wantErr: "list error",
		},
		{
			name: "save error",
			args: args{
				listReturn: []*types.Metrics{{ID: "metric1"}},
				listErr:    nil,
				saveErr:    errors.New("save error"),
			},
			wantErr: "save error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLister := NewMockLister(ctrl)
			mockSaver := NewMockSaver(ctrl)

			ctx := context.Background()
			mockLister.EXPECT().List(ctx).Return(tt.args.listReturn, tt.args.listErr)
			if tt.args.listErr == nil {
				for _, metric := range tt.args.listReturn {
					mockSaver.EXPECT().Save(ctx, *metric).Return(tt.args.saveErr)
					if tt.args.saveErr != nil {
						break
					}
				}
			}

			err := saveMetrics(ctx, mockLister, mockSaver)
			if tt.wantErr == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tt.wantErr)
			}
		})
	}
}

func TestStartMetricServerWorker(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	type fields struct {
		listerMem     *MockLister
		saverMem      *MockSaver
		listerFile    *MockLister
		saverFile     *MockSaver
		restore       bool
		storeInterval int
		ctxTimeout    time.Duration
	}

	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "restore success",
			fields: fields{
				listerMem:     nil,
				saverMem:      NewMockSaver(ctrl),
				listerFile:    NewMockLister(ctrl),
				saverFile:     nil,
				restore:       true,
				storeInterval: 0,
				ctxTimeout:    100 * time.Millisecond,
			},
			wantErr: false,
		},
		{
			name: "periodic save",
			fields: fields{
				listerMem:     NewMockLister(ctrl),
				saverMem:      nil,
				listerFile:    nil,
				saverFile:     NewMockSaver(ctrl),
				restore:       false,
				storeInterval: 1,
				ctxTimeout:    350 * time.Millisecond,
			},
			wantErr: false,
		},
		{
			name: "full mocks and periodic save",
			fields: fields{
				listerMem:     NewMockLister(ctrl),
				saverMem:      NewMockSaver(ctrl),
				listerFile:    NewMockLister(ctrl),
				saverFile:     NewMockSaver(ctrl),
				restore:       false,
				storeInterval: 1,
				ctxTimeout:    3 * time.Second,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.fields.listerMem != nil {
				tt.fields.listerMem.EXPECT().List(gomock.Any()).Return([]*types.Metrics{{ID: "metric1"}}, nil).AnyTimes()
			}
			if tt.fields.saverMem != nil {
				tt.fields.saverMem.EXPECT().Save(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
			}
			if tt.fields.listerFile != nil {
				tt.fields.listerFile.EXPECT().List(gomock.Any()).Return([]*types.Metrics{{ID: "restore1"}}, nil).AnyTimes()
			}
			if tt.fields.saverFile != nil {
				tt.fields.saverFile.EXPECT().Save(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
			}

			ctx, cancel := context.WithTimeout(context.Background(), tt.fields.ctxTimeout)
			defer cancel()

			err := startMetricServerWorker(ctx,
				tt.fields.listerMem,
				tt.fields.saverMem,
				tt.fields.listerFile,
				tt.fields.saverFile,
				tt.fields.restore,
				tt.fields.storeInterval,
			)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestStartMetricServerWorker_Errors(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name              string
		restore           bool
		storeInterval     int
		listerFileListErr error
		listerFileSaveErr error
		listerListErr     error
		saverFileSaveErr  error
		wantErr           bool

		setup func() (Lister, Saver, Lister, Saver)
	}{
		{
			name:          "restore success",
			restore:       true,
			storeInterval: 0,
			wantErr:       false,
			setup: func() (Lister, Saver, Lister, Saver) {
				mockListerFile := NewMockLister(ctrl)
				mockSaverFile := NewMockSaver(ctrl)
				mockListerMem := NewMockLister(ctrl)
				mockSaverMem := NewMockSaver(ctrl)

				mockListerFile.EXPECT().List(gomock.Any()).Return([]*types.Metrics{{ID: "metric"}}, nil).AnyTimes()
				mockSaverMem.EXPECT().Save(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

				mockListerMem.EXPECT().List(gomock.Any()).Return([]*types.Metrics{{ID: "metric"}}, nil).AnyTimes()
				mockSaverMem.EXPECT().Save(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

				return mockListerMem, mockSaverMem, mockListerFile, mockSaverFile
			},
		},
		{
			name:          "restore list error",
			restore:       true,
			storeInterval: 0,
			wantErr:       true,
			setup: func() (Lister, Saver, Lister, Saver) {
				mockListerFile := NewMockLister(ctrl)
				mockSaverFile := NewMockSaver(ctrl)
				mockListerMem := NewMockLister(ctrl)
				mockSaverMem := NewMockSaver(ctrl)

				mockListerFile.EXPECT().List(gomock.Any()).Return(nil, errors.New("list error")).AnyTimes()
				// no saver calls expected

				return mockListerMem, mockSaverMem, mockListerFile, mockSaverFile
			},
		},
		{
			name:          "restore save error",
			restore:       true,
			storeInterval: 0,
			wantErr:       true,
			setup: func() (Lister, Saver, Lister, Saver) {
				mockListerFile := NewMockLister(ctrl)
				mockSaverFile := NewMockSaver(ctrl)
				mockListerMem := NewMockLister(ctrl)
				mockSaverMem := NewMockSaver(ctrl)

				mockListerFile.EXPECT().List(gomock.Any()).Return([]*types.Metrics{{ID: "metric"}}, nil).AnyTimes()
				mockSaverMem.EXPECT().Save(gomock.Any(), gomock.Any()).Return(errors.New("save error")).AnyTimes()

				return mockListerMem, mockSaverMem, mockListerFile, mockSaverFile
			},
		},
		{
			name:          "periodic save error on save",
			restore:       false,
			storeInterval: 1,
			wantErr:       true,
			setup: func() (Lister, Saver, Lister, Saver) {
				mockListerFile := NewMockLister(ctrl)
				mockSaverFile := NewMockSaver(ctrl)
				mockListerMem := NewMockLister(ctrl)
				mockSaverMem := NewMockSaver(ctrl)

				mockListerMem.EXPECT().List(gomock.Any()).Return([]*types.Metrics{{ID: "metric"}}, nil).AnyTimes()
				mockSaverFile.EXPECT().Save(gomock.Any(), gomock.Any()).Return(errors.New("save error")).AnyTimes()

				return mockListerMem, mockSaverMem, mockListerFile, mockSaverFile
			},
		},
		{
			name:          "periodic save error on list",
			restore:       false,
			storeInterval: 1,
			wantErr:       true,
			setup: func() (Lister, Saver, Lister, Saver) {
				mockListerFile := NewMockLister(ctrl)
				mockSaverFile := NewMockSaver(ctrl)
				mockListerMem := NewMockLister(ctrl)
				mockSaverMem := NewMockSaver(ctrl)

				mockListerMem.EXPECT().List(gomock.Any()).Return(nil, errors.New("list error")).AnyTimes()
				// no saver calls expected

				return mockListerMem, mockSaverMem, mockListerFile, mockSaverFile
			},
		},
		{
			name:          "shutdown save error",
			restore:       false,
			storeInterval: 0,
			wantErr:       true,
			setup: func() (Lister, Saver, Lister, Saver) {
				mockListerFile := NewMockLister(ctrl)
				mockSaverFile := NewMockSaver(ctrl)
				mockListerMem := NewMockLister(ctrl)
				mockSaverMem := NewMockSaver(ctrl)

				mockListerFile.EXPECT().List(gomock.Any()).Return([]*types.Metrics{{ID: "metric"}}, nil).AnyTimes()
				mockSaverMem.EXPECT().Save(gomock.Any(), gomock.Any()).Return(errors.New("shutdown save error")).AnyTimes()

				mockListerMem.EXPECT().List(gomock.Any()).Return([]*types.Metrics{{ID: "metric"}}, nil).AnyTimes()
				mockSaverMem.EXPECT().Save(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

				return mockListerMem, mockSaverMem, mockListerFile, mockSaverFile
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockListerMem, mockSaverMem, mockListerFile, mockSaverFile := tt.setup()

			ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
			defer cancel()

			err := startMetricServerWorker(ctx,
				mockListerMem,
				mockSaverMem,
				mockListerFile,
				mockSaverFile,
				tt.restore,
				tt.storeInterval,
			)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
