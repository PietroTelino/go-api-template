package sessions

import (
	"errors"
	"log"
	"net/http"

	"github.com/PietroTelino/go-api-template/internal/auth"
	"github.com/PietroTelino/go-api-template/internal/i18n"
	"github.com/PietroTelino/go-api-template/internal/modules/audit"
	"github.com/PietroTelino/go-api-template/internal/response"
	"github.com/go-chi/chi/v5"
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

func (handler *Handler) writeUnexpectedError(writer http.ResponseWriter, request *http.Request, err error, fallbackKey string) {
	log.Printf("Erro inesperado em %s %s: %v", request.Method, request.URL.Path, err)
	response.WriteTranslatedError(writer, request, handler.translator, http.StatusInternalServerError, fallbackKey)
}

func (handler *Handler) ListMySessions(writer http.ResponseWriter, request *http.Request) {
	authenticatedUser, isAuthenticated := auth.AuthenticatedUserFromContext(request.Context())
	if !isAuthenticated {
		response.WriteTranslatedError(writer, request, handler.translator, http.StatusUnauthorized, "auth.notAuthenticated")
		return
	}

	activeSessions, err := handler.service.ListMySessions(request.Context(), authenticatedUser.ID)
	if err != nil {
		handler.writeUnexpectedError(writer, request, err, "user.errorListSessions")
		return
	}

	response.WriteJSON(writer, http.StatusOK, activeSessions)
}

func (handler *Handler) TerminateSession(writer http.ResponseWriter, request *http.Request) {
	authenticatedUser, isAuthenticated := auth.AuthenticatedUserFromContext(request.Context())
	if !isAuthenticated {
		response.WriteTranslatedError(writer, request, handler.translator, http.StatusUnauthorized, "auth.notAuthenticated")
		return
	}

	sessionID := chi.URLParam(request, "sessionId")

	err := handler.service.TerminateSession(request.Context(), authenticatedUser.ID, sessionID)
	if errors.Is(err, ErrSessionNotFound) {
		response.WriteTranslatedError(writer, request, handler.translator, http.StatusNotFound, err.Error())
		return
	}

	if err != nil {
		handler.writeUnexpectedError(writer, request, err, "user.errorTerminateSession")
		return
	}

	ipAddress, userAgent := audit.ClientInfoFromRequest(request)

	handler.auditService.Log(request.Context(), audit.Entry{
		UserID:        &authenticatedUser.ID,
		PerformedByID: &authenticatedUser.ID,
		Action:        audit.ActionSessionLogout,
		Entity:        audit.EntitySession,
		EntityID:      &authenticatedUser.ID,
		Details:       map[string]any{"sessionId": sessionID},
		IPAddress:     ipAddress,
		UserAgent:     userAgent,
	})

	language := i18n.LanguageFromContext(request.Context())
	response.WriteJSON(writer, http.StatusOK, messageResponse{Message: handler.translator.Translate(language, "user.sessionTerminated")})
}
