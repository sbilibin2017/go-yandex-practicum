package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func TestGetURLParam(t *testing.T) {
	req, err := http.NewRequest("GET", "/test/123", nil)
	if err != nil {
		t.Fatal(err)
	}
	r := chi.NewRouter()
	r.Get("/test/{id}", func(w http.ResponseWriter, r *http.Request) {
		param := getURLParam(r, "id")
		assert.Equal(t, "123", param)
	})
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
}
