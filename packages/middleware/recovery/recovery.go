package recovery

import (
	"context"
	"runtime"

	"p9e.in/ugcl/packages/errors"
	"p9e.in/ugcl/packages/middleware"
	"p9e.in/ugcl/packages/p9log"

	"google.golang.org/grpc"
)

// ErrUnknownRequest is unknown request error.
var ErrUnknownRequest = errors.InternalServer("UNKNOWN", "unknown request error")

// HandlerFunc is recovery handler func.
type HandlerFunc func(ctx context.Context, req, err interface{}) error

// Option is recovery option.
type Option func(*options)

type options struct {
	handler HandlerFunc
}

// WithHandler with recovery handler.
func WithHandler(h HandlerFunc) Option {
	return func(o *options) {
		o.handler = h
	}
}

// WithLogger with recovery logger.
// Deprecated: use global logger instead.
func WithLogger(logger p9log.Logger) Option {
	return func(o *options) {}
}

// Recovery is a server middleware that recovers from any panics.
func TestRecovery(opts ...Option) middleware.Middleware {
	op := options{
		handler: func(ctx context.Context, req, err interface{}) error {
			return ErrUnknownRequest
		},
	}
	for _, o := range opts {
		o(&op)
	}
	return func(handler middleware.Handler) middleware.Handler {

		return func(ctx context.Context, req interface{}) (reply interface{}, err error) {
			defer func() {
				if rerr := recover(); rerr != nil {
					buf := make([]byte, 64<<10) //nolint:gomnd
					n := runtime.Stack(buf, false)
					buf = buf[:n]
					p9log.Context(ctx).Errorf("%v: %+v\n%s\n", rerr, req, buf)

					err = op.handler(ctx, req, rerr)
				}
			}()
			return handler(ctx, req)
		}
	}
}

func Recovery() grpc.UnaryServerInterceptor {

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler) (reply interface{}, err error) {
		defer func() {
			if rerr := recover(); rerr != nil {
				buf := make([]byte, 64<<10) //nolint:gomnd
				n := runtime.Stack(buf, false)
				buf = buf[:n]
				p9log.Context(ctx).Errorf("%v: %+v\n%s\n", rerr, req, buf)

				// err = op.handler(ctx, req, rerr)
			}
		}()
		return handler(ctx, req)
	}
}
