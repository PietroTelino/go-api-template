package users

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/PietroTelino/go-api-template/internal/i18n"
	"github.com/PietroTelino/go-api-template/internal/modules/audit"
	"github.com/PietroTelino/go-api-template/internal/response"
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

func NewHandler(service *Service, auditService *audit.Service, translator *i18n.Translator) *Handler {
	return &Handler{service: service, auditService: auditService, translator: translator}
}

func statusCodeForError(err error) (int, bool) {
	switch {
	case errors.Is(err, ErrEmailAlreadyInUse):
		return http.StatusConflict, true
	case errors.Is(err, ErrEmailRequired),
		errors.Is(err, ErrEmailInvalidFormat),
		errors.Is(err, ErrPasswordTooShort),
		errors.Is(err, ErrPasswordTooLong),
		errors.Is(err, ErrPasswordNoSpecialChar):
		return http.StatusBadRequest, true
	default:
		return 0, false
	}
}

func (handler *Handler) writeServiceError(writer http.ResponseWriter, request *http.Request, err error, fallbackKey string) {
	if statusCode, isKnownError := statusCodeForError(err); isKnownError {
		response.WriteTranslatedError(writer, request, handler.translator, statusCode, err.Error())
		return
	}

	log.Printf("Erro inesperado em %s %s: %v", request.Method, request.URL.Path, err)
	response.WriteTranslatedError(writer, request, handler.translator, http.StatusInternalServerError, fallbackKey)
}

func (handler *Handler) Register(writer http.ResponseWriter, request *http.Request) {
	var body struct {
		Name        string         `json:"name"`
		Email       string         `json:"email"`
		Password    string         `json:"password"`
		Preferences map[string]any `json:"preferences"`
	}

	if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
		response.WriteTranslatedError(writer, request, handler.translator, http.StatusBadRequest, "common.invalidRequestBody")
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

	ipAddress, userAgent := audit.ClientInfoFromRequest(request)

	handler.auditService.Log(request.Context(), audit.Entry{
		UserID:        &createdUser.ID,
		PerformedByID: &createdUser.ID,
		Action:        audit.ActionUserRegister,
		Entity:        audit.EntityUser,
		EntityID:      &createdUser.ID,
		Details:       map[string]any{"email": createdUser.Email},
		IPAddress:     ipAddress,
		UserAgent:     userAgent,
	})

	response.WriteJSON(writer, http.StatusCreated, registerResponse{
		ID:        createdUser.ID,
		Email:     createdUser.Email,
		CreatedAt: createdUser.CreatedAt,
	})
}
