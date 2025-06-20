package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewTrustedSubnetMiddleware(t *testing.T) {
	trustedCIDR := "192.168.1.0/24"
	middleware := TrustedSubnetMiddleware(&trustedCIDR)

	// Тестовый хендлер, который будет вызван, если middleware пропустит запрос
	okHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// helper функция для выполнения запроса с заданным X-Real-IP
	doRequest := func(ip string, mw func(http.Handler) http.Handler) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		if ip != "" {
			req.Header.Set("X-Real-IP", ip)
		}

		rr := httptest.NewRecorder()
		handler := mw(okHandler)
		handler.ServeHTTP(rr, req)

		return rr
	}

	t.Run("valid IP inside subnet", func(t *testing.T) {
		rr := doRequest("192.168.1.42", middleware)
		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("valid IP outside subnet", func(t *testing.T) {
		rr := doRequest("10.0.0.1", middleware)
		assert.Equal(t, http.StatusForbidden, rr.Code)
	})

	t.Run("missing X-Real-IP header", func(t *testing.T) {
		rr := doRequest("", middleware)
		assert.Equal(t, http.StatusForbidden, rr.Code)
	})

	t.Run("invalid IP format", func(t *testing.T) {
		rr := doRequest("not-an-ip", middleware)
		assert.Equal(t, http.StatusForbidden, rr.Code)
	})

	t.Run("nil trustedSubnet disables check and allows all requests", func(t *testing.T) {
		mw := TrustedSubnetMiddleware(nil)

		rr := doRequest("any-ip", mw)
		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("invalid CIDR disables check and allows all requests", func(t *testing.T) {
		badCIDR := "invalid-cidr"
		mw := TrustedSubnetMiddleware(&badCIDR)

		rr := doRequest("192.168.1.42", mw)
		assert.Equal(t, http.StatusOK, rr.Code)
	})
}
