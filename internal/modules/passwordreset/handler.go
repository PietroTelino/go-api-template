package passwordreset

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/PietroTelino/go-api-template/internal/i18n"
	"github.com/PietroTelino/go-api-template/internal/modules/audit"
	"github.com/PietroTelino/go-api-template/internal/modules/users"
	"github.com/PietroTelino/go-api-template/internal/response"
)

type Handler struct {
	service      *Service
	auditService *audit.Service
	translator   *i18n.Translator
}

type messageResponse struct {
	Message string `json:"message"`
}

func NewHandler(service *Service, auditService *audit.Service, translator *i18n.Translator) *Handler {
	return &Handler{service: service, auditService: auditService, translator: translator}
}

var knownErrors = []error{
	ErrInvalidOrExpiredToken,
	users.ErrPasswordTooShort,
	users.ErrPasswordTooLong,
	users.ErrPasswordNoSpecialChar,
}

func isKnownError(err error) bool {
	for _, knownError := range knownErrors {
		if errors.Is(err, knownError) {
			return true
		}
	}

	return false
}

func (handler *Handler) writeMessage(writer http.ResponseWriter, request *http.Request, translationKey string) {
	language := i18n.LanguageFromContext(request.Context())
	response.WriteJSON(writer, http.StatusOK, messageResponse{Message: handler.translator.Translate(language, translationKey)})
}

func (handler *Handler) Request(writer http.ResponseWriter, request *http.Request) {
	var body struct {
		Email string `json:"email"`
	}

	if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
		response.WriteTranslatedError(writer, request, handler.translator, http.StatusBadRequest, "common.invalidRequestBody")
		return
	}

	if body.Email == "" {
		response.WriteTranslatedError(writer, request, handler.translator, http.StatusBadRequest, "password.emailRequired")
		return
	}

	if err := handler.service.RequestReset(request.Context(), body.Email); err != nil {
		log.Printf("Erro inesperado em %s %s: %v", request.Method, request.URL.Path, err)
		response.WriteTranslatedError(writer, request, handler.translator, http.StatusInternalServerError, "password.errorRequest")
		return
	}

	ipAddress, userAgent := audit.ClientInfoFromRequest(request)

	handler.auditService.Log(request.Context(), audit.Entry{
		Action:    audit.ActionPasswordResetRequest,
		Entity:    audit.EntityPassword,
		Details:   map[string]any{"email": users.NormalizeEmail(body.Email)},
		IPAddress: ipAddress,
		UserAgent: userAgent,
	})

	handler.writeMessage(writer, request, "password.resetEmailSent")
}

func (handler *Handler) Confirm(writer http.ResponseWriter, request *http.Request) {
	var body struct {
		Token    string `json:"token"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
		response.WriteTranslatedError(writer, request, handler.translator, http.StatusBadRequest, "common.invalidRequestBody")
		return
	}

	if body.Token == "" || body.Password == "" {
		response.WriteTranslatedError(writer, request, handler.translator, http.StatusBadRequest, "password.tokenPasswordRequired")
		return
	}

	result, err := handler.service.ConfirmReset(request.Context(), body.Token, body.Password)
	if err != nil {
		if isKnownError(err) {
			response.WriteTranslatedError(writer, request, handler.translator, http.StatusBadRequest, err.Error())
			return
		}

		log.Printf("Erro inesperado em %s %s: %v", request.Method, request.URL.Path, err)
		response.WriteTranslatedError(writer, request, handler.translator, http.StatusInternalServerError, "common.internalError")
		return
	}

	ipAddress, userAgent := audit.ClientInfoFromRequest(request)

	handler.auditService.Log(request.Context(), audit.Entry{
		UserID:        &result.UserID,
		PerformedByID: &result.UserID,
		Action:        audit.ActionPasswordResetConfirm,
		Entity:        audit.EntityPassword,
		EntityID:      &result.UserID,
		Details:       map[string]any{"email": result.Email},
		IPAddress:     ipAddress,
		UserAgent:     userAgent,
	})

	handler.writeMessage(writer, request, "password.resetSuccess")
}
