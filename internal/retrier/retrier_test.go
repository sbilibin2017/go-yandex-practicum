package retrier

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgconn"
	"github.com/stretchr/testify/assert"
)

func TestWithRetry(t *testing.T) {
	retriableErr := errors.New("retriable error")
	nonRetriableErr := errors.New("non-retriable error")

	isRetriable := func(err error) bool {
		return err == retriableErr
	}

	tests := []struct {
		name              string
		attempts          []time.Duration
		fnResults         []error
		isRetriableFuncs  []func(err error) bool
		expectedErr       error
		expectedCallCount int
	}{
		{
			name:              "success on first try",
			attempts:          []time.Duration{time.Millisecond},
			fnResults:         []error{nil},
			isRetriableFuncs:  []func(err error) bool{isRetriable},
			expectedErr:       nil,
			expectedCallCount: 1,
		},
		{
			name:              "non retriable error stops retries",
			attempts:          []time.Duration{time.Millisecond, time.Millisecond},
			fnResults:         []error{nonRetriableErr},
			isRetriableFuncs:  []func(err error) bool{isRetriable},
			expectedErr:       nonRetriableErr,
			expectedCallCount: 1,
		},
		{
			name:              "retries and eventually succeeds",
			attempts:          []time.Duration{time.Millisecond, time.Millisecond},
			fnResults:         []error{retriableErr, retriableErr, nil},
			isRetriableFuncs:  []func(err error) bool{isRetriable},
			expectedErr:       nil,
			expectedCallCount: 3,
		},
		{
			name:              "all retries fail with retriable error",
			attempts:          []time.Duration{time.Millisecond, time.Millisecond},
			fnResults:         []error{retriableErr, retriableErr, retriableErr},
			isRetriableFuncs:  []func(err error) bool{isRetriable},
			expectedErr:       retriableErr,
			expectedCallCount: 3,
		},
		{
			name:              "no retriable funcs disables retry",
			attempts:          []time.Duration{time.Millisecond, time.Millisecond},
			fnResults:         []error{retriableErr, retriableErr},
			isRetriableFuncs:  nil,
			expectedErr:       retriableErr,
			expectedCallCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			callCount := 0

			ctx := context.Background()

			fn := func(ctx context.Context) error {
				if callCount >= len(tt.fnResults) {
					return nil
				}
				err := tt.fnResults[callCount]
				callCount++
				return err
			}

			err := WithRetry(ctx, tt.attempts, fn, tt.isRetriableFuncs...)

			assert.Equal(t, tt.expectedErr, err)
			assert.Equal(t, tt.expectedCallCount, callCount)
		})
	}

	t.Run("context canceled before retry delay", func(t *testing.T) {
		callCount := 0
		ctx, cancel := context.WithCancel(context.Background())

		fn := func(ctx context.Context) error {
			callCount++
			return retriableErr
		}

		isRetriable := func(err error) bool {
			return err == retriableErr
		}

		// Запускаем отмену контекста через 10 мс — в момент ожидания задержки
		go func() {
			time.Sleep(10 * time.Millisecond)
			cancel()
		}()

		// Ставим задержку больше 10мс, чтобы точно было ожидание
		err := WithRetry(ctx, []time.Duration{50 * time.Millisecond, 50 * time.Millisecond}, fn, isRetriable)

		// Ожидаем именно context.Canceled, а вызов fn должен быть 1 раз,
		// потому что отмена происходит в ожидании retry, не в fn
		assert.Equal(t, context.Canceled, err)
		assert.Equal(t, 1, callCount)
	})
}

func TestIsRetriableDBError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
		{
			name: "non-PgError error",
			err:  errors.New("some error"),
			want: false,
		},
		{
			name: "PgError with code 08xxx",
			err: &pgconn.PgError{
				Code: "08003", // connection_does_not_exist
			},
			want: true,
		},
		{
			name: "PgError with code not starting with 08",
			err: &pgconn.PgError{
				Code: "23505", // unique_violation
			},
			want: false,
		},
		{
			name: "PgError with short code",
			err: &pgconn.PgError{
				Code: "0",
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsRetriableDBError(tt.err)
			assert.Equal(t, tt.want, got)
		})
	}
}
