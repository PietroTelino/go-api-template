package authentication

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"

	"github.com/PietroTelino/go-api-template/internal/auth"
	"github.com/PietroTelino/go-api-template/internal/i18n"
	"github.com/PietroTelino/go-api-template/internal/modules/audit"
	"github.com/PietroTelino/go-api-template/internal/response"
)

type Handler struct {
	service      *Service
	auditService *audit.Service
	translator   *i18n.Translator
}

func NewHandler(service *Service, auditService *audit.Service, translator *i18n.Translator) *Handler {
	return &Handler{service: service, auditService: auditService, translator: translator}
}

func isKnownError(err error) bool {
	return errors.Is(err, ErrInvalidCredentials) ||
		errors.Is(err, ErrAccountDeleted) ||
		errors.Is(err, ErrAccountInactive) ||
		errors.Is(err, ErrRefreshTokenInvalidOrExpired) ||
		errors.Is(err, ErrRefreshTokenMalformed) ||
		errors.Is(err, ErrRefreshTokenInvalid) ||
		errors.Is(err, ErrUserNotFound)
}

func (handler *Handler) writeServiceError(writer http.ResponseWriter, request *http.Request, err error) {
	if isKnownError(err) {
		response.WriteTranslatedError(writer, request, handler.translator, http.StatusUnauthorized, err.Error())
		return
	}

	log.Printf("Erro inesperado em %s %s: %v", request.Method, request.URL.Path, err)
	response.WriteTranslatedError(writer, request, handler.translator, http.StatusInternalServerError, "common.internalError")
}

func clientInfoFromRequest(request *http.Request) ClientInfo {
	ipAddress, userAgent := audit.ClientInfoFromRequest(request)
	return ClientInfo{IPAddress: ipAddress, UserAgent: userAgent}
}

func (handler *Handler) logSessionAction(request *http.Request, userID string, action audit.Action) {
	clientInfo := clientInfoFromRequest(request)

	handler.auditService.Log(request.Context(), audit.Entry{
		UserID:        &userID,
		PerformedByID: &userID,
		Action:        action,
		Entity:        audit.EntitySession,
		EntityID:      &userID,
		IPAddress:     clientInfo.IPAddress,
		UserAgent:     clientInfo.UserAgent,
	})
}

func (handler *Handler) Login(writer http.ResponseWriter, request *http.Request) {
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
		response.WriteTranslatedError(writer, request, handler.translator, http.StatusBadRequest, "common.invalidRequestBody")
		return
	}

	if body.Email == "" || body.Password == "" {
		response.WriteTranslatedError(writer, request, handler.translator, http.StatusBadRequest, "auth.emailPasswordRequired")
		return
	}

	clientInfo := clientInfoFromRequest(request)

	result, err := handler.service.Login(request.Context(), body.Email, body.Password, clientInfo)
	if err != nil {
		handler.writeServiceError(writer, request, err)
		return
	}

	handler.auditService.Log(request.Context(), audit.Entry{
		UserID:        &result.User.ID,
		PerformedByID: &result.User.ID,
		Action:        audit.ActionUserLogin,
		Entity:        audit.EntityUser,
		EntityID:      &result.User.ID,
		Details:       map[string]any{"email": result.User.Email},
		IPAddress:     clientInfo.IPAddress,
		UserAgent:     clientInfo.UserAgent,
	})

	response.WriteJSON(writer, http.StatusOK, result)
}

func (handler *Handler) Refresh(writer http.ResponseWriter, request *http.Request) {
	var body struct {
		RefreshToken string `json:"refreshToken"`
	}

	if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
		response.WriteTranslatedError(writer, request, handler.translator, http.StatusBadRequest, "common.invalidRequestBody")
		return
	}

	if body.RefreshToken == "" {
		response.WriteTranslatedError(writer, request, handler.translator, http.StatusBadRequest, "auth.refreshTokenRequired")
		return
	}

	result, err := handler.service.Refresh(request.Context(), body.RefreshToken, clientInfoFromRequest(request))
	if err != nil {
		handler.writeServiceError(writer, request, err)
		return
	}

	response.WriteJSON(writer, http.StatusOK, result)
}

func (handler *Handler) Logout(writer http.ResponseWriter, request *http.Request) {
	authenticatedUser, isAuthenticated := auth.AuthenticatedUserFromContext(request.Context())
	if !isAuthenticated {
		response.WriteTranslatedError(writer, request, handler.translator, http.StatusUnauthorized, "auth.notAuthenticated")
		return
	}

	var body struct {
		RefreshToken string `json:"refreshToken"`
	}

	if err := json.NewDecoder(request.Body).Decode(&body); err != nil && !errors.Is(err, io.EOF) {
		response.WriteTranslatedError(writer, request, handler.translator, http.StatusBadRequest, "common.invalidRequestBody")
		return
	}

	refreshToken := body.RefreshToken
	if refreshToken == "" {
		refreshToken = request.Header.Get("X-Refresh-Token")
	}

	if err := handler.service.Logout(request.Context(), refreshToken, authenticatedUser.ID); err != nil {
		handler.writeServiceError(writer, request, err)
		return
	}

	handler.logSessionAction(request, authenticatedUser.ID, audit.ActionUserLogout)

	writer.WriteHeader(http.StatusNoContent)
}

func (handler *Handler) LogoutAll(writer http.ResponseWriter, request *http.Request) {
	authenticatedUser, isAuthenticated := auth.AuthenticatedUserFromContext(request.Context())
	if !isAuthenticated {
		response.WriteTranslatedError(writer, request, handler.translator, http.StatusUnauthorized, "auth.notAuthenticated")
		return
	}

	if err := handler.service.LogoutAll(request.Context(), authenticatedUser.ID); err != nil {
		handler.writeServiceError(writer, request, err)
		return
	}

	handler.logSessionAction(request, authenticatedUser.ID, audit.ActionUserLogoutAll)

	writer.WriteHeader(http.StatusNoContent)
}
