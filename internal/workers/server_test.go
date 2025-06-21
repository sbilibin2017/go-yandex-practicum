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

	// Helper to create mocks with error responses on saveMetrics
	makeMocks := func(
		listReturn []*types.Metrics,
		listErr error,
		saveErr error,
	) (Lister, Saver) {
		mockLister := NewMockLister(ctrl)
		mockSaver := NewMockSaver(ctrl)
		mockLister.EXPECT().List(gomock.Any()).Return(listReturn, listErr).AnyTimes()
		if listErr == nil {
			for _, m := range listReturn {
				mockSaver.EXPECT().Save(gomock.Any(), *m).Return(saveErr).AnyTimes()
				if saveErr != nil {
					break
				}
			}
		}
		return mockLister, mockSaver
	}

	tests := []struct {
		name          string
		restore       bool
		storeInterval int
		// Mocks setup: listerFile/saver, lister/saverFile
		listerFileListErr error
		listerFileSaveErr error
		listerListErr     error
		saverFileSaveErr  error
		wantErr           bool
	}{
		{
			name:              "restore success",
			restore:           true,
			storeInterval:     0,
			listerFileListErr: nil,
			listerFileSaveErr: nil,
			wantErr:           false,
		},
		{
			name:              "restore list error",
			restore:           true,
			storeInterval:     0,
			listerFileListErr: errors.New("restore list failure"),
			listerFileSaveErr: nil,
			wantErr:           true,
		},
		{
			name:              "restore save error",
			restore:           true,
			storeInterval:     0,
			listerFileListErr: nil,
			listerFileSaveErr: errors.New("restore save failure"),
			wantErr:           true,
		},
		{
			name:             "periodic save success",
			restore:          false,
			storeInterval:    1,
			listerListErr:    nil,
			saverFileSaveErr: nil,
			wantErr:          false,
		},
		{
			name:             "periodic save error",
			restore:          false,
			storeInterval:    1,
			listerListErr:    nil,
			saverFileSaveErr: errors.New("periodic save failure"),
			wantErr:          true,
		},
		{
			name:              "shutdown save error",
			restore:           false,
			storeInterval:     0,
			listerFileListErr: nil,
			listerFileSaveErr: errors.New("shutdown save failure"),
			wantErr:           true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Context with timeout to avoid blocking forever
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()

			mockListerFile, mockSaver := makeMocks(
				[]*types.Metrics{{ID: "metric1"}}, tt.listerFileListErr, tt.listerFileSaveErr,
			)
			mockLister, mockSaverFile := makeMocks(
				[]*types.Metrics{{ID: "metric1"}}, tt.listerListErr, tt.saverFileSaveErr,
			)

			errCh := make(chan error, 1)
			go func() {
				errCh <- startMetricServerWorker(ctx, mockLister, mockSaver, mockListerFile, mockSaverFile, tt.restore, tt.storeInterval)
			}()

			// Wait for the worker to finish or timeout
			select {
			case err := <-errCh:
				if tt.wantErr {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
			case <-time.After(4 * time.Second):
				t.Fatal("test timed out waiting for startMetricServerWorker")
			}
		})
	}
}

func TestStartMetricServerWorker_ShutdownSaveError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Mocks for lister and saverFile, to simulate error on shutdown save
	mockLister := NewMockLister(ctrl)
	mockSaverFile := NewMockSaver(ctrl)

	// Setup the lister to return some metrics (assuming saveMetrics reads from lister.List)
	mockLister.EXPECT().List(gomock.Any()).Return([]*types.Metrics{{ID: "shutdownMetric"}}, nil).AnyTimes()

	// Setup saverFile.Save to return error on save call during shutdown
	mockSaverFile.EXPECT().Save(gomock.Any(), gomock.Any()).Return(errors.New("shutdown save failure")).AnyTimes()

	// Create a cancelable context and cancel it immediately to trigger shutdown path
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // trigger context done immediately

	err := startMetricServerWorker(ctx, mockLister, nil, nil, mockSaverFile, false, 1) // storeInterval=1 to enter ticker branch but context done immediately

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "shutdown save failure")
}

func TestServerWorkerConfigOptions(t *testing.T) {
	// Mocks or dummy implementations for interfaces (replace with your mocks if needed)
	var dummyLister Lister = nil
	var dummySaver Saver = nil

	tests := []struct {
		name     string
		options  []ServerWorkerOption
		expected ServerWorkerConfig
	}{
		{
			name:    "default config",
			options: nil,
			expected: ServerWorkerConfig{
				restore:       false,
				storeInterval: 0,
				lister:        nil,
				saver:         nil,
				listerFile:    nil,
				saverFile:     nil,
			},
		},
		{
			name:    "set restore true",
			options: []ServerWorkerOption{WithRestore(true)},
			expected: ServerWorkerConfig{
				restore: true,
			},
		},
		{
			name:    "set storeInterval",
			options: []ServerWorkerOption{WithStoreInterval(10)},
			expected: ServerWorkerConfig{
				storeInterval: 10,
			},
		},
		{
			name:    "set lister",
			options: []ServerWorkerOption{WithLister(dummyLister)},
			expected: ServerWorkerConfig{
				lister: dummyLister,
			},
		},
		{
			name:    "set saver",
			options: []ServerWorkerOption{WithSaver(dummySaver)},
			expected: ServerWorkerConfig{
				saver: dummySaver,
			},
		},
		{
			name:    "set listerFile",
			options: []ServerWorkerOption{WithListerFile(dummyLister)},
			expected: ServerWorkerConfig{
				listerFile: dummyLister,
			},
		},
		{
			name:    "set saverFile",
			options: []ServerWorkerOption{WithSaverFile(dummySaver)},
			expected: ServerWorkerConfig{
				saverFile: dummySaver,
			},
		},
		{
			name: "set multiple options",
			options: []ServerWorkerOption{
				WithRestore(true),
				WithStoreInterval(5),
				WithLister(dummyLister),
				WithSaver(dummySaver),
				WithListerFile(dummyLister),
				WithSaverFile(dummySaver),
			},
			expected: ServerWorkerConfig{
				restore:       true,
				storeInterval: 5,
				lister:        dummyLister,
				saver:         dummySaver,
				listerFile:    dummyLister,
				saverFile:     dummySaver,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := NewServerWorkerConfig(tt.options...)

			// Check fields individually to avoid failing if nil interface values are tricky
			if cfg.restore != tt.expected.restore {
				t.Errorf("restore = %v, want %v", cfg.restore, tt.expected.restore)
			}
			if cfg.storeInterval != tt.expected.storeInterval {
				t.Errorf("storeInterval = %d, want %d", cfg.storeInterval, tt.expected.storeInterval)
			}
			if cfg.lister != tt.expected.lister {
				t.Errorf("lister = %v, want %v", cfg.lister, tt.expected.lister)
			}
			if cfg.saver != tt.expected.saver {
				t.Errorf("saver = %v, want %v", cfg.saver, tt.expected.saver)
			}
			if cfg.listerFile != tt.expected.listerFile {
				t.Errorf("listerFile = %v, want %v", cfg.listerFile, tt.expected.listerFile)
			}
			if cfg.saverFile != tt.expected.saverFile {
				t.Errorf("saverFile = %v, want %v", cfg.saverFile, tt.expected.saverFile)
			}
		})
	}
}

func TestNewServerWorker_CallsStartMetricServerWorker(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockListerFile := NewMockLister(ctrl)
	mockSaver := NewMockSaver(ctrl)

	// Allow multiple calls because startMetricServerWorker calls List and Save more than once
	mockListerFile.EXPECT().
		List(gomock.Any()).
		Return([]*types.Metrics{{ID: "metric1"}}, nil).
		AnyTimes()

	mockSaver.EXPECT().
		Save(gomock.Any(), types.Metrics{ID: "metric1"}).
		Return(nil).
		AnyTimes()

	worker := NewServerWorker(
		WithListerFile(mockListerFile),
		WithSaver(mockSaver),
		WithRestore(true),
		WithStoreInterval(0),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err := worker(ctx)
	assert.NoError(t, err)
}
