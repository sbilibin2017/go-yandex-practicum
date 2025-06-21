package facades

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
)

// MetricFacadeConfig holds configuration options for MetricHTTPFacade.
type MetricFacadeConfig struct {
	ServerAddress string // ServerAddress is the address of the metrics server (e.g. "localhost:8080").
	Header        string // Header is the HTTP header name used to send the HMAC hash sum.
	Key           string // Key is the secret key used to compute HMAC hash sums for request bodies.
	CryptoKeyPath string // CryptoKeyPath is the filesystem path to the RSA public key used to encrypt the payload.
}

// MetricHTTPFacade provides methods to send metrics data to a remote HTTP server.
type MetricHTTPFacade struct {
	config    MetricFacadeConfig
	client    *resty.Client
	publicKey *rsa.PublicKey
}

// MetricFacadeOption defines a functional option type for configuring MetricHTTPFacade.
type MetricFacadeOption func(*MetricFacadeConfig)

// WithMetricFacadeServerAddress returns a MetricFacadeOption that sets the server address.
func WithMetricFacadeServerAddress(address string) MetricFacadeOption {
	return func(cfg *MetricFacadeConfig) {
		cfg.ServerAddress = address
	}
}

// WithMetricFacadeHeader returns a MetricFacadeOption that sets the HTTP header name for hash sums.
func WithMetricFacadeHeader(header string) MetricFacadeOption {
	return func(cfg *MetricFacadeConfig) {
		cfg.Header = header
	}
}

// WithMetricFacadeKey returns a MetricFacadeOption that sets the key for HMAC hashing.
func WithMetricFacadeKey(key string) MetricFacadeOption {
	return func(cfg *MetricFacadeConfig) {
		cfg.Key = key
	}
}

// WithMetricFacadeCryptoKeyPath returns a MetricFacadeOption that sets the path to the RSA public key file.
func WithMetricFacadeCryptoKeyPath(path string) MetricFacadeOption {
	return func(cfg *MetricFacadeConfig) {
		cfg.CryptoKeyPath = path
	}
}

// NewMetricHTTPFacade creates a new MetricHTTPFacade configured with the given options.
// It initializes the HTTP client, sets the server base URL, and loads the RSA public key if specified.
func NewMetricHTTPFacade(opts ...MetricFacadeOption) (*MetricHTTPFacade, error) {
	cfg := MetricFacadeConfig{}
	for _, opt := range opts {
		opt(&cfg)
	}

	mf := &MetricHTTPFacade{
		config: cfg,
		client: resty.New(),
	}

	if mf.config.ServerAddress != "" {
		if !strings.HasPrefix(mf.config.ServerAddress, "http://") && !strings.HasPrefix(mf.config.ServerAddress, "https://") {
			mf.config.ServerAddress = "http://" + mf.config.ServerAddress
		}
		mf.client.SetBaseURL(mf.config.ServerAddress)
	}

	if mf.config.CryptoKeyPath != "" {
		pubKey, err := loadPublicKey(mf.config.CryptoKeyPath)
		if err != nil {
			return nil, fmt.Errorf("error loading public key: %w", err)
		}
		mf.publicKey = pubKey
	}

	return mf, nil
}

// Updates sends a batch of metrics to the configured server.
// The metrics slice is JSON marshaled, optionally HMAC hashed and RSA encrypted,
// then sent via HTTP POST to the "/updates/" endpoint.
func (mf *MetricHTTPFacade) Updates(
	ctx context.Context,
	metrics []*types.Metrics,
) error {
	if len(metrics) == 0 {
		return nil
	}

	bodyBytes, err := json.Marshal(metrics)
	if err != nil {
		return err
	}

	var hashSum string
	if mf.config.Key != "" {
		hashSum = calcBodyHashSum(bodyBytes, mf.config.Key)
	}

	if mf.publicKey != nil {
		bodyBytes, err = encryptBody(bodyBytes, mf.publicKey)
		if err != nil {
			return fmt.Errorf("failed to encrypt metrics payload: %w", err)
		}
	}

	err = sendRequest(mf.client, ctx, "/updates/", bodyBytes, mf.config.Header, hashSum)
	if err != nil {
		return err
	}

	return nil
}

// sendRequest sends an HTTP POST request with the given body and headers using the resty client.
func sendRequest(
	client *resty.Client,
	ctx context.Context,
	urlPath string,
	body []byte,
	headerName string,
	hashSum string,
) error {
	req := client.R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept-Encoding", "gzip").
		SetBody(body)

	if headerName != "" && hashSum != "" {
		req.SetHeader(headerName, hashSum)
	}

	resp, err := req.Post(urlPath)
	if err != nil {
		return fmt.Errorf("failed to send metrics: %w", err)
	}
	if resp.IsError() {
		return fmt.Errorf("error response from server for metrics: %s", resp.String())
	}
	return nil
}

// calcBodyHashSum computes an HMAC SHA-256 hash of the given body using the provided key,
// and returns the result as a hex-encoded string.
func calcBodyHashSum(body []byte, key string) string {
	h := hmac.New(sha256.New, []byte(key))
	h.Write(body)
	return hex.EncodeToString(h.Sum(nil))
}

// encryptBody encrypts the given body using RSA OAEP with the provided public key.
// Returns the encrypted bytes or an error.
func encryptBody(body []byte, pubKey *rsa.PublicKey) ([]byte, error) {
	if pubKey == nil {
		return body, nil
	}
	return rsa.EncryptOAEP(sha256.New(), rand.Reader, pubKey, body, nil)
}

// loadPublicKey loads an RSA public key from a PEM encoded file at the given path.
// Returns the public key or an error if the file cannot be read or parsed.
func loadPublicKey(path string) (*rsa.PublicKey, error) {
	keyData, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(keyData)
	if block == nil || block.Type != "PUBLIC KEY" {
		return nil, errors.New("invalid PEM block or type")
	}
	pubKeyInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	pubKey, ok := pubKeyInterface.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("not an RSA public key")
	}
	return pubKey, nil
}
