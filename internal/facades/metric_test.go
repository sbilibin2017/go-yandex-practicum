package facades

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/go-resty/resty/v2"

	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetricFacade_Updates_EmptyMetrics(t *testing.T) {
	client := resty.New()
	mf, err := NewMetricFacade(client, "http://localhost", "X-Test-Header", "testkey", "")
	require.NoError(t, err)

	err = mf.Updates(context.Background(), []types.Metrics{})
	assert.NoError(t, err, "should not fail on empty metrics slice")
}

func TestMetricFacade_Updates_SuccessWithKey(t *testing.T) {
	metrics := []types.Metrics{
		{
			MetricID: types.MetricID{
				ID:   "metric1",
				Type: types.GaugeMetricType,
			},
			Value: ptrFloat64(123.45),
		},
		{
			MetricID: types.MetricID{
				ID:   "metric2",
				Type: types.CounterMetricType,
			},
			Delta: ptrInt64(10),
		},
	}

	bodyBytes, err := json.Marshal(metrics)
	assert.NoError(t, err)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/updates/", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "gzip", r.Header.Get("Accept-Encoding"))

		// Проверяем HMAC-заголовок
		expectedHmac := computeHmac256(bodyBytes, []byte("testkey"))
		assert.Equal(t, expectedHmac, r.Header.Get("X-Signature"))

		// Проверяем тело запроса
		var received []types.Metrics
		err := json.NewDecoder(r.Body).Decode(&received)
		assert.NoError(t, err)
		assert.Equal(t, metrics, received)

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := resty.New()
	mf, err := NewMetricFacade(client, server.URL, "X-Signature", "testkey", "")
	require.NoError(t, err)

	err = mf.Updates(context.Background(), metrics)
	assert.NoError(t, err)
}

func TestMetricFacade_Updates_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "server error", http.StatusInternalServerError)
	}))
	defer server.Close()

	client := resty.New()
	mf, err := NewMetricFacade(client, server.URL, "X-Signature", "testkey", "")
	require.NoError(t, err)

	metrics := []types.Metrics{
		{
			MetricID: types.MetricID{
				ID:   "m1",
				Type: types.GaugeMetricType,
			},
			Value: ptrFloat64(1.23),
		},
	}

	err = mf.Updates(context.Background(), metrics)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error response from server")
}

func TestMetricFacade_Updates_NetworkError(t *testing.T) {
	client := resty.New()
	mf, err := NewMetricFacade(client, "http://invalid.invalid", "X-Signature", "testkey", "")
	require.NoError(t, err)

	metrics := []types.Metrics{
		{
			MetricID: types.MetricID{
				ID:   "m1",
				Type: types.CounterMetricType,
			},
			Delta: ptrInt64(5),
		},
	}

	err = mf.Updates(context.Background(), metrics)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to send metrics")
}

// --- Вспомогательные функции ---

func ptrFloat64(v float64) *float64 {
	return &v
}

func ptrInt64(v int64) *int64 {
	return &v
}

func computeHmac256(message, key []byte) string {
	h := hmac.New(sha256.New, key)
	h.Write(message)
	return hex.EncodeToString(h.Sum(nil))
}

func TestNewMetricFacade_AddsHTTPPrefix(t *testing.T) {
	client := resty.New()

	// Случай без префикса
	addr := "localhost:8080"
	mf, err := NewMetricFacade(client, addr, "X-Header", "secret", "")
	require.NoError(t, err)
	require.Equal(t, "http://"+addr, mf.serverAddress)

	// Случай с http://
	addr2 := "http://example.com"
	mf2, err := NewMetricFacade(client, addr2, "X-Header", "secret", "")
	require.NoError(t, err)
	require.Equal(t, addr2, mf2.serverAddress)

	// Случай с https://
	addr3 := "https://secure.example.com"
	mf3, err := NewMetricFacade(client, addr3, "X-Header", "secret", "")
	require.NoError(t, err)
	require.Equal(t, addr3, mf3.serverAddress)
}

func TestNewMetricFacade_LoadsPublicKey(t *testing.T) {
	// 1. Создаём временный PEM-файл с валидным RSA-ключом
	keyPEM := `-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAvD8uQ+F3k8YNUrj3eKy1
p4eOXDDAI3zWyZjW1I2Ku6Yu+eyZsn0tSy6jvyi5HXzXYc9iRgq1YueX3Vdt/vZx
o/UMV5+oYHhnw2lj4YFDTKzKXPnkg5oYlbF7HD3MPcRzyxt7u2ePSMG+qk64pJZC
Zslf0hwZtKzlfykU3oRKh8U1OJqkS3t+0aFb+UvqcUXk9+ScZqD8vAVaZFM/xZqj
Zmsu4X3lURSn9gbn7eEkZcU4+SgbiVzEXeD3IHV2tzvi0+GP2TYGzA3hF2We+dA2
l9uGvciThK+vsyqJcOZ8RpsGTiEmxayZoLZtkEXeEZ5PCVLyJzCv9LgxDbPEpHTO
YwIDAQAB
-----END PUBLIC KEY-----`

	tmpFile, err := os.CreateTemp("", "testkey_*.pem")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(keyPEM)
	require.NoError(t, err)
	tmpFile.Close()

	// 2. Инициализация MetricFacade с этим ключом
	client := resty.New()
	mf, err := NewMetricFacade(client, "localhost:8080", "X-Header", "testkey", tmpFile.Name())
	assert.NoError(t, err)
	assert.NotNil(t, mf.publicKey, "publicKey should be loaded")
}

func TestNewMetricFacade_InvalidPublicKeyPath(t *testing.T) {
	client := resty.New()
	_, err := NewMetricFacade(client, "localhost:8080", "X-Header", "testkey", "/non/existent/key.pem")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to load public key")
}

func TestLoadPublicKey_Errors(t *testing.T) {
	// 1. Файл не существует
	_, err := loadPublicKey("/non/existent/file.pem")
	assert.Error(t, err)

	// 2. PEM блок пустой или неправильного типа
	tmpFile, err := os.CreateTemp("", "empty_pem_*.pem")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	// Записываем некорректный PEM без BEGIN PUBLIC KEY
	_, err = tmpFile.WriteString("INVALID PEM DATA")
	require.NoError(t, err)
	tmpFile.Close()

	_, err = loadPublicKey(tmpFile.Name())
	assert.Error(t, err)

	// 3. PEM с правильным заголовком, но поврежденный ключ (невалидные байты)
	tmpFile2, err := os.CreateTemp("", "bad_key_*.pem")
	require.NoError(t, err)
	defer os.Remove(tmpFile2.Name())

	_, err = tmpFile2.WriteString("-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQ\n-----END PUBLIC KEY-----")
	require.NoError(t, err)
	tmpFile2.Close()

	_, err = loadPublicKey(tmpFile2.Name())
	assert.Error(t, err)

	// 4. PEM с не-RSA ключом (например, EC ключ)
	ecKeyPEM := `-----BEGIN PUBLIC KEY-----
MHYwEAYHKoZIzj0CAQYFK4EEACIDYgAE87Vb6E3duPx4lfz7Kf0Lw/JT99lIWhx
58T5JZxbz6t4k+Gdt2g9dzpHgwiuEs0v97fJ7RGqVwA8+q0bD46q+vZkGUu6WXip
s1y4e5sRmD8V5PxU3xZm+rs2T54DmKsm
-----END PUBLIC KEY-----`

	tmpFile3, err := os.CreateTemp("", "ec_key_*.pem")
	require.NoError(t, err)
	defer os.Remove(tmpFile3.Name())

	_, err = tmpFile3.WriteString(ecKeyPEM)
	require.NoError(t, err)
	tmpFile3.Close()

	_, err = loadPublicKey(tmpFile3.Name())
	assert.Error(t, err)

}

func TestCalcBodyHashSum(t *testing.T) {
	body := []byte(`{"metric":"value"}`)
	key := "secretkey"

	hash1 := calcBodyHashSum(body, key)
	hash2 := calcBodyHashSum(body, key)

	// Same input and key should produce same hash
	assert.Equal(t, hash1, hash2)

	// Different key produces different hash
	otherKey := "otherkey"
	hash3 := calcBodyHashSum(body, otherKey)
	assert.NotEqual(t, hash1, hash3)

	// Different body produces different hash
	otherBody := []byte(`{"metric":"other"}`)
	hash4 := calcBodyHashSum(otherBody, key)
	assert.NotEqual(t, hash1, hash4)
}

func TestEncryptBody_NoPublicKey(t *testing.T) {
	body := []byte(`{"metric":"value"}`)

	encrypted, err := encryptBody(body, nil)
	assert.NoError(t, err)
	assert.Equal(t, body, encrypted) // unchanged if no pubKey
}

func TestEncryptBody_WithPublicKey(t *testing.T) {
	body := []byte(`{"metric":"value"}`)

	// Generate RSA key pair for test
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.NoError(t, err)
	pubKey := &privKey.PublicKey

	encrypted, err := encryptBody(body, pubKey)
	assert.NoError(t, err)
	assert.NotEqual(t, body, encrypted) // encrypted should differ from original

	// Decrypt to verify content correctness
	decrypted, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, privKey, encrypted, nil)
	assert.NoError(t, err)
	assert.Equal(t, body, decrypted)
}

func TestSendRequest_Success(t *testing.T) {
	// Mock server to validate request
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/updates/", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "gzip", r.Header.Get("Accept-Encoding"))
		assert.Equal(t, "somehash", r.Header.Get("X-Test-Header"))

		// Properly read the entire body
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		assert.JSONEq(t, `{"key":"value"}`, string(body))

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := resty.New().SetBaseURL(server.URL)

	err := sendRequest(
		client,
		context.Background(),
		"/updates/",
		[]byte(`{"key":"value"}`),
		"X-Test-Header",
		"somehash",
	)
	assert.NoError(t, err)
}

func TestMetricFacade_Updates_EncryptionSuccess(t *testing.T) {
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.NoError(t, err)
	pubKey := &privKey.PublicKey

	client := resty.New()
	mf := &MetricFacade{
		client:    client,
		key:       "testkey",
		header:    "X-Signature",
		publicKey: pubKey,
	}

	metrics := []types.Metrics{}

	err = mf.Updates(context.Background(), metrics)
	assert.NoError(t, err)
}
