package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func TestParseURLParam_Success(t *testing.T) {
	req, err := http.NewRequest("GET", "/update/John/type1/100", nil)
	if err != nil {
		t.Fatal(err)
	}
	r := chi.NewRouter()
	r.Get("/update/{name}/{type}/{value}", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Name  string `urlparam:"name"`
			Type  string `urlparam:"type"`
			Value string `urlparam:"value"`
		}
		parseURLParam(r, &req)

		assert.Equal(t, "John", req.Name)
		assert.Equal(t, "type1", req.Type)
		assert.Equal(t, "100", req.Value)
		w.Write([]byte("OK"))
	})
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestParseURLParam_EmptyParam(t *testing.T) {
	req, err := http.NewRequest("GET", "/update//type1/100", nil)
	if err != nil {
		t.Fatal(err)
	}
	r := chi.NewRouter()
	r.Get("/update/{name}/{type}/{value}", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Name  string `urlparam:"name"`
			Type  string `urlparam:"type"`
			Value string `urlparam:"value"`
		}
		parseURLParam(r, &req)

		assert.Empty(t, req.Name)
		assert.Equal(t, "type1", req.Type)
		assert.Equal(t, "100", req.Value)
		w.Write([]byte("OK"))
	})
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestParseURLParam_SkipFieldWithoutURLParamTag(t *testing.T) {
	type TestRequest struct {
		Name    string `urlparam:"name"`
		Age     string
		Country string `urlparam:"country"`
	}
	req, err := http.NewRequest("GET", "/update/John//USA", nil)
	if err != nil {
		t.Fatal(err)
	}
	r := chi.NewRouter()
	r.Get("/update/{name}/{age}/{country}", func(w http.ResponseWriter, r *http.Request) {
		var request TestRequest
		parseURLParam(r, &request)
		assert.Equal(t, "", request.Age)
		assert.Equal(t, "John", request.Name)
		assert.Equal(t, "USA", request.Country)

		w.Write([]byte("OK"))
	})
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
}
