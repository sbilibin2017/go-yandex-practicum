package runners

import (
	"context"
)

// Server описывает интерфейс сервера с возможностью запуска и корректного завершения работы.
type Server interface {
	// ListenAndServe запускает сервер и блокирует выполнение до завершения.
	// Возвращает ошибку, если сервер завершился с ошибкой.
	ListenAndServe() error

	// Shutdown выполняет корректное завершение работы сервера с учётом контекста ctx.
	// Позволяет завершить работу с таймаутом или отменой.
	Shutdown(ctx context.Context) error
}

// RunServer запускает сервер srv и контролирует его жизненный цикл в соответствии с ctx.
// Запускает srv.ListenAndServe() в отдельной горутине.
// Если ctx отменяется, вызывает srv.Shutdown для корректного завершения.
// Возвращает ошибку, возникшую при работе сервера или при его завершении.
func RunServer(ctx context.Context, srv Server) error {
	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.ListenAndServe()
	}()
	select {
	case <-ctx.Done():
		shutdownErr := srv.Shutdown(context.Background())
		if shutdownErr != nil {
			return shutdownErr
		}
		return ctx.Err()
	case err := <-errCh:
		return err
	}
}
