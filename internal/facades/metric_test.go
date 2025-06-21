package facades

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
)

func TestNewMetricFacade_DefaultConfig(t *testing.T) {
	_, err := NewMetricHTTPFacade()
	require.NoError(t, err)

}

func TestMetricFacadeOptionFunctions(t *testing.T) {
	tests := []struct {
		name     string
		option   MetricFacadeOption
		expected func(cfg *MetricFacadeConfig)
		assertFn func(t *testing.T, cfg *MetricFacadeConfig)
	}{
		{
			name:   "WithMetricFacadeServerAddress",
			option: WithMetricFacadeServerAddress("localhost:8080"),
			assertFn: func(t *testing.T, cfg *MetricFacadeConfig) {
				assert.Equal(t, "localhost:8080", cfg.ServerAddress)
			},
		},
		{
			name:   "WithMetricFacadeHeader",
			option: WithMetricFacadeHeader("X-Custom-Header"),
			assertFn: func(t *testing.T, cfg *MetricFacadeConfig) {
				assert.Equal(t, "X-Custom-Header", cfg.Header)
			},
		},
		{
			name:   "WithMetricFacadeKey",
			option: WithMetricFacadeKey("mysecretkey"),
			assertFn: func(t *testing.T, cfg *MetricFacadeConfig) {
				assert.Equal(t, "mysecretkey", cfg.Key)
			},
		},
		{
			name:   "WithMetricFacadeCryptoKeyPath",
			option: WithMetricFacadeCryptoKeyPath("/path/to/key.pem"),
			assertFn: func(t *testing.T, cfg *MetricFacadeConfig) {
				assert.Equal(t, "/path/to/key.pem", cfg.CryptoKeyPath)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &MetricFacadeConfig{}
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
	assert.Len(t, sum, 64) // length of sha256 hex string
}

func mustCreateTempFile(t *testing.T, content []byte) string {
	tmpFile, err := os.CreateTemp("", "testkey_*.pem")
	require.NoError(t, err)

	_, err = tmpFile.Write(content)
	require.NoError(t, err)

	err = tmpFile.Close()
	require.NoError(t, err)

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
					tmp := mustCreateTempFile(t, []byte("not a pem"))
					return tmp
				},
				expectError: true,
			},
			{
				name: "wrong PEM block type",
				setupFile: func(t *testing.T) string {
					block := &pem.Block{
						Type:  "WRONG TYPE",
						Bytes: []byte("some bytes"),
					}
					return mustCreateTempFile(t, pem.EncodeToMemory(block))
				},
				expectError: true,
			},
			{
				name: "invalid key bytes in PEM",
				setupFile: func(t *testing.T) string {
					block := &pem.Block{
						Type:  "PUBLIC KEY",
						Bytes: []byte("invalid key bytes"),
					}
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
					block := &pem.Block{
						Type:  "PUBLIC KEY",
						Bytes: pubASN1,
					}
					return mustCreateTempFile(t, pem.EncodeToMemory(block))
				},
				expectError: false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				path := tt.setupFile(t)
				if path != "nonexistent_file.pem" {
					defer os.Remove(path)
				}
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
	})
}

func createTempPublicKeyFile(t *testing.T) string {
	// Generate a test RSA key pair
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.NoError(t, err)

	pubASN1, err := x509.MarshalPKIXPublicKey(&privKey.PublicKey)
	assert.NoError(t, err)

	pubPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubASN1,
	})

	tmpFile, err := os.CreateTemp("", "test_pubkey_*.pem") // updated from ioutil.TempFile
	assert.NoError(t, err)

	_, err = tmpFile.Write(pubPEM)
	assert.NoError(t, err)

	err = tmpFile.Close()
	assert.NoError(t, err)

	return tmpFile.Name()
}

func TestNewMetricFacade_WithValidCryptoKeyPath(t *testing.T) {
	path := createTempPublicKeyFile(t)
	defer os.Remove(path)

	mf, err := NewMetricHTTPFacade(WithMetricFacadeCryptoKeyPath(path))
	assert.NoError(t, err)
	assert.NotNil(t, mf)
	assert.NotNil(t, mf.publicKey)
}

func TestNewMetricFacade_WithInvalidCryptoKeyPath(t *testing.T) {
	// File does not exist
	mf, err := NewMetricHTTPFacade(WithMetricFacadeCryptoKeyPath("/non/existing/file.pem"))
	assert.Error(t, err)
	assert.Nil(t, mf)
}

func TestNewMetricFacade_WithInvalidPEMFile(t *testing.T) {
	// Create temp file with invalid content
	tmpFile, err := os.CreateTemp("", "invalid_pem_*.pem")
	assert.NoError(t, err)
	_, err = tmpFile.Write([]byte("not a valid pem"))
	assert.NoError(t, err)
	err = tmpFile.Close()
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	mf, err := NewMetricHTTPFacade(WithMetricFacadeCryptoKeyPath(tmpFile.Name()))
	assert.Error(t, err)
	assert.Nil(t, mf)
}

func TestMetricFacade_Updates_WithRealHTTPServer(t *testing.T) {
	ctx := context.Background()

	v := float64(42)
	sampleMetrics := []*types.Metrics{
		{ID: "metric1", MType: "gauge", Value: &v},
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
				// Just respond OK
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
				// Check the presence of HMAC header
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
				N: nil, // invalid key will cause encryption error
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
			// Create a real HTTP test server with the handler from the test case
			server := httptest.NewServer(tt.serverHandler)
			defer server.Close()

			// Build MetricFacade with the server URL and config from test case
			mf, err := NewMetricHTTPFacade(
				WithMetricFacadeServerAddress(server.URL),
				WithMetricFacadeKey(tt.key),
				WithMetricFacadeHeader(tt.header),
				func(cfg *MetricFacadeConfig) {
					cfg.CryptoKeyPath = "" // disable loading from path
				},
			)
			assert.NoError(t, err)

			// Override publicKey directly if set in test case
			mf.publicKey = tt.publicKey

			// Call Updates and check result
			err = mf.Updates(ctx, tt.metrics)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
