package i18n

import "context"

type contextKey string

const resolvedLanguageContextKey contextKey = "resolvedLanguage"

func WithLanguage(ctx context.Context, resolvedLanguage string) context.Context {
	return context.WithValue(ctx, resolvedLanguageContextKey, resolvedLanguage)
}

func LanguageFromContext(ctx context.Context) string {
	resolvedLanguage, _ := ctx.Value(resolvedLanguageContextKey).(string)
	return resolvedLanguage
}
