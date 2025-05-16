package runners

import (
	"context"
)

type Worker interface {
	Start(ctx context.Context) error
}

func RunWorker(ctx context.Context, w Worker) error {
	errCh := make(chan error, 1)
	go func() {
		errCh <- w.Start(ctx)
	}()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-errCh:
		return err
	}
}
