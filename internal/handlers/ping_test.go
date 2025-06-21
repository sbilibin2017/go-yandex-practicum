package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupMockDB(t *testing.T) (*sqlx.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	require.NoError(t, err)

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	return sqlxDB, mock
}

func TestPingHandler_ServeHTTP_Success(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	mock.ExpectPing()

	handler := NewPingDBHandler(WithPingDB(db))

	req := httptest.NewRequest("GET", "/ping", nil)
	w := httptest.NewRecorder()

	handler.serveHTTP(w, req)

	resp := w.Result()
	defer resp.Body.Close() // <--- Add this line

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "text/plain", resp.Header.Get("Content-Type"))

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPingHandler_ServeHTTP_NilDB(t *testing.T) {
	handler := NewPingDBHandler() // no DB configured

	req := httptest.NewRequest("GET", "/ping", nil)
	w := httptest.NewRecorder()

	handler.serveHTTP(w, req)

	resp := w.Result()
	defer resp.Body.Close() // <--- Add this line

	assert.Equal(t, http.StatusServiceUnavailable, resp.StatusCode)
}

func TestPingHandler_ServeHTTP_PingError(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	mock.ExpectPing().WillReturnError(assert.AnError)

	handler := NewPingDBHandler(WithPingDB(db))

	req := httptest.NewRequest("GET", "/ping", nil)
	w := httptest.NewRecorder()

	handler.serveHTTP(w, req)

	resp := w.Result()
	defer resp.Body.Close() // <--- Add this line

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPingHandler_RegisterRoute(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	mock.ExpectPing()

	handler := NewPingDBHandler(WithPingDB(db))

	r := chi.NewRouter()
	handler.RegisterRoute(r)

	req := httptest.NewRequest("GET", "/ping", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	resp := w.Result()
	defer resp.Body.Close() // <--- Add this line

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	assert.NoError(t, mock.ExpectationsWereMet())
}
