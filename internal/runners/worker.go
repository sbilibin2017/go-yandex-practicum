package runners

import (
	"context"
)

// RunWorker запускает переданную функцию worker в отдельной горутине,
// контролируя её выполнение с помощью контекста ctx.
//
// Функция блокирует выполнение до тех пор, пока:
//   - worker завершится и вернёт ошибку (или nil),
//   - либо контекст ctx не будет отменён.
//
// Возвращает ошибку от worker или ошибку контекста (например, context.Canceled или context.DeadlineExceeded).
func RunWorker(ctx context.Context, worker func(ctx context.Context) error) error {
	errCh := make(chan error, 1)

	go func() {
		errCh <- worker(ctx)
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-errCh:
		return err
	}
}
