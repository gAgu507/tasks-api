package auth

import "context"

type ctxKey struct{}

func WithUserID(ctx context.Context, id int) context.Context {
	return context.WithValue(ctx, ctxKey{}, id)
}

func UserID(ctx context.Context) (int, bool) {
	id, ok := ctx.Value(ctxKey{}).(int)
	return id, ok
}
