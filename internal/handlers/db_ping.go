package http

import (
	"net/http"

	"github.com/jmoiron/sqlx"
)

func DBPingHandler(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if db == nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}

		if err := db.PingContext(r.Context()); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
