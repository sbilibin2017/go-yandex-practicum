package retrier

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgconn"
)

func WithRetry(
	ctx context.Context,
	attempts []time.Duration,
	fn func(ctx context.Context) error,
	isRetriableErrorFuncs ...func(err error) bool,
) error {
	var lastErr error
	total := len(attempts)

	isRetriableErrorFunc := func(err error) bool {
		if len(isRetriableErrorFuncs) == 0 {
			return false
		}
		for _, f := range isRetriableErrorFuncs {
			if f(err) {
				return true
			}
		}
		return false
	}

	for i := 0; i <= total; i++ {

		err := fn(ctx)
		if err == nil || !isRetriableErrorFunc(err) {
			return err
		}

		lastErr = err

		if i == total {
			break
		}

		select {
		case <-time.After(attempts[i]):
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return lastErr
}

func IsRetriableDBError(err error) bool {
	if err == nil {
		return false
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && len(pgErr.Code) >= 2 && pgErr.Code[:2] == "08" {
		return true
	}
	return false
}
