package workers

import (
	"context"
	"crypto/hmac"
	cryptoRand "crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/sbilibin2017/go-yandex-practicum/internal/logger"
	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
	"go.uber.org/zap"
)

var (
	client *resty.Client
)

func init() {
	client = resty.New()

}

type result struct {
	Data []types.Metrics
	Err  error
}

func StartMetricAgent(
	ctx context.Context,
	serverAddress string,
	header string,
	key string,
	cryptoKeyPath string,
	pollInterval int,
	reportInterval int,
	batchSize int,
	rateLimit int,
) error {

	metricsCh := startMetricsPolling(ctx, pollInterval)
	resultsCh := startMetricsReporting(ctx, sendRequest, reportInterval, serverAddress, header, key, cryptoKeyPath, metricsCh, batchSize, rateLimit)
	return logResults(ctx, resultsCh)
}

func startMetricsPolling(ctx context.Context, pollInterval int) <-chan types.Metrics {
	out := make(chan types.Metrics, 100)
	ticker := time.NewTicker(time.Duration(pollInterval) * time.Second)

	go func() {
		defer ticker.Stop()
		defer close(out)

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				runtimeCh := generatorMetrics(ctx, getRuntimeGaugeMetrics())
				counterCh := generatorMetrics(ctx, getRuntimeCounterMetrics())
				gopsutilCh := generatorMetrics(ctx, getGoputilMetrics(ctx))

				for metric := range fanInMetrics(ctx, runtimeCh, counterCh, gopsutilCh) {
					out <- metric
				}
			}
		}
	}()

	return out
}

func startMetricsReporting(
	ctx context.Context,
	handler func(ctx context.Context, urlPath string, body []byte, headerName string, hashSum string) error,
	reportInterval int,
	serverAddress string,
	header string,
	key string,
	cryptoKeyPath string,
	in <-chan types.Metrics,
	batchSize int,
	rateLimit int,
) <-chan result {
	resultsCh := make(chan result, 100)

	ticker := time.NewTicker(time.Duration(reportInterval) * time.Second)

	go func() {
		defer ticker.Stop()
		defer close(resultsCh)

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				results := workerPoolMetricsUpdate(
					ctx,
					handler,
					serverAddress,
					header,
					key,
					cryptoKeyPath,
					in,
					batchSize,
					rateLimit,
				)

				for res := range results {
					resultsCh <- res
				}
			}
		}
	}()

	return resultsCh
}

func getRuntimeGaugeMetrics() []types.Metrics {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	v1 := float64(memStats.Alloc)
	v2 := float64(memStats.BuckHashSys)
	v3 := float64(memStats.Frees)
	v4 := memStats.GCCPUFraction
	v5 := float64(memStats.GCSys)
	v6 := float64(memStats.HeapAlloc)
	v7 := float64(memStats.HeapIdle)
	v8 := float64(memStats.HeapInuse)
	v9 := float64(memStats.HeapObjects)
	v10 := float64(memStats.HeapReleased)
	v11 := float64(memStats.HeapSys)
	v12 := float64(memStats.LastGC)
	v13 := float64(memStats.Lookups)
	v14 := float64(memStats.MCacheInuse)
	v15 := float64(memStats.MCacheSys)
	v16 := float64(memStats.MSpanInuse)
	v17 := float64(memStats.MSpanSys)
	v18 := float64(memStats.Mallocs)
	v19 := float64(memStats.NextGC)
	v20 := float64(memStats.NumForcedGC)
	v21 := float64(memStats.NumGC)
	v22 := float64(memStats.OtherSys)
	v23 := float64(memStats.PauseTotalNs)
	v24 := float64(memStats.StackInuse)
	v25 := float64(memStats.StackSys)
	v26 := float64(memStats.Sys)
	v27 := float64(memStats.TotalAlloc)
	v28 := rand.Float64()

	return []types.Metrics{
		{ID: "Alloc", Type: types.Gauge, Value: &v1},
		{ID: "BuckHashSys", Type: types.Gauge, Value: &v2},
		{ID: "Frees", Type: types.Gauge, Value: &v3},
		{ID: "GCCPUFraction", Type: types.Gauge, Value: &v4},
		{ID: "GCSys", Type: types.Gauge, Value: &v5},
		{ID: "HeapAlloc", Type: types.Gauge, Value: &v6},
		{ID: "HeapIdle", Type: types.Gauge, Value: &v7},
		{ID: "HeapInuse", Type: types.Gauge, Value: &v8},
		{ID: "HeapObjects", Type: types.Gauge, Value: &v9},
		{ID: "HeapReleased", Type: types.Gauge, Value: &v10},
		{ID: "HeapSys", Type: types.Gauge, Value: &v11},
		{ID: "LastGC", Type: types.Gauge, Value: &v12},
		{ID: "Lookups", Type: types.Gauge, Value: &v13},
		{ID: "MCacheInuse", Type: types.Gauge, Value: &v14},
		{ID: "MCacheSys", Type: types.Gauge, Value: &v15},
		{ID: "MSpanInuse", Type: types.Gauge, Value: &v16},
		{ID: "MSpanSys", Type: types.Gauge, Value: &v17},
		{ID: "Mallocs", Type: types.Gauge, Value: &v18},
		{ID: "NextGC", Type: types.Gauge, Value: &v19},
		{ID: "NumForcedGC", Type: types.Gauge, Value: &v20},
		{ID: "NumGC", Type: types.Gauge, Value: &v21},
		{ID: "OtherSys", Type: types.Gauge, Value: &v22},
		{ID: "PauseTotalNs", Type: types.Gauge, Value: &v23},
		{ID: "StackInuse", Type: types.Gauge, Value: &v24},
		{ID: "StackSys", Type: types.Gauge, Value: &v25},
		{ID: "Sys", Type: types.Gauge, Value: &v26},
		{ID: "TotalAlloc", Type: types.Gauge, Value: &v27},
		{ID: "RandomValue", Type: types.Gauge, Value: &v28},
	}
}

func getRuntimeCounterMetrics() []types.Metrics {
	pollCount := int64(1)

	return []types.Metrics{
		{
			ID:    "PollCount",
			Type:  types.Counter,
			Delta: &pollCount,
		},
	}
}

func getGoputilMetrics(ctx context.Context) []types.Metrics {
	var result []types.Metrics

	if vmStat, err := mem.VirtualMemory(); err == nil {
		total := float64(vmStat.Total)
		free := float64(vmStat.Free)

		result = append(result, types.Metrics{
			ID:    "TotalMemory",
			Type:  types.Gauge,
			Value: &total,
		})
		result = append(result, types.Metrics{
			ID:    "FreeMemory",
			Type:  types.Gauge,
			Value: &free,
		})
	}

	if cpuPercents, err := cpu.PercentWithContext(ctx, 0, true); err == nil {
		for i, percent := range cpuPercents {
			p := percent
			id := "CPUutilization" + strconv.Itoa(i)
			result = append(result, types.Metrics{
				ID:    id,
				Type:  types.Gauge,
				Value: &p,
			})
		}
	}

	return result
}

func generatorMetrics(
	ctx context.Context,
	input []types.Metrics,
) chan types.Metrics {
	out := make(chan types.Metrics, 100)

	go func() {
		defer close(out)

		select {
		case <-ctx.Done():
			return
		default:
		}

		for _, item := range input {
			out <- item
		}
	}()

	return out
}

func fanInMetrics(ctx context.Context, resultChs ...chan types.Metrics) chan types.Metrics {
	finalCh := make(chan types.Metrics)

	var wg sync.WaitGroup

	for _, ch := range resultChs {
		chClosure := ch
		wg.Add(1)

		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case data, ok := <-chClosure:
					if !ok {
						return
					}
					finalCh <- data
				}
			}
		}()
	}

	go func() {
		wg.Wait()
		close(finalCh)
	}()

	return finalCh
}

func workerMetricsUpdate(
	ctx context.Context,
	handler func(ctx context.Context, urlPath string, body []byte, headerName string, hashSum string) error,
	serverAddress string,
	header string,
	key string,
	cryptoKeyPath string,
	jobs <-chan types.Metrics,
	batchSize int,
) <-chan result {
	results := make(chan result)

	go func() {
		defer close(results)

		for {
			select {
			case <-ctx.Done():
				return
			case m, ok := <-jobs:
				if !ok {
					return
				}

				batch := []types.Metrics{m}

			collectLoop:
				for len(batch) < batchSize {
					m, ok := <-jobs
					if !ok {
						break collectLoop
					}
					batch = append(batch, m)
				}

				err := updates(ctx, handler, serverAddress, header, key, cryptoKeyPath, batch)

				results <- result{
					Data: batch,
					Err:  err,
				}
			}
		}
	}()

	return results
}

func workerPoolMetricsUpdate(
	ctx context.Context,
	handler func(ctx context.Context, urlPath string, body []byte, headerName string, hashSum string) error,
	serverAddress string,
	header string,
	key string,
	cryptoKeyPath string,
	jobs <-chan types.Metrics,
	batchSize int,
	rateLimit int,
) chan result {
	results := make(chan result)
	var wg sync.WaitGroup
	wg.Add(rateLimit)

	// fan-in all workers results into results chan
	for i := 0; i < rateLimit; i++ {
		go func() {
			defer wg.Done()
			workerResults := workerMetricsUpdate(ctx, handler, serverAddress, header, key, cryptoKeyPath, jobs, batchSize)
			for r := range workerResults {
				results <- r
			}
		}()
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	return results
}

func logResults(ctx context.Context, results <-chan result) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case res, ok := <-results:
			if !ok {
				return ctx.Err()
			}
			if res.Err != nil {
				logger.Log.Error("worker pool task error", zap.Error(res.Err), zap.Any("data", res.Data))
			} else {
				logger.Log.Info("worker pool task success", zap.Any("data", res.Data))
			}
		}
	}
}

func updates(
	ctx context.Context,
	handler func(ctx context.Context, urlPath string, body []byte, headerName string, hashSum string) error,
	serverAddress string,
	header string,
	key string,
	cryptoKeyPath string,
	metrics []types.Metrics,
) error {
	if !strings.HasPrefix(serverAddress, "http://") && !strings.HasPrefix(serverAddress, "https://") {
		serverAddress = "http://" + serverAddress
	}

	client.SetBaseURL(serverAddress)

	var pubKey *rsa.PublicKey
	var err error
	if cryptoKeyPath != "" {
		pubKey, err = loadPublicKey(cryptoKeyPath)
		if err != nil {
			return fmt.Errorf("failed to load public key: %w", err)
		}
	}

	if len(metrics) == 0 {
		return nil
	}

	bodyBytes, err := json.Marshal(metrics)
	if err != nil {
		return err
	}

	var hashSum string
	if key != "" {
		hashSum = calcBodyHashSum(bodyBytes, key)
	}

	if pubKey != nil {
		bodyBytes, err = encryptBody(bodyBytes, pubKey)
		if err != nil {
			return fmt.Errorf("failed to encrypt metrics payload: %w", err)
		}
	}

	return handler(ctx, "/updates/", bodyBytes, header, hashSum)
}

func sendRequest(
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

	hostIP := extractIP(urlPath)
	if hostIP != "" {
		req.SetHeader("X-Real-IP", hostIP)
	}

	if headerName != "" && hashSum != "" {
		req.SetHeader(headerName, hashSum)
	}

	resp, err := req.Post(urlPath)
	if err != nil {
		return fmt.Errorf("failed to send metrics: %w", err)
	}

	if resp.StatusCode() >= 400 {
		return fmt.Errorf("server error %d: %s", resp.StatusCode(), resp.String())
	}

	return nil
}

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
		return nil, errors.New("not RSA public key")
	}

	return pubKey, nil
}

func calcBodyHashSum(body []byte, key string) string {
	mac := hmac.New(sha256.New, []byte(key))
	mac.Write(body)
	return hex.EncodeToString(mac.Sum(nil))
}

func encryptBody(body []byte, pubKey *rsa.PublicKey) ([]byte, error) {
	encryptedBytes, err := rsa.EncryptOAEP(sha256.New(), cryptoRand.Reader, pubKey, body, nil)
	if err != nil {
		return nil, err
	}
	return encryptedBytes, nil
}

func extractIP(address string) string {
	host := address
	if strings.Contains(address, ":") {
		host, _, _ = net.SplitHostPort(address)
	}

	if host == "" || host == "localhost" {
		return ""
	}

	return host
}
