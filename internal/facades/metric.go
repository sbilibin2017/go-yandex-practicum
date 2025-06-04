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
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
)

type MetricFacade struct {
	client        *resty.Client
	serverAddress string
	header        string
	key           string
	cryptoKeyPath string
	publicKey     *rsa.PublicKey
}

func NewMetricFacade(
	client *resty.Client,
	serverAddress string,
	header string,
	key string,
	cryptoKeyPath string,
) (*MetricFacade, error) {
	if !strings.HasPrefix(serverAddress, "http://") && !strings.HasPrefix(serverAddress, "https://") {
		serverAddress = "http://" + serverAddress
	}

	client.
		SetBaseURL(serverAddress).
		SetRetryCount(3).
		AddRetryCondition(func(r *resty.Response, err error) bool {
			return err != nil
		}).
		SetRetryAfter(func(client *resty.Client, resp *resty.Response) (time.Duration, error) {
			switch resp.Request.Attempt {
			case 1:
				return 1 * time.Second, nil
			case 2:
				return 3 * time.Second, nil
			case 3:
				return 5 * time.Second, nil
			default:
				return 0, nil
			}
		})

	var pubKey *rsa.PublicKey
	var err error
	if cryptoKeyPath != "" {
		pubKey, err = loadPublicKey(cryptoKeyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load public key: %w", err)
		}
	}

	return &MetricFacade{
		client:        client,
		serverAddress: serverAddress,
		header:        header,
		key:           key,
		cryptoKeyPath: cryptoKeyPath,
		publicKey:     pubKey,
	}, nil
}

func (mf *MetricFacade) Updates(
	ctx context.Context,
	metrics []types.Metrics,
) error {
	if len(metrics) == 0 {
		return nil
	}

	bodyBytes, err := json.Marshal(metrics)
	if err != nil {
		return err
	}

	var hashSum string
	if mf.key != "" {
		hashSum = calcBodyHashSum(bodyBytes, mf.key)
	}

	if mf.publicKey != nil {
		bodyBytes, err = encryptBody(bodyBytes, mf.publicKey)
		if err != nil {
			return fmt.Errorf("failed to encrypt metrics payload: %w", err)
		}
	}

	return sendRequest(mf.client, ctx, "/updates/", bodyBytes, mf.header, hashSum)
}

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

func calcBodyHashSum(body []byte, key string) string {
	h := hmac.New(sha256.New, []byte(key))
	h.Write(body)
	return hex.EncodeToString(h.Sum(nil))
}

func encryptBody(body []byte, pubKey *rsa.PublicKey) ([]byte, error) {
	if pubKey == nil {
		return body, nil
	}
	return rsa.EncryptOAEP(sha256.New(), rand.Reader, pubKey, body, nil)
}

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
