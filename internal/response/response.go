package response

import (
	"encoding/json"
	"net/http"

	"github.com/PietroTelino/go-api-template/internal/i18n"
)

func WriteJSON(responseWriter http.ResponseWriter, statusCode int, payload any) {
	responseWriter.Header().Set("Content-Type", "application/json")
	responseWriter.WriteHeader(statusCode)
	json.NewEncoder(responseWriter).Encode(payload)
}

func WriteTranslatedError(
	responseWriter http.ResponseWriter,
	request *http.Request,
	translator *i18n.Translator,
	statusCode int,
	translationKey string,
) {
	resolvedLanguage := i18n.LanguageFromContext(request.Context())
	translatedMessage := translator.Translate(resolvedLanguage, translationKey)

	WriteJSON(responseWriter, statusCode, map[string]string{"message": translatedMessage})
}
