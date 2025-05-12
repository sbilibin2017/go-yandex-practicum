package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/sbilibin2017/go-yandex-practicum/internal/logger"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestLoggingMiddleware(t *testing.T) {
	core, observed := observer.New(zapcore.InfoLevel)
	logger.Log = zap.New(core)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
		w.Write([]byte("I'm a teapot"))
	})
	middleware := LoggingMiddleware(handler)
	req := httptest.NewRequest(http.MethodGet, "/test-uri", nil)
	rec := httptest.NewRecorder()
	middleware.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusTeapot, rec.Code)
	assert.Equal(t, "I'm a teapot", rec.Body.String())
	logs := observed.All()
	assert.Len(t, logs, 2)
	assert.Equal(t, "Request info", logs[0].Message)
	fields := logs[0].ContextMap()
	assert.Equal(t, "/test-uri", fields["uri"])
	assert.Equal(t, "GET", fields["method"])
	_, ok := fields["duration"].(time.Duration)
	assert.True(t, ok, "duration should be a time.Duration")
	assert.Equal(t, "Response info", logs[1].Message)
	respFields := logs[1].ContextMap()
	assert.Equal(t, int64(http.StatusTeapot), respFields["status"])
	assert.Equal(t, int64(len("I'm a teapot")), respFields["size"])
}
