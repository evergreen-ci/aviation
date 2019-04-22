package aviation

import (
	"context"
	"time"

	"github.com/jpillora/backoff"
	"github.com/mongodb/grip"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

func MakeRetryUnaryClientInterceptor(maxRetries int) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		var lastErr error
		b := &backoff.Backoff{
			// TODO: figure out the best options for this.
			Min:    100 * time.Millisecond,
			Max:    time.Second,
			Factor: 2,
		}
		callCtx, cancel := context.WithCancel(ctx)
		defer cancel()

	retry:
		for i := 0; i < maxRetries; i++ {
			lastErr = invoker(callCtx, method, req, reply, cc, opts...)
			grip.Infof("GRPC client retry attempt: %d, error: %v", i, lastErr)
			if lastErr == nil {
				return nil
			}

			if !isRetriable(lastErr) {
				break
			}

			timer := time.NewTimer(b.Duration())
			select {
			case <-ctx.Done():
				break retry
			case <-timer.C:
				continue retry
			}
		}

		grip.Warning("GRPC client retries exceeded or canceled!")
		return lastErr
	}
}

func MakeRetryStreamClientInterceptor(maxRetries int) grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		var clientStream grpc.ClientStream
		var lastErr error
		b := &backoff.Backoff{
			// TODO: figure out the best options for this.
			Min:    100 * time.Millisecond,
			Max:    time.Second,
			Factor: 2,
		}
		callCtx, cancel := context.WithCancel(ctx)
		defer cancel()

	retry:
		for i := 0; i < maxRetries; i++ {
			clientStream, lastErr = streamer(callCtx, desc, cc, method, opts...)
			grip.Infof("GRPC client retry attempt: %d, error: %v", i, lastErr)
			if lastErr == nil {
				return clientStream, nil
			}

			if !isRetriable(lastErr) {
				break
			}

			timer := time.NewTimer(b.Duration())
			select {
			case <-ctx.Done():
				break retry
			case <-timer.C:
				continue retry
			}
		}

		grip.Warning("GRPC client retries exceeded or canceled!")
		return nil, lastErr
	}
}

func isRetriable(err error) bool {
	switch errCode := grpc.Code(err); errCode {
	case codes.Unknown:
		return true
	case codes.ResourceExhausted:
		return true
	case codes.Aborted:
		return true
	case codes.Internal:
		return true
	case codes.Unavailable:
		return true
	default:
		return false
	}
}
