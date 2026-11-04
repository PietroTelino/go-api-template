package users

import (
	"encoding/json"
	"errors"
	"net/http"

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

func (handler *Handler) Register(writer http.ResponseWriter, request *http.Request) {
	var body struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
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
		Name:     body.Name,
		Email:    body.Email,
		Password: body.Password,
	})

	if errors.Is(err, ErrEmailAlreadyInUse) {
		response.WriteTranslatedError(writer, request, handler.translator, http.StatusConflict, "user.emailInUse")
		return
	}

	if err != nil {
		response.WriteTranslatedError(writer, request, handler.translator, http.StatusInternalServerError, "user.errorRegister")
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

	response.WriteJSON(writer, http.StatusCreated, createdUser)
}
