package facades

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"math/big"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"

	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
	pb "github.com/sbilibin2017/go-yandex-practicum/protos"
)

func TestNewMetricFacade_DefaultConfig(t *testing.T) {
	_, err := NewMetricHTTPFacade()
	require.NoError(t, err)
}

func TestMetricFacadeOptionFunctions(t *testing.T) {
	tests := []struct {
		name     string
		option   MetricHTTPFacadeOpt
		assertFn func(t *testing.T, cfg *MetricHTTPFacade)
	}{
		{
			name:   "WithMetricFacadeServerAddress",
			option: WithMetricFacadeServerAddress("localhost:8080"),
			assertFn: func(t *testing.T, cfg *MetricHTTPFacade) {
				assert.Equal(t, "localhost:8080", cfg.serverAddress)
			},
		},
		{
			name:   "WithMetricFacadeHeader",
			option: WithMetricFacadeHeader("X-Custom-Header"),
			assertFn: func(t *testing.T, cfg *MetricHTTPFacade) {
				assert.Equal(t, "X-Custom-Header", cfg.header)
			},
		},
		{
			name:   "WithMetricFacadeKey",
			option: WithMetricFacadeKey("mysecretkey"),
			assertFn: func(t *testing.T, cfg *MetricHTTPFacade) {
				assert.Equal(t, "mysecretkey", cfg.key)
			},
		},
		{
			name:   "WithMetricFacadeCryptoKeyPath",
			option: WithMetricFacadeCryptoKeyPath("/path/to/key.pem"),
			assertFn: func(t *testing.T, cfg *MetricHTTPFacade) {
				assert.Equal(t, "/path/to/key.pem", cfg.cryptoKeyPath)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &MetricHTTPFacade{}
			tt.option(cfg)
			tt.assertFn(t, cfg)
		})
	}
}

func TestCalcBodyHashSum(t *testing.T) {
	data := []byte(`{"foo":"bar"}`)
	key := "secret"
	sum := calcBodyHashSum(data, key)
	assert.NotEmpty(t, sum)
	assert.Equal(t, 64, len(sum), "expected SHA256 hex string to be 64 characters")
}

func mustCreateTempFile(t *testing.T, content []byte) string {
	tmpFile, err := os.CreateTemp("", "testkey_*.pem")
	require.NoError(t, err)

	_, err = tmpFile.Write(content)
	require.NoError(t, err)
	require.NoError(t, tmpFile.Close())

	t.Cleanup(func() { os.Remove(tmpFile.Name()) })

	return tmpFile.Name()
}

func TestLoadPublicKey(t *testing.T) {
	t.Run("invalid and valid cases", func(t *testing.T) {
		tests := []struct {
			name        string
			setupFile   func(t *testing.T) string
			expectError bool
		}{
			{
				name:        "file not found",
				setupFile:   func(t *testing.T) string { return "nonexistent_file.pem" },
				expectError: true,
			},
			{
				name: "invalid PEM block",
				setupFile: func(t *testing.T) string {
					return mustCreateTempFile(t, []byte("not a pem"))
				},
				expectError: true,
			},
			{
				name: "wrong PEM block type",
				setupFile: func(t *testing.T) string {
					block := &pem.Block{Type: "WRONG TYPE", Bytes: []byte("some bytes")}
					return mustCreateTempFile(t, pem.EncodeToMemory(block))
				},
				expectError: true,
			},
			{
				name: "invalid key bytes in PEM",
				setupFile: func(t *testing.T) string {
					block := &pem.Block{Type: "PUBLIC KEY", Bytes: []byte("invalid key bytes")}
					return mustCreateTempFile(t, pem.EncodeToMemory(block))
				},
				expectError: true,
			},
			{
				name: "valid RSA public key",
				setupFile: func(t *testing.T) string {
					privKey, err := rsa.GenerateKey(rand.Reader, 2048)
					require.NoError(t, err)
					pubASN1, err := x509.MarshalPKIXPublicKey(&privKey.PublicKey)
					require.NoError(t, err)
					block := &pem.Block{Type: "PUBLIC KEY", Bytes: pubASN1}
					return mustCreateTempFile(t, pem.EncodeToMemory(block))
				},
				expectError: false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				path := tt.setupFile(t)
				pubKey, err := loadPublicKey(path)
				if tt.expectError {
					assert.Error(t, err)
					assert.Nil(t, pubKey)
				} else {
					assert.NoError(t, err)
					assert.NotNil(t, pubKey)
				}
			})
		}
	})
}

func TestEncryptBody(t *testing.T) {
	data := []byte("test data")

	t.Run("nil public key returns original data", func(t *testing.T) {
		encrypted, err := encryptBody(data, nil)
		require.NoError(t, err)
		assert.Equal(t, data, encrypted)
	})

	t.Run("with valid public key encrypts successfully", func(t *testing.T) {
		privKey, err := rsa.GenerateKey(rand.Reader, 2048)
		require.NoError(t, err)
		encrypted, err := encryptBody(data, &privKey.PublicKey)
		require.NoError(t, err)
		assert.NotEqual(t, data, encrypted)
		assert.Greater(t, len(encrypted), 0)
	})
}

func createTempPublicKeyFile(t *testing.T) string {
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	pubASN1, err := x509.MarshalPKIXPublicKey(&privKey.PublicKey)
	require.NoError(t, err)

	pubPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubASN1})
	tmpFile, err := os.CreateTemp("", "test_pubkey_*.pem")
	require.NoError(t, err)
	_, err = tmpFile.Write(pubPEM)
	require.NoError(t, err)
	require.NoError(t, tmpFile.Close())

	t.Cleanup(func() { os.Remove(tmpFile.Name()) })

	return tmpFile.Name()
}

func TestNewMetricFacade_WithValidCryptoKeyPath(t *testing.T) {
	path := createTempPublicKeyFile(t)
	mf, err := NewMetricHTTPFacade(WithMetricFacadeCryptoKeyPath(path))
	assert.NoError(t, err)
	assert.NotNil(t, mf)
	assert.NotNil(t, mf.publicKey)
}

func TestNewMetricFacade_WithInvalidCryptoKeyPath(t *testing.T) {
	mf, err := NewMetricHTTPFacade(WithMetricFacadeCryptoKeyPath("/non/existing/file.pem"))
	assert.Error(t, err)
	assert.Nil(t, mf)
}

func TestNewMetricFacade_WithInvalidPEMFile(t *testing.T) {
	tmpFile := mustCreateTempFile(t, []byte("not a valid pem"))
	mf, err := NewMetricHTTPFacade(WithMetricFacadeCryptoKeyPath(tmpFile))
	assert.Error(t, err)
	assert.Nil(t, mf)
}

func TestMetricFacade_Updates_WithRealHTTPServer(t *testing.T) {
	ctx := context.Background()
	v := float64(42)
	sampleMetrics := []*types.Metrics{
		{ID: "metric1", Type: "gauge", Value: &v},
	}

	tests := []struct {
		name           string
		metrics        []*types.Metrics
		key            string
		header         string
		publicKey      *rsa.PublicKey
		serverHandler  http.HandlerFunc
		wantErr        bool
		expectedStatus int
	}{
		{
			name:    "empty metrics slice should skip sending",
			metrics: []*types.Metrics{},
			wantErr: false,
		},
		{
			name:    "successful update without encryption and hash",
			metrics: sampleMetrics,
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
			wantErr:        false,
			expectedStatus: http.StatusOK,
		},
		{
			name:    "server returns error status",
			metrics: sampleMetrics,
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, "internal server error", http.StatusInternalServerError)
			},
			wantErr:        true,
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:    "successful update with HMAC header",
			metrics: sampleMetrics,
			key:     "test-secret",
			header:  "X-Signature",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				if r.Header.Get("X-Signature") == "" {
					http.Error(w, "missing signature header", http.StatusBadRequest)
					return
				}
				w.WriteHeader(http.StatusOK)
			},
			wantErr: false,
		},
		{
			name:    "encryption error triggers failure",
			metrics: sampleMetrics,
			publicKey: &rsa.PublicKey{
				N: new(big.Int), // non-nil but invalid
				E: 65537,
			},
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(tt.serverHandler)
			defer server.Close()

			mf, err := NewMetricHTTPFacade(
				WithMetricFacadeServerAddress(server.URL),
				WithMetricFacadeKey(tt.key),
				WithMetricFacadeHeader(tt.header),
				func(f *MetricHTTPFacade) { f.cryptoKeyPath = "" },
			)
			require.NoError(t, err)

			mf.publicKey = tt.publicKey

			err = mf.Updates(ctx, tt.metrics)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// MockMetricUpdaterClient mocks pb.MetricUpdaterClient interface
type MockMetricUpdaterClient struct {
	mock.Mock
}

func (m *MockMetricUpdaterClient) Updates(ctx context.Context, req *pb.UpdateMetricsRequest, opts ...grpc.CallOption) (*pb.UpdateMetricsResponse, error) {
	args := m.Called(ctx, req)
	if resp, ok := args.Get(0).(*pb.UpdateMetricsResponse); ok {
		return resp, args.Error(1)
	}
	return nil, args.Error(1)
}

func TestMetricGRPCFacade_Updates(t *testing.T) {
	float64Ptr := func(f float64) *float64 { return &f }
	int64Ptr := func(i int64) *int64 { return &i }

	mockClient := new(MockMetricUpdaterClient)
	facade := &MetricGRPCFacade{client: mockClient}

	t.Run("returns nil on empty metrics slice", func(t *testing.T) {
		err := facade.Updates(context.Background(), []*types.Metrics{})
		assert.NoError(t, err)
	})

	t.Run("returns error when grpc client returns error", func(t *testing.T) {
		mockClient.On("Updates", mock.Anything, mock.Anything).Return(nil, errors.New("rpc error")).Once()
		err := facade.Updates(context.Background(), []*types.Metrics{{ID: "metric1", Type: "gauge", Value: float64Ptr(1.23)}})
		assert.EqualError(t, err, "rpc error")
		mockClient.AssertExpectations(t)
	})

	t.Run("returns error when response contains error string", func(t *testing.T) {
		mockClient.On("Updates", mock.Anything, mock.Anything).Return(&pb.UpdateMetricsResponse{Error: "some error"}, nil).Once()
		err := facade.Updates(context.Background(), []*types.Metrics{{ID: "metric2", Type: "counter", Delta: int64Ptr(42)}})
		assert.EqualError(t, err, "some error")
		mockClient.AssertExpectations(t)
	})

	t.Run("successfully sends metrics", func(t *testing.T) {
		mockClient.On("Updates", mock.Anything, mock.MatchedBy(func(req *pb.UpdateMetricsRequest) bool {
			return len(req.Metrics) == 2 && req.Metrics[0].Id == "metric1" && req.Metrics[1].Id == "metric2"
		})).Return(&pb.UpdateMetricsResponse{}, nil).Once()

		metrics := []*types.Metrics{
			{ID: "metric1", Type: "gauge", Value: float64Ptr(3.14)},
			{ID: "metric2", Type: "counter", Delta: int64Ptr(100)},
		}
		err := facade.Updates(context.Background(), metrics)
		assert.NoError(t, err)
		mockClient.AssertExpectations(t)
	})
}

func TestNewMetricGRPCFacade(t *testing.T) {
	t.Run("creates facade successfully with valid address (may fail if no server)", func(t *testing.T) {
		f, err := NewMetricGRPCFacade(WithMetricGRPCServerAddress("localhost:50051"))
		if err != nil {
			t.Logf("expected error dialing gRPC server: %v", err)
			assert.Nil(t, f)
		} else {
			assert.NotNil(t, f)
			assert.Equal(t, "localhost:50051", f.serverAddress)
			assert.NotNil(t, f.conn)
			assert.NotNil(t, f.client)
			_ = f.Close()
		}
	})
}
