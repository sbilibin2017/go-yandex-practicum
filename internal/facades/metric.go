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
	"net"
	"os"
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/sbilibin2017/go-yandex-practicum/protos"
)

// MetricHTTPFacade provides methods to send metrics data to a remote HTTP server.
type MetricHTTPFacade struct {
	serverAddress string
	header        string
	key           string
	cryptoKeyPath string

	client    *resty.Client
	publicKey *rsa.PublicKey
}

// MetricFacadeOption defines functional option type for MetricHTTPFacade.
type MetricHTTPFacadeOpt func(*MetricHTTPFacade)

func WithMetricFacadeServerAddress(address string) MetricHTTPFacadeOpt {
	return func(f *MetricHTTPFacade) {
		f.serverAddress = address
	}
}

func WithMetricFacadeHeader(header string) MetricHTTPFacadeOpt {
	return func(f *MetricHTTPFacade) {
		f.header = header
	}
}

func WithMetricFacadeKey(key string) MetricHTTPFacadeOpt {
	return func(f *MetricHTTPFacade) {
		f.key = key
	}
}

func WithMetricFacadeCryptoKeyPath(path string) MetricHTTPFacadeOpt {
	return func(f *MetricHTTPFacade) {
		f.cryptoKeyPath = path
	}
}

// NewMetricHTTPFacade creates a new MetricHTTPFacade configured with options.
func NewMetricHTTPFacade(opts ...MetricHTTPFacadeOpt) (*MetricHTTPFacade, error) {
	f := &MetricHTTPFacade{}
	for _, opt := range opts {
		opt(f)
	}

	f.client = resty.New()

	if f.serverAddress != "" {
		if !strings.HasPrefix(f.serverAddress, "http://") && !strings.HasPrefix(f.serverAddress, "https://") {
			f.serverAddress = "http://" + f.serverAddress
		}
		f.client.SetBaseURL(f.serverAddress)
	}

	if f.cryptoKeyPath != "" {
		pubKey, err := loadPublicKey(f.cryptoKeyPath)
		if err != nil {
			return nil, fmt.Errorf("error loading public key: %w", err)
		}
		f.publicKey = pubKey
	}

	return f, nil
}

func (f *MetricHTTPFacade) Updates(ctx context.Context, metrics []*types.Metrics) error {
	if len(metrics) == 0 {
		return nil
	}

	bodyBytes, err := json.Marshal(metrics)
	if err != nil {
		return err
	}

	var hashSum string
	if f.key != "" {
		hashSum = calcBodyHashSum(bodyBytes, f.key)
	}

	if f.publicKey != nil {
		bodyBytes, err = encryptBody(bodyBytes, f.publicKey)
		if err != nil {
			return fmt.Errorf("failed to encrypt metrics payload: %w", err)
		}
	}

	return sendRequest(f.client, ctx, "/updates/", bodyBytes, f.header, hashSum)
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

// MetricGRPCFacade provides methods to send metrics data to a remote gRPC server.
type MetricGRPCFacade struct {
	serverAddress string

	client pb.MetricUpdaterClient
	conn   *grpc.ClientConn
}

type MetricGRPCFacadeOpt func(*MetricGRPCFacade)

func WithMetricGRPCServerAddress(address string) MetricGRPCFacadeOpt {
	return func(f *MetricGRPCFacade) {
		f.serverAddress = address
	}
}

func NewMetricGRPCFacade(opts ...MetricGRPCFacadeOpt) (*MetricGRPCFacade, error) {
	f := &MetricGRPCFacade{}
	for _, opt := range opts {
		opt(f)
	}

	conn, err := grpc.NewClient(
		f.serverAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(func(ctx context.Context, addr string) (net.Conn, error) {
			d := net.Dialer{}
			return d.DialContext(ctx, "tcp", addr)
		}),
	)
	if err != nil {
		return nil, err
	}

	f.conn = conn
	f.client = pb.NewMetricUpdaterClient(conn)

	return f, nil
}

func (f *MetricGRPCFacade) Updates(ctx context.Context, metrics []*types.Metrics) error {
	if len(metrics) == 0 {
		return nil
	}

	pbMetrics := make([]*pb.Metric, 0, len(metrics))
	for _, m := range metrics {
		pbMetric := &pb.Metric{
			Id:    m.ID,
			Type:  m.Type,
			Value: 0,
			Delta: 0,
		}

		if m.Value != nil {
			pbMetric.Value = *m.Value
		}
		if m.Delta != nil {
			pbMetric.Delta = *m.Delta
		}

		pbMetrics = append(pbMetrics, pbMetric)
	}

	req := &pb.UpdateMetricsRequest{
		Metrics: pbMetrics,
	}

	resp, err := f.client.Updates(ctx, req)
	if err != nil {
		return err
	}
	if resp.Error != "" {
		return errors.New(resp.Error)
	}
	return nil
}

func (f *MetricGRPCFacade) Close() error {
	if f.conn != nil {
		return f.conn.Close()
	}
	return nil
}

// MetricFacade defines interface for sending metrics (Update).
type MetricFacade interface {
	Updates(ctx context.Context, metrics []*types.Metrics) error
}

// AgentFacadeContext holds a facade strategy for sending metrics.
type MetricFacadeContext struct {
	strategy MetricFacade
}

// NewAgentFacadeContext creates a new AgentFacadeContext.
func NewMetricFacadeContext() *MetricFacadeContext {
	return &MetricFacadeContext{}
}

// SetContext sets the MetricFacade strategy to use.
func (c *MetricFacadeContext) SetContext(strategy MetricFacade) {
	c.strategy = strategy
}

// Updates sends metrics using the current strategy.
func (c *MetricFacadeContext) Updates(ctx context.Context, metrics []*types.Metrics) error {
	if c.strategy == nil {
		return nil // or return an error: "no strategy set"
	}
	return c.strategy.Updates(ctx, metrics)
}
