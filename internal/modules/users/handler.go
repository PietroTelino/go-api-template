package users

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

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

type registerResponse struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"createdAt"`
}

type messageResponse struct {
	Message string `json:"message"`
}

func NewHandler(service *Service, auditService *audit.Service, translator *i18n.Translator) *Handler {
	return &Handler{service: service, auditService: auditService, translator: translator}
}

var errorStatusCodes = map[error]int{
	ErrEmailAlreadyInUse:     http.StatusConflict,
	ErrEmailRequired:         http.StatusBadRequest,
	ErrEmailInvalidFormat:    http.StatusBadRequest,
	ErrPasswordTooShort:      http.StatusBadRequest,
	ErrPasswordTooLong:       http.StatusBadRequest,
	ErrPasswordNoSpecialChar: http.StatusBadRequest,
	ErrUserNotFound:          http.StatusNotFound,
	ErrAlreadyDeleted:        http.StatusBadRequest,
	ErrNotDeleted:            http.StatusBadRequest,
	ErrAlreadyInactive:       http.StatusBadRequest,
	ErrAlreadyActive:         http.StatusBadRequest,
	ErrIncorrectPassword:     http.StatusBadRequest,
	ErrCurrentPasswordWrong:  http.StatusBadRequest,
	ErrInvalidTheme:          http.StatusBadRequest,
	ErrInvalidRole:           http.StatusBadRequest,
	ErrCannotAssignRole:      http.StatusForbidden,
	ErrCannotModifyGod:       http.StatusForbidden,
	ErrCannotInactivateGod:   http.StatusForbidden,
	ErrCannotDeleteGod:       http.StatusForbidden,
}

func statusCodeForError(err error) (int, bool) {
	for knownError, statusCode := range errorStatusCodes {
		if errors.Is(err, knownError) {
			return statusCode, true
		}
	}

	return 0, false
}

func (handler *Handler) writeServiceError(writer http.ResponseWriter, request *http.Request, err error, fallbackKey string) {
	if statusCode, isKnownError := statusCodeForError(err); isKnownError {
		response.WriteTranslatedError(writer, request, handler.translator, statusCode, err.Error())
		return
	}

	log.Printf("Erro inesperado em %s %s: %v", request.Method, request.URL.Path, err)
	response.WriteTranslatedError(writer, request, handler.translator, http.StatusInternalServerError, fallbackKey)
}

func (handler *Handler) writeMessage(writer http.ResponseWriter, request *http.Request, translationKey string) {
	language := i18n.LanguageFromContext(request.Context())
	response.WriteJSON(writer, http.StatusOK, messageResponse{Message: handler.translator.Translate(language, translationKey)})
}

func (handler *Handler) decodeBody(writer http.ResponseWriter, request *http.Request, body any) bool {
	if err := json.NewDecoder(request.Body).Decode(body); err != nil {
		response.WriteTranslatedError(writer, request, handler.translator, http.StatusBadRequest, "common.invalidRequestBody")
		return false
	}

	return true
}

func (handler *Handler) authenticatedUser(writer http.ResponseWriter, request *http.Request) (*auth.AuthenticatedUser, bool) {
	authenticatedUser, isAuthenticated := auth.AuthenticatedUserFromContext(request.Context())
	if !isAuthenticated {
		response.WriteTranslatedError(writer, request, handler.translator, http.StatusUnauthorized, "auth.notAuthenticated")
	}

	return authenticatedUser, isAuthenticated
}

func (handler *Handler) logAction(request *http.Request, userID *string, performedByID string, action audit.Action, entity audit.Entity, entityID string, details map[string]any) {
	ipAddress, userAgent := audit.ClientInfoFromRequest(request)

	handler.auditService.Log(request.Context(), audit.Entry{
		UserID:        userID,
		PerformedByID: &performedByID,
		Action:        action,
		Entity:        entity,
		EntityID:      &entityID,
		Details:       details,
		IPAddress:     ipAddress,
		UserAgent:     userAgent,
	})
}

func (handler *Handler) Register(writer http.ResponseWriter, request *http.Request) {
	var body struct {
		Name        string         `json:"name"`
		Email       string         `json:"email"`
		Password    string         `json:"password"`
		Preferences map[string]any `json:"preferences"`
	}

	if !handler.decodeBody(writer, request, &body) {
		return
	}

	if body.Name == "" || body.Email == "" || body.Password == "" {
		response.WriteTranslatedError(writer, request, handler.translator, http.StatusBadRequest, "user.fieldRequired")
		return
	}

	createdUser, err := handler.service.Register(request.Context(), RegisterInput{
		Name:        body.Name,
		Email:       body.Email,
		Password:    body.Password,
		Preferences: body.Preferences,
	})

	if err != nil {
		handler.writeServiceError(writer, request, err, "user.errorRegister")
		return
	}

	handler.logAction(request, &createdUser.ID, createdUser.ID, audit.ActionUserRegister, audit.EntityUser, createdUser.ID, map[string]any{"email": createdUser.Email})

	response.WriteJSON(writer, http.StatusCreated, registerResponse{
		ID:        createdUser.ID,
		Email:     createdUser.Email,
		CreatedAt: createdUser.CreatedAt,
	})
}

func (handler *Handler) List(writer http.ResponseWriter, request *http.Request) {
	foundUsers, err := handler.service.List(request.Context())
	if err != nil {
		handler.writeServiceError(writer, request, err, "user.errorList")
		return
	}

	response.WriteJSON(writer, http.StatusOK, foundUsers)
}

func (handler *Handler) ListDeleted(writer http.ResponseWriter, request *http.Request) {
	foundUsers, err := handler.service.ListDeleted(request.Context())
	if err != nil {
		handler.writeServiceError(writer, request, err, "user.errorListDeleted")
		return
	}

	response.WriteJSON(writer, http.StatusOK, foundUsers)
}

func (handler *Handler) GetByID(writer http.ResponseWriter, request *http.Request) {
	user, err := handler.service.GetByID(request.Context(), chi.URLParam(request, "id"))
	if err != nil {
		handler.writeServiceError(writer, request, err, "user.errorFind")
		return
	}

	response.WriteJSON(writer, http.StatusOK, user)
}

func (handler *Handler) Create(writer http.ResponseWriter, request *http.Request) {
	actor, isAuthenticated := handler.authenticatedUser(writer, request)
	if !isAuthenticated {
		return
	}

	var body struct {
		Name        string         `json:"name"`
		Email       string         `json:"email"`
		Password    string         `json:"password"`
		Role        string         `json:"role"`
		Preferences map[string]any `json:"preferences"`
	}

	if !handler.decodeBody(writer, request, &body) {
		return
	}

	if body.Name == "" || body.Email == "" || body.Password == "" {
		response.WriteTranslatedError(writer, request, handler.translator, http.StatusBadRequest, "user.fieldRequired")
		return
	}

	createdUser, err := handler.service.Create(request.Context(), actor.Role, CreateInput{
		Name:        body.Name,
		Email:       body.Email,
		Password:    body.Password,
		Role:        body.Role,
		Preferences: body.Preferences,
	})

	if err != nil {
		handler.writeServiceError(writer, request, err, "user.errorCreate")
		return
	}

	handler.logAction(request, &createdUser.ID, actor.ID, audit.ActionUserCreate, audit.EntityUser, createdUser.ID, map[string]any{"email": createdUser.Email, "role": createdUser.Role})

	response.WriteJSON(writer, http.StatusCreated, createdUser)
}

func (handler *Handler) Update(writer http.ResponseWriter, request *http.Request) {
	actor, isAuthenticated := handler.authenticatedUser(writer, request)
	if !isAuthenticated {
		return
	}

	var body struct {
		Name  *string `json:"name"`
		Email *string `json:"email"`
	}

	if !handler.decodeBody(writer, request, &body) {
		return
	}

	id := chi.URLParam(request, "id")

	updatedUser, err := handler.service.Update(request.Context(), id, UpdateInput{Name: body.Name, Email: body.Email})
	if err != nil {
		handler.writeServiceError(writer, request, err, "user.errorUpdate")
		return
	}

	updatedFields := []string{}
	if body.Name != nil {
		updatedFields = append(updatedFields, "name")
	}
	if body.Email != nil {
		updatedFields = append(updatedFields, "email")
	}

	handler.logAction(request, &updatedUser.ID, actor.ID, audit.ActionUserUpdate, audit.EntityUser, updatedUser.ID, map[string]any{"fields": updatedFields})

	response.WriteJSON(writer, http.StatusOK, updatedUser)
}

func (handler *Handler) ChangeMyPassword(writer http.ResponseWriter, request *http.Request) {
	actor, isAuthenticated := handler.authenticatedUser(writer, request)
	if !isAuthenticated {
		return
	}

	var body struct {
		CurrentPassword string `json:"currentPassword"`
		NewPassword     string `json:"newPassword"`
	}

	if !handler.decodeBody(writer, request, &body) {
		return
	}

	if body.CurrentPassword == "" || body.NewPassword == "" {
		response.WriteTranslatedError(writer, request, handler.translator, http.StatusBadRequest, "user.passwordsRequired")
		return
	}

	if err := handler.service.ChangeMyPassword(request.Context(), actor.ID, body.CurrentPassword, body.NewPassword); err != nil {
		handler.writeServiceError(writer, request, err, "user.errorChangePassword")
		return
	}

	handler.logAction(request, &actor.ID, actor.ID, audit.ActionPasswordChange, audit.EntityPassword, actor.ID, map[string]any{"message": "Senha alterada pelo próprio usuário"})

	handler.writeMessage(writer, request, "user.passwordChanged")
}

func (handler *Handler) UpdateMyPreferences(writer http.ResponseWriter, request *http.Request) {
	actor, isAuthenticated := handler.authenticatedUser(writer, request)
	if !isAuthenticated {
		return
	}

	var body struct {
		Theme string `json:"theme"`
	}

	if !handler.decodeBody(writer, request, &body) {
		return
	}

	if body.Theme == "" {
		response.WriteTranslatedError(writer, request, handler.translator, http.StatusBadRequest, "user.themeRequired")
		return
	}

	if err := handler.service.UpdateMyTheme(request.Context(), actor.ID, body.Theme); err != nil {
		handler.writeServiceError(writer, request, err, "user.errorPreferences")
		return
	}

	preferences := map[string]any{"theme": body.Theme}

	handler.logAction(request, &actor.ID, actor.ID, audit.ActionPreferencesUpdate, audit.EntityPreferences, actor.ID, map[string]any{"preferences": preferences})

	response.WriteJSON(writer, http.StatusOK, map[string]any{"preferences": preferences})
}

func (handler *Handler) SelfDelete(writer http.ResponseWriter, request *http.Request) {
	actor, isAuthenticated := handler.authenticatedUser(writer, request)
	if !isAuthenticated {
		return
	}

	var body struct {
		Password string `json:"password"`
	}

	if !handler.decodeBody(writer, request, &body) {
		return
	}

	if body.Password == "" {
		response.WriteTranslatedError(writer, request, handler.translator, http.StatusBadRequest, "user.passwordRequired")
		return
	}

	deletedUser, err := handler.service.SelfDelete(request.Context(), actor.ID, body.Password)
	if err != nil {
		handler.writeServiceError(writer, request, err, "user.errorSelfDelete")
		return
	}

	handler.logAction(request, &actor.ID, actor.ID, audit.ActionUserSelfDelete, audit.EntityUser, actor.ID, map[string]any{"email": deletedUser.Email})

	handler.writeMessage(writer, request, "user.accountRemoved")
}

func (handler *Handler) Delete(writer http.ResponseWriter, request *http.Request) {
	handler.runAdminAction(writer, request, handler.service.Delete, audit.ActionUserDelete, "user.errorDelete", "user.accountRemoved")
}

func (handler *Handler) Restore(writer http.ResponseWriter, request *http.Request) {
	handler.runAdminAction(writer, request, handler.service.Restore, audit.ActionUserRestore, "user.errorRestore", "user.restored")
}

func (handler *Handler) Inactivate(writer http.ResponseWriter, request *http.Request) {
	handler.runAdminAction(writer, request, handler.service.Inactivate, audit.ActionUserInactivate, "user.errorInactivate", "user.inactivated")
}

func (handler *Handler) Reactivate(writer http.ResponseWriter, request *http.Request) {
	handler.runAdminAction(writer, request, handler.service.Reactivate, audit.ActionUserReactivate, "user.errorReactivate", "user.reactivated")
}

type adminAction func(ctx context.Context, id string) (*User, error)

func (handler *Handler) runAdminAction(
	writer http.ResponseWriter,
	request *http.Request,
	action adminAction,
	auditAction audit.Action,
	errorKey string,
	successKey string,
) {
	actor, isAuthenticated := handler.authenticatedUser(writer, request)
	if !isAuthenticated {
		return
	}

	targetUser, err := action(request.Context(), chi.URLParam(request, "id"))
	if err != nil {
		handler.writeServiceError(writer, request, err, errorKey)
		return
	}

	handler.logAction(request, &targetUser.ID, actor.ID, auditAction, audit.EntityUser, targetUser.ID, map[string]any{"email": targetUser.Email})

	handler.writeMessage(writer, request, successKey)
}

func (handler *Handler) HardDelete(writer http.ResponseWriter, request *http.Request) {
	actor, isAuthenticated := handler.authenticatedUser(writer, request)
	if !isAuthenticated {
		return
	}

	deletedUser, err := handler.service.HardDelete(request.Context(), chi.URLParam(request, "id"))
	if err != nil {
		handler.writeServiceError(writer, request, err, "user.errorHardDelete")
		return
	}

	handler.logAction(request, nil, actor.ID, audit.ActionUserHardDelete, audit.EntityUser, deletedUser.ID, map[string]any{"email": deletedUser.Email})

	writer.WriteHeader(http.StatusNoContent)
}
