package http

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
	"net"
	"net/url"
	"os"
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
)

// MetricsFacade provides an interface for sending metrics updates to a server,
// including optional payload hashing and encryption.
type MetricsFacade struct {
	serverAddress string
	headerName    string
	hashKey       string
	cryptoKeyPath string
	client        *resty.Client
}

// NewMetricsFacade creates a new MetricsFacade instance with the given server address,
// HTTP header name for hash authentication, HMAC hash key, and optional path to a
// public key for encrypting the payload.
func NewMetricsFacade(
	serverAddress string,
	headerName string,
	hashKey string,
	cryptoKeyPath string,
) (*MetricsFacade, error) {
	return &MetricsFacade{
		serverAddress: serverAddress,
		headerName:    headerName,
		hashKey:       hashKey,
		cryptoKeyPath: cryptoKeyPath,
		client:        resty.New(),
	}, nil
}

// Updates sends a batch of metrics to the configured server endpoint.
// If the metrics slice is empty, it returns immediately without error.
// Metrics payload may be hashed and/or encrypted depending on configuration.
func (m *MetricsFacade) Updates(ctx context.Context, metrics []*types.Metrics) error {
	if len(metrics) == 0 {
		return nil
	}
	bodyBytes, hashSum, err := prepareRequest(metrics, m.hashKey, m.cryptoKeyPath)
	if err != nil {
		return err
	}
	fullURL := prepareURL(prepareServerAddress(m.serverAddress), "/updates/")
	return sendRequest(ctx, m.client, fullURL, bodyBytes, m.headerName, hashSum)
}

// prepareServerAddress ensures the server address has a scheme prefix ("http://" by default).
func prepareServerAddress(serverAddress string) string {
	if !strings.HasPrefix(serverAddress, "http://") && !strings.HasPrefix(serverAddress, "https://") {
		return "http://" + serverAddress
	}
	return serverAddress
}

// prepareURL concatenates the server address and endpoint path into a single URL string,
// normalizing slashes as needed.
func prepareURL(serverAddress string, serverEndpoint string) string {
	serverAddress = strings.TrimRight(serverAddress, "/")
	serverEndpoint = "/" + strings.TrimLeft(serverEndpoint, "/")
	return serverAddress + serverEndpoint
}

// prepareRequest marshals metrics into JSON, optionally loads a public key for encryption,
// calculates a hash signature if hashKey is provided, and returns the prepared body and hash sum.
func prepareRequest(metrics []*types.Metrics, hashKey, cryptoKeyPath string) ([]byte, string, error) {
	bodyBytes, err := json.Marshal(metrics)
	if err != nil {
		return nil, "", fmt.Errorf("failed to marshal metrics: %w", err)
	}
	var pubKey *rsa.PublicKey
	if cryptoKeyPath != "" {
		pubKey, err = loadPublicKey(cryptoKeyPath)
		if err != nil {
			return nil, "", fmt.Errorf("failed to load public key: %w", err)
		}
	}
	var hashSum string
	if hashKey != "" {
		hashSum = calcBodyHashSum(bodyBytes, hashKey)
	}
	if pubKey != nil {
		bodyBytes, err = encryptBody(bodyBytes, pubKey)
		if err != nil {
			return nil, "", fmt.Errorf("failed to encrypt metrics payload: %w", err)
		}
	}
	return bodyBytes, hashSum, nil
}

// sendRequest performs an HTTP POST request with the given context, client, URL, body,
// and optional authentication header containing the hash sum.
func sendRequest(
	ctx context.Context,
	client *resty.Client,
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
	hostIP := extractIP(urlPath)
	if hostIP != "" {
		req.SetHeader("X-Real-IP", hostIP)
	}
	if headerName != "" && hashSum != "" {
		req.SetHeader(headerName, hashSum)
	}
	resp, err := req.Post(urlPath)
	if err != nil {
		return err
	}
	if resp.StatusCode() >= 400 {
		return fmt.Errorf("server error %d: %s", resp.StatusCode(), resp.String())
	}
	return nil
}

// loadPublicKey reads an RSA public key from a PEM-encoded file at the given path.
func loadPublicKey(path string) (*rsa.PublicKey, error) {
	keyBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(keyBytes)
	if block == nil {
		return nil, errors.New("failed to decode PEM block")
	}
	pubInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	pubKey, ok := pubInterface.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("not an RSA public key")
	}
	return pubKey, nil
}

// calcBodyHashSum calculates an HMAC-SHA256 hash of the given body using the provided key
// and returns the hexadecimal encoded hash string.
func calcBodyHashSum(body []byte, key string) string {
	mac := hmac.New(sha256.New, []byte(key))
	mac.Write(body)
	return hex.EncodeToString(mac.Sum(nil))
}

// encryptBody encrypts the given body bytes using RSA OAEP with the provided public key.
func encryptBody(body []byte, pubKey *rsa.PublicKey) ([]byte, error) {
	encryptedBytes, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, pubKey, body, nil)
	if err != nil {
		return nil, err
	}
	return encryptedBytes, nil
}

// extractIP attempts to extract the host IP or hostname from a URL or raw address string.
// Returns an empty string if the host is "localhost" or cannot be determined.
func extractIP(address string) string {
	if strings.HasPrefix(address, "http://") || strings.HasPrefix(address, "https://") {
		parsedURL, err := url.Parse(address)
		if err != nil {
			return ""
		}
		host := parsedURL.Host
		if host == "" {
			host = address
		}
		if strings.Contains(host, ":") {
			host, _, _ = net.SplitHostPort(host)
		}
		if host == "" || host == "localhost" {
			return ""
		}
		return host
	}
	host := address
	if strings.Contains(host, ":") {
		host, _, _ = net.SplitHostPort(host)
	}
	if host == "" || host == "localhost" {
		return ""
	}
	return host
}
