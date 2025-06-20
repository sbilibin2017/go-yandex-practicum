package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

func TestDBPingHandler(t *testing.T) {
	t.Run("nil db returns 503", func(t *testing.T) {
		handler := DBPingHandler(nil)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()

		handler(w, req)
		resp := w.Result()

		assert.Equal(t, http.StatusServiceUnavailable, resp.StatusCode)
	})

	t.Run("db ping error returns 500", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
		assert.NoError(t, err)
		defer db.Close()

		mock.ExpectPing().WillReturnError(assert.AnError)

		sqlxDB := sqlx.NewDb(db, "sqlmock")

		handler := DBPingHandler(sqlxDB)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()

		handler(w, req)
		resp := w.Result()

		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("successful ping returns 200", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
		assert.NoError(t, err)
		defer db.Close()

		mock.ExpectPing().WillReturnError(nil)

		sqlxDB := sqlx.NewDb(db, "sqlmock")

		handler := DBPingHandler(sqlxDB)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()

		handler(w, req)
		resp := w.Result()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
