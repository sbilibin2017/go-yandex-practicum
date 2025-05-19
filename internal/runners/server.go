package runners

import (
	"context"
)

type Server interface {
	ListenAndServe() error
	Shutdown(ctx context.Context) error
}

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
