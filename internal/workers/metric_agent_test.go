package workers

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
	"github.com/stretchr/testify/assert"
)

func TestGetRuntimeGaugeMetrics(t *testing.T) {
	metrics := getRuntimeGaugeMetrics()

	expectedIDs := []string{
		"Alloc", "BuckHashSys", "Frees", "GCCPUFraction", "GCSys", "HeapAlloc",
		"HeapIdle", "HeapInuse", "HeapObjects", "HeapReleased", "HeapSys",
		"LastGC", "Lookups", "MCacheInuse", "MCacheSys", "MSpanInuse",
		"MSpanSys", "Mallocs", "NextGC", "NumForcedGC", "NumGC", "OtherSys",
		"PauseTotalNs", "StackInuse", "StackSys", "Sys", "TotalAlloc",
		"RandomValue",
	}

	assert.Equal(t, len(expectedIDs), len(metrics), "metrics length mismatch")

	for i := range expectedIDs {
		assert.Equal(t, expectedIDs[i], metrics[i].ID, "metric ID mismatch at index %d", i)
		assert.Equal(t, types.Gauge, metrics[i].Type, "metric Type mismatch at index %d", i)
		assert.NotNil(t, metrics[i].Value, "metric Value should not be nil at index %d", i)
	}
}

func TestGetRuntimeCounterMetrics(t *testing.T) {
	metrics := getRuntimeCounterMetrics()

	// Expect exactly 1 metric
	assert.Len(t, metrics, 1, "metrics slice length should be 1")

	metric := metrics[0]
	assert.Equal(t, "PollCount", metric.ID, "metric ID mismatch")
	assert.Equal(t, types.Counter, metric.Type, "metric Type mismatch")
	assert.NotNil(t, metric.Delta, "metric Delta should not be nil")
	assert.Equal(t, int64(1), *metric.Delta, "metric Delta value mismatch")
}

func TestGetGoputilMetrics(t *testing.T) {
	ctx := context.Background()
	metrics := getGoputilMetrics(ctx)

	assert.NotEmpty(t, metrics, "metrics should not be empty")

	// Check that TotalMemory and FreeMemory metrics exist and have values
	foundTotalMem := false
	foundFreeMem := false
	cpuCount := 0

	for _, metric := range metrics {
		switch metric.ID {
		case "TotalMemory":
			foundTotalMem = true
			assert.Equal(t, types.Gauge, metric.Type)
			assert.NotNil(t, metric.Value)
			assert.Greater(t, *metric.Value, 0.0)
		case "FreeMemory":
			foundFreeMem = true
			assert.Equal(t, types.Gauge, metric.Type)
			assert.NotNil(t, metric.Value)
			assert.GreaterOrEqual(t, *metric.Value, 0.0)
		default:
			// Check for CPU utilization metrics named "CPUutilizationN"
			if strings.HasPrefix(metric.ID, "CPUutilization") {
				cpuCount++
				assert.Equal(t, types.Gauge, metric.Type)
				assert.NotNil(t, metric.Value)
				assert.GreaterOrEqual(t, *metric.Value, 0.0)
				assert.LessOrEqual(t, *metric.Value, 100.0)
			}
		}
	}

	assert.True(t, foundTotalMem, "TotalMemory metric not found")
	assert.True(t, foundFreeMem, "FreeMemory metric not found")
	assert.Greater(t, cpuCount, 0, "no CPU utilization metrics found")
}

func TestGeneratorMetrics(t *testing.T) {
	ctx := context.Background()

	ptrFloat64 := func(f float64) *float64 { return &f }
	ptrInt64 := func(i int64) *int64 { return &i }

	input := []types.Metrics{
		{ID: "metric1", Type: types.Gauge, Value: ptrFloat64(1.1)},
		{ID: "metric2", Type: types.Counter, Delta: ptrInt64(2)},
	}

	ch := generatorMetrics(ctx, input)

	var output []types.Metrics
	for metric := range ch {
		output = append(output, metric)
	}

	assert.Equal(t, len(input), len(output), "output length should match input length")

	for i, metric := range input {
		assert.Equal(t, metric.ID, output[i].ID)
		assert.Equal(t, metric.Type, output[i].Type)
		if metric.Value != nil {
			assert.NotNil(t, output[i].Value)
			assert.Equal(t, *metric.Value, *output[i].Value)
		}
		if metric.Delta != nil {
			assert.NotNil(t, output[i].Delta)
			assert.Equal(t, *metric.Delta, *output[i].Delta)
		}
	}
}

func TestGeneratorMetrics_ContextDone(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	ptrFloat64 := func(f float64) *float64 { return &f }

	input := []types.Metrics{
		{ID: "metric1", Type: types.Gauge, Value: ptrFloat64(1.1)},
	}

	ch := generatorMetrics(ctx, input)

	_, ok := <-ch
	assert.False(t, ok, "channel should be closed immediately when context is done")
}

func TestFanInMetrics(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ptrFloat64 := func(f float64) *float64 { return &f }
	ptrInt64 := func(i int64) *int64 { return &i }

	ch1 := make(chan types.Metrics, 2)
	ch2 := make(chan types.Metrics, 2)

	m1 := types.Metrics{ID: "m1", Type: types.Gauge, Value: ptrFloat64(1)}
	m2 := types.Metrics{ID: "m2", Type: types.Counter, Delta: ptrInt64(2)}

	ch1 <- m1
	ch2 <- m2

	close(ch1)
	close(ch2)

	out := fanInMetrics(ctx, ch1, ch2)

	var results []types.Metrics
	for m := range out {
		results = append(results, m)
	}

	assert.Len(t, results, 2)

	ids := []string{results[0].ID, results[1].ID}
	assert.Contains(t, ids, "m1")
	assert.Contains(t, ids, "m2")
}
func TestFanInMetrics_ContextDone(t *testing.T) {
	// Create a cancelable context and cancel it immediately
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Create an input channel that would otherwise send metrics
	ch := make(chan types.Metrics)

	out := fanInMetrics(ctx, ch)

	// Since context is canceled, the output channel should close quickly with no data
	select {
	case _, ok := <-out:
		assert.False(t, ok, "output channel should be closed immediately when context is done")
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timeout waiting for output channel to close")
	}
}

func TestSendRequest(t *testing.T) {
	// Setup a test HTTP server that validates request and returns 200 OK
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "gzip", r.Header.Get("Accept-Encoding"))
		assert.Equal(t, "testhash", r.Header.Get("X-Test-Hash"))

		// Read and verify body
		buf := new(bytes.Buffer)
		_, err := buf.ReadFrom(r.Body)
		assert.NoError(t, err)
		assert.Equal(t, `{"key":"value"}`, buf.String())

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	defer ts.Close()

	ctx := context.Background()
	body := []byte(`{"key":"value"}`)
	headerName := "X-Test-Hash"
	hashSum := "testhash"

	err := sendRequest(ctx, ts.URL, body, headerName, hashSum)
	assert.NoError(t, err)
}

func TestSendRequest_ServerError(t *testing.T) {
	// Setup a test HTTP server that returns 500 Internal Server Error
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "internal error", http.StatusInternalServerError)
	}))
	defer ts.Close()

	ctx := context.Background()
	err := sendRequest(ctx, ts.URL, []byte{}, "", "") // <-- pass empty body, not nil
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "server error 500")
}

func TestSendRequest_RequestError(t *testing.T) {
	// Simulate request error by using invalid URL
	ctx := context.Background()
	err := sendRequest(ctx, "http://invalid-url", nil, "", "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to send metrics")
}

func TestLoadPublicKey(t *testing.T) {
	// 1) Setup valid RSA public key PEM file (PKIX format)
	validPEM := `-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAvkJ9YHPJfgov8mjmiBQA
GhcXPoLZxpsvRJ0luwOBr6q5uVv7yJZVaZCdj5ikQwQKlzpmjoY83Pq2mJ0CuXvV
pmJ8vt3gjFvShHt1+I1gB8k7KlIjXv6rl69wQJrcBldoIm5YbUN2Gf9Q39mNs8+h
zyYZ0NAD8gWdFtHq2xZvqzFlL7Jd6cqLrWtIGxOdvJ7/U7cN5A6Jt+q34LQVBBJX
KlV7IBnHV4bqOXH+htP2qQGXEj5uyE+3XkyNu0rZfXzK42Kl8MEZC9d2uvihdK2M
qO0noFS6oiTARzYlXGMaMXYD0fM0W55epGxl6bwR9MwZbEgC3yD+P15QGOyYPciW
LwIDAQAB
-----END PUBLIC KEY-----`

	tmpDir := t.TempDir()

	validKeyFile := filepath.Join(tmpDir, "valid_rsa_pub.pem")
	err := os.WriteFile(validKeyFile, []byte(validPEM), 0644)
	assert.NoError(t, err)

	// 2) Setup invalid PEM file (just garbage text)
	invalidPEMFile := filepath.Join(tmpDir, "invalid.pem")
	err = os.WriteFile(invalidPEMFile, []byte("not a pem key"), 0644)
	assert.NoError(t, err)

	// 3) Setup non-RSA PEM file (e.g. EC public key)
	ecPEM := `-----BEGIN PUBLIC KEY-----
MHYwEAYHKoZIzj0CAQYFK4EEACIDYgAEzRqaJ63zIzwEGP6EMISklcltLh3Nkd3m
sfR5OeX+bTF1qQnRDBkdoUb+LHlfnFlqPULyXXnFp/6Rk5HXs7/KsQYmT0VPc7tv
r59vHjEeV+3rCK93Fj3iTJGpnqTDp/BX
-----END PUBLIC KEY-----`

	ecKeyFile := filepath.Join(tmpDir, "ec_pub.pem")
	err = os.WriteFile(ecKeyFile, []byte(ecPEM), 0644)
	assert.NoError(t, err)

	tests := []struct {
		name       string
		path       string
		expectErr  bool
		errMessage string
	}{
		{
			name:      "Valid RSA Public Key",
			path:      validKeyFile,
			expectErr: false,
		},
		{
			name:       "File Not Exist",
			path:       "nonexistent.pem",
			expectErr:  true,
			errMessage: "no such file or directory",
		},
		{
			name:       "Invalid PEM",
			path:       invalidPEMFile,
			expectErr:  true,
			errMessage: "failed to decode PEM block",
		},
		{
			name:       "Not RSA Public Key",
			path:       ecKeyFile,
			expectErr:  true,
			errMessage: "failed to unmarshal elliptic curve point", // matches actual parser error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pubKey, err := loadPublicKey(tt.path)
			if tt.expectErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMessage)
				assert.Nil(t, pubKey)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, pubKey)
				_, ok := interface{}(pubKey).(*rsa.PublicKey)
				assert.True(t, ok)
			}
		})
	}
}

func TestCalcBodyHashSum(t *testing.T) {
	body := []byte(`{"metric":"value"}`)
	key := "secretkey"

	expectedMac := hmac.New(sha256.New, []byte(key))
	expectedMac.Write(body)
	expectedHash := hex.EncodeToString(expectedMac.Sum(nil))

	gotHash := calcBodyHashSum(body, key)

	assert.Equal(t, expectedHash, gotHash)
}

func TestEncryptBody(t *testing.T) {
	// Generate RSA key pair for testing
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.NoError(t, err)
	publicKey := &privateKey.PublicKey

	plainText := []byte("test data for encryption")

	// Encrypt the plaintext using your function
	encrypted, err := encryptBody(plainText, publicKey)
	assert.NoError(t, err)
	assert.NotNil(t, encrypted)
	assert.NotEqual(t, plainText, encrypted)

	// Decrypt to verify correct encryption
	decrypted, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, privateKey, encrypted, nil)
	assert.NoError(t, err)
	assert.Equal(t, plainText, decrypted)
}

func TestLogResults_ContextDoneReturnsError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	resultsCh := make(chan result)

	err := logResults(ctx, resultsCh)

	assert.ErrorIs(t, err, context.Canceled)
}

func TestLogResults_ClosedResultsChannelReturnsNil(t *testing.T) {
	ctx := context.Background()
	resultsCh := make(chan result)

	go func() {
		close(resultsCh)
	}()

	err := logResults(ctx, resultsCh)

	assert.NoError(t, err)
}

func TestLogResults_ExitsAfterProcessingResults(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	resultsCh := make(chan result, 1)
	resultsCh <- result{Err: nil} // no error

	go func() {
		time.Sleep(50 * time.Millisecond)
		close(resultsCh)
	}()

	err := logResults(ctx, resultsCh)

	assert.NoError(t, err)
}

func TestLogResults_HandlesErrorResult(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	resultsCh := make(chan result, 1)

	// Send a result with an error to trigger error logging branch
	resultsCh <- result{
		Data: nil,
		Err:  errors.New("some error"),
	}
	close(resultsCh)

	// It should run without panic and return nil after channel close
	err := logResults(ctx, resultsCh)
	assert.NoError(t, err)
}

func TestLogResults_HandlesNilErrorResult(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	resultsCh := make(chan result, 1)

	// Send a result without error to trigger success logging branch
	resultsCh <- result{
		Data: nil,
		Err:  nil,
	}
	close(resultsCh)

	err := logResults(ctx, resultsCh)
	assert.NoError(t, err)
}

func testServerHandler(t *testing.T, expectedHeader string, expectedHash string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Check path
		assert.Equal(t, "/updates/", r.URL.Path)

		// Check header if expectedHeader is set
		if expectedHeader != "" && expectedHash != "" {
			got := r.Header.Get(expectedHeader)
			assert.Equal(t, expectedHash, got)
		}

		body, err := io.ReadAll(r.Body)
		assert.NoError(t, err)
		defer r.Body.Close()

		// Make sure body is not empty
		assert.NotEmpty(t, body)

		w.WriteHeader(http.StatusOK)
	}
}

func TestUpdates_WithRealHTTPServer(t *testing.T) {
	assert := assert.New(t)

	// Prepare sample metrics using your Metrics struct
	metricValue := 42.0
	metrics := []types.Metrics{
		{
			ID:    "TestMetric",
			Type:  "gauge", // adjust to your valid type string
			Value: &metricValue,
		},
	}

	key := "secret-key"
	headerName := "X-Metric-Hash"

	// Marshal metrics JSON
	jsonBytes, err := json.Marshal(metrics)
	assert.NoError(err)

	// Compute expected HMAC SHA256 hash of JSON
	h := hmac.New(sha256.New, []byte(key))
	_, err = h.Write(jsonBytes)
	assert.NoError(err)
	expectedHash := hex.EncodeToString(h.Sum(nil))

	// Create test HTTP server with handler that verifies hash header
	ts := httptest.NewServer(testServerHandler(t, headerName, expectedHash))
	defer ts.Close()

	// Handler that calls the real HTTP server using Resty client
	handler := func(ctx context.Context, urlPath string, body []byte, header string, hashSum string) error {
		req := client.R().
			SetContext(ctx).
			SetHeader("Content-Type", "application/json").
			SetBody(body)
		if header != "" && hashSum != "" {
			req.SetHeader(header, hashSum)
		}

		// Prepend test server URL if urlPath is a path (starts with "/")
		var fullURL string
		if len(urlPath) > 0 && urlPath[0] == '/' {
			fullURL = ts.URL + urlPath
		} else {
			fullURL = urlPath
		}

		resp, err := req.Post(fullURL)
		if err != nil {
			return err
		}
		defer func() {
			if resp.RawResponse != nil && resp.RawResponse.Body != nil {
				resp.RawResponse.Body.Close()
			}
		}()

		if resp.StatusCode() >= 400 {
			return fmt.Errorf("server returned status %d", resp.StatusCode())
		}
		return nil
	}

	ctx := context.Background()

	// Pass only the path to avoid doubling
	err = updates(ctx, handler, "/updates/", headerName, key, "", metrics)
	assert.NoError(err)
}

func TestWorkerMetricsUpdate_BatchingAndResults(t *testing.T) {
	assert := assert.New(t)

	floatPtr := func(f float64) *float64 { return &f }

	// Prepare sample metrics (3 metrics)
	metrics := []types.Metrics{
		{ID: "m1", Type: types.Gauge, Value: floatPtr(1.0)},
		{ID: "m2", Type: types.Gauge, Value: floatPtr(2.0)},
		{ID: "m3", Type: types.Gauge, Value: floatPtr(3.0)},
	}

	// Create a test HTTP server that validates incoming requests
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal("/updates/", r.URL.Path)

		body, err := io.ReadAll(r.Body)
		defer r.Body.Close()
		assert.NoError(err)

		var received []types.Metrics
		err = json.Unmarshal(body, &received)
		assert.NoError(err)
		assert.NotEmpty(received)

		// Ensure batch size does not exceed 2
		assert.LessOrEqual(len(received), 2)

		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	// Handler to send requests to the test server
	handler := func(ctx context.Context, urlPath string, body []byte, headerName string, hashSum string) error {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, ts.URL+urlPath, bytes.NewReader(body))
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/json")
		if headerName != "" && hashSum != "" {
			req.Header.Set(headerName, hashSum)
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 400 {
			return fmt.Errorf("server returned status %d", resp.StatusCode)
		}
		return nil
	}

	jobs := make(chan types.Metrics)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	batchSize := 2

	// Make sure the arguments match your workerMetricsUpdate signature exactly (8 params)
	resultsCh := workerMetricsUpdate(ctx, handler, "/updates/", "X-Header", "", "", jobs, batchSize)

	// Send metrics asynchronously
	go func() {
		for _, m := range metrics {
			jobs <- m
		}
		close(jobs)
	}()

	var results []result
	timeout := time.After(2 * time.Second)

collectLoop:
	for {
		select {
		case res, ok := <-resultsCh:
			if !ok {
				break collectLoop
			}
			results = append(results, res)
		case <-timeout:
			t.Fatal("Timeout waiting for results")
		}
	}

	// Expect two batches due to batch size 2 and 3 total metrics
	assert.Len(results, 2)

	// First batch should have 2 metrics
	assert.Len(results[0].Data, 2)
	assert.NoError(results[0].Err)

	// Second batch should have 1 metric
	assert.Len(results[1].Data, 1)
	assert.NoError(results[1].Err)
}
func TestWorkerMetricsUpdate_ContextCancellation(t *testing.T) {
	assert := assert.New(t)

	// Create a dummy HTTP server (similar to the first test)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	handler := func(ctx context.Context, urlPath string, body []byte, headerName string, hashSum string) error {
		return nil
	}

	jobs := make(chan types.Metrics)
	ctx, cancel := context.WithCancel(context.Background())
	batchSize := 2

	// Pass the jobs channel as a receive-only channel
	resultsCh := workerMetricsUpdate(ctx, handler, ts.URL, "X-Header", "", "", (<-chan types.Metrics)(jobs), batchSize)

	// Cancel the context immediately
	cancel()

	// The results channel should close shortly after cancel
	select {
	case _, ok := <-resultsCh:
		assert.False(ok, "results channel should be closed after context cancel")
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for results channel to close after context cancellation")
	}
}

func TestStartMetricsReporting(t *testing.T) {
	assert := assert.New(t)

	floatPtr := func(f float64) *float64 { return &f }

	metrics := []types.Metrics{
		{ID: "m1", Type: "gauge", Value: floatPtr(1.0)},
		{ID: "m2", Type: "gauge", Value: floatPtr(2.0)},
		{ID: "m3", Type: "gauge", Value: floatPtr(3.0)},
	}

	var mu sync.Mutex
	var batches [][]types.Metrics

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal("/updates/", r.URL.Path)

		defer r.Body.Close()
		body, err := io.ReadAll(r.Body)
		assert.NoError(err)

		var received []types.Metrics
		err = json.Unmarshal(body, &received)
		assert.NoError(err)
		assert.NotEmpty(received)

		mu.Lock()
		batches = append(batches, received)
		mu.Unlock()

		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	handler := func(ctx context.Context, urlPath string, body []byte, headerName string, hashSum string) error {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, ts.URL+urlPath, bytes.NewReader(body))
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/json")
		if headerName != "" && hashSum != "" {
			req.Header.Set(headerName, hashSum)
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 400 {
			return fmt.Errorf("server returned status %d", resp.StatusCode)
		}
		return nil
	}

	jobs := make(chan types.Metrics, len(metrics))
	for _, m := range metrics {
		jobs <- m
	}
	close(jobs)

	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()

	reportInterval := 1
	batchSize := 2
	rateLimit := 2

	resultsCh := startMetricsReporting(
		ctx,
		handler,
		reportInterval,
		ts.URL,
		"X-Header",
		"",
		"",
		jobs,
		batchSize,
		rateLimit,
	)

	var results []result
collectLoop:
	for {
		select {
		case res, ok := <-resultsCh:
			if !ok {
				break collectLoop
			}
			results = append(results, res)
		case <-ctx.Done():
			break collectLoop
		}
	}

	mu.Lock()
	defer mu.Unlock()

	assert.GreaterOrEqual(len(batches), 1, "should have sent at least one batch")
	for _, batch := range batches {
		assert.LessOrEqual(len(batch), batchSize, "batch size should not exceed limit")
	}

	assert.GreaterOrEqual(len(results), 1)
	for _, res := range results {
		assert.LessOrEqual(len(res.Data), batchSize)
		assert.NoError(res.Err)
	}
}

func TestStartMetricsPolling(t *testing.T) {
	assert := assert.New(t)

	// Use a short poll interval so test runs fast
	pollInterval := 1 // seconds

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	metricsCh := startMetricsPolling(ctx, pollInterval)

	var metricsReceived []types.Metrics

collectLoop:
	for {
		select {
		case metric, ok := <-metricsCh:
			if !ok {
				break collectLoop
			}
			metricsReceived = append(metricsReceived, metric)
		case <-ctx.Done():
			break collectLoop
		}
	}

	assert.NotEmpty(metricsReceived, "expected to receive some metrics")

	// Optionally, do some deeper checks, e.g. metric types or values
	for _, m := range metricsReceived {
		assert.NotEmpty(m.ID, "metric ID should not be empty")
		// you can add more asserts based on your metric structure
	}
}

func TestStartMetricAgent(t *testing.T) {
	assert := assert.New(t)

	// Start a test HTTP server that accepts POST requests and returns 200 OK
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal("/updates/", r.URL.Path)
		body, err := io.ReadAll(r.Body)
		assert.NoError(err)
		defer r.Body.Close()

		// Optionally check body is not empty
		assert.NotEmpty(body)

		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	// Use a context with timeout so the test doesn't hang indefinitely
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := StartMetricAgent(ctx, ts.URL, "X-Test-Header", "test-key", "", 1, 1, 2, 1)

	// We expect either nil or context cancellation error because StartMetricAgent runs until ctx is done
	assert.True(err == nil || errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled),
		"expected nil or context cancellation error, got %v", err)
}
