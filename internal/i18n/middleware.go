package i18n

import (
	"net/http"
	"strings"
)

func Middleware(translator *Translator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
			resolvedLanguage := resolveLanguageFromHeader(
				request.Header.Get("Accept-Language"),
				translator.SupportedLanguages(),
				translator.FallbackLanguage(),
			)

			requestWithLanguage := request.WithContext(WithLanguage(request.Context(), resolvedLanguage))
			next.ServeHTTP(responseWriter, requestWithLanguage)
		})
	}
}

func resolveLanguageFromHeader(acceptLanguageHeader string, supportedLanguages []string, fallbackLanguage string) string {
	if acceptLanguageHeader == "" {
		return fallbackLanguage
	}

	preferredLanguages := strings.Split(acceptLanguageHeader, ",")

	for _, preferredLanguage := range preferredLanguages {
		languageTag := strings.TrimSpace(strings.SplitN(preferredLanguage, ";", 2)[0])

		for _, supportedLanguage := range supportedLanguages {
			if strings.EqualFold(languageTag, supportedLanguage) {
				return supportedLanguage
			}
		}

		for _, supportedLanguage := range supportedLanguages {
			if strings.HasPrefix(strings.ToLower(supportedLanguage), strings.ToLower(languageTag)) {
				return supportedLanguage
			}
		}
	}

	return fallbackLanguage
}
