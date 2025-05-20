package handlers

import (
	"net/http"

	"github.com/jmoiron/sqlx"
)

// NewDBPingHandler возвращает HTTP-обработчик, который проверяет доступность базы данных.
//
// Обработчик пингует базу данных с использованием контекста запроса.
// В случае успешного пинга возвращается статус 200 OK и тело "OK".
// Если возникает ошибка при пинге базы, возвращается 500 Internal Server Error с сообщением "Database connection error".
//
// Параметры:
//   - db: объект подключения к базе данных (*sqlx.DB).
//
// Возвращает:
//   - http.HandlerFunc — функцию-обработчик HTTP-запросов.
func NewDBPingHandler(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := db.PingContext(r.Context()); err != nil {
			http.Error(w, "Database connection error", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}
}
