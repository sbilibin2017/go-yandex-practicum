package middlewares

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHashMiddleware(t *testing.T) {
	const headerName = "X-Hash-Signature"
	const key = "test-secret-key"
	validBody := "test request body"
	validResponseBody := "response body"

	// Helper to generate HMAC hash
	makeHash := func(body []byte) string {
		mac := hmac.New(sha256.New, []byte(key))
		mac.Write(body)
		return hex.EncodeToString(mac.Sum(nil))
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(validResponseBody))
	})

	tests := []struct {
		name           string
		key            string
		header         string
		requestBody    string
		requestHash    string
		wantStatusCode int
		wantRespBody   string
		wantRespHeader string
	}{
		{
			name:           "No key, passes through",
			key:            "",
			header:         headerName,
			requestBody:    validBody,
			requestHash:    "",
			wantStatusCode: http.StatusOK,
			wantRespBody:   validResponseBody,
			wantRespHeader: "",
		},
		{
			name:           "Key set, no hash header, passes through",
			key:            key,
			header:         headerName,
			requestBody:    validBody,
			requestHash:    "",
			wantStatusCode: http.StatusOK,
			wantRespBody:   validResponseBody,
			wantRespHeader: "", // no incoming hash so no check; still sets response hash
		},
		{
			name:           "Key set, valid hash header",
			key:            key,
			header:         headerName,
			requestBody:    validBody,
			requestHash:    makeHash([]byte(validBody)),
			wantStatusCode: http.StatusOK,
			wantRespBody:   validResponseBody,
			wantRespHeader: "", // will be set in response, checked below
		},
		{
			name:           "Key set, invalid hash header",
			key:            key,
			header:         headerName,
			requestBody:    validBody,
			requestHash:    "invalidhash",
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   "",
			wantRespHeader: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware, err := HashMiddleware(
				WithHashKey(tt.key),
				WithHashHeader(tt.header),
			)
			require.NoError(t, err)

			h := middleware(handler)

			req := httptest.NewRequest(http.MethodPost, "http://example.com", strings.NewReader(tt.requestBody))
			if tt.requestHash != "" {
				req.Header.Set(tt.header, tt.requestHash)
			}
			w := httptest.NewRecorder()

			h.ServeHTTP(w, req)
			resp := w.Result()

			require.Equal(t, tt.wantStatusCode, resp.StatusCode)

			respBody, err := io.ReadAll(resp.Body)
			require.NoError(t, err)
			defer resp.Body.Close()

			if tt.wantStatusCode == http.StatusOK {
				require.Equal(t, tt.wantRespBody, string(respBody))

				if tt.key != "" {
					// Check response hash header is set correctly
					respHash := resp.Header.Get(tt.header)
					require.NotEmpty(t, respHash)

					// Validate hash value matches response body
					mac := hmac.New(sha256.New, []byte(tt.key))
					mac.Write(respBody)
					expectedHash := hex.EncodeToString(mac.Sum(nil))
					require.Equal(t, expectedHash, respHash)
				}
			} else {
				require.Empty(t, respBody)
			}
		})
	}
}
