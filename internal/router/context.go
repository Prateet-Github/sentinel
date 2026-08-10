package router

import "context"

type contextKey struct{}

var paramsContextKey contextKey

func WithParams(ctx context.Context, params Params) context.Context {
	return context.WithValue(ctx, paramsContextKey, params)
}

func ParamsFromContext(ctx context.Context) (Params, bool) {
	params, ok := ctx.Value(paramsContextKey).(Params)
	return params, ok
}
