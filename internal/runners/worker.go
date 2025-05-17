package runners

import (
	"context"
)

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
