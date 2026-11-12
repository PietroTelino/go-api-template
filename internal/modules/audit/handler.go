package audit

import (
	"log"
	"net/http"
	"strconv"

	"github.com/PietroTelino/go-api-template/internal/auth"
	"github.com/PietroTelino/go-api-template/internal/i18n"
	"github.com/PietroTelino/go-api-template/internal/response"
	"github.com/PietroTelino/go-api-template/internal/validation"
	"github.com/go-chi/chi/v5"
)

const (
	defaultLimit = 100
	maxLimit     = 100
)

type Handler struct {
	service    *Service
	translator *i18n.Translator
}

type pagination struct {
	Total  int `json:"total"`
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

type listResponse struct {
	Data       []*Record  `json:"data"`
	Pagination pagination `json:"pagination"`
}

func NewHandler(service *Service, translator *i18n.Translator) *Handler {
	return &Handler{service: service, translator: translator}
}

func queryInt(request *http.Request, key string, fallback int, minimum int, maximum int) int {
	number, err := strconv.Atoi(request.URL.Query().Get(key))
	if err != nil {
		return fallback
	}

	return min(max(number, minimum), maximum)
}

func paginationFromRequest(request *http.Request) (int, int) {
	limit := queryInt(request, "limit", defaultLimit, 1, maxLimit)
	offset := queryInt(request, "offset", 0, 0, 1_000_000)

	return limit, offset
}

func (handler *Handler) writeList(writer http.ResponseWriter, request *http.Request, filters Filters, errorKey string) {
	records, total, err := handler.service.List(request.Context(), filters)
	if err != nil {
		log.Printf("Erro inesperado em %s %s: %v", request.Method, request.URL.Path, err)
		response.WriteTranslatedError(writer, request, handler.translator, http.StatusInternalServerError, errorKey)
		return
	}

	response.WriteJSON(writer, http.StatusOK, listResponse{
		Data:       records,
		Pagination: pagination{Total: total, Limit: filters.Limit, Offset: filters.Offset},
	})
}

func (handler *Handler) List(writer http.ResponseWriter, request *http.Request) {
	query := request.URL.Query()

	filters := Filters{
		UserID:        query.Get("userId"),
		PerformedByID: query.Get("performedById"),
		Action:        query.Get("action"),
		Entity:        query.Get("entity"),
	}

	filters.Limit, filters.Offset = paginationFromRequest(request)

	if (filters.UserID != "" && !validation.IsUUID(filters.UserID)) ||
		(filters.PerformedByID != "" && !validation.IsUUID(filters.PerformedByID)) {
		response.WriteJSON(writer, http.StatusOK, listResponse{
			Data:       []*Record{},
			Pagination: pagination{Total: 0, Limit: filters.Limit, Offset: filters.Offset},
		})
		return
	}

	handler.writeList(writer, request, filters, "audit.errorList")
}

func (handler *Handler) ListMine(writer http.ResponseWriter, request *http.Request) {
	authenticatedUser, isAuthenticated := auth.AuthenticatedUserFromContext(request.Context())
	if !isAuthenticated {
		response.WriteTranslatedError(writer, request, handler.translator, http.StatusUnauthorized, "auth.notAuthenticated")
		return
	}

	filters := Filters{UserID: authenticatedUser.ID}
	filters.Limit, filters.Offset = paginationFromRequest(request)

	handler.writeList(writer, request, filters, "audit.errorListMine")
}

func (handler *Handler) GetByID(writer http.ResponseWriter, request *http.Request) {
	id := chi.URLParam(request, "id")

	if !validation.IsUUID(id) {
		response.WriteTranslatedError(writer, request, handler.translator, http.StatusNotFound, "audit.notFound")
		return
	}

	record, err := handler.service.GetByID(request.Context(), id)
	if err != nil {
		log.Printf("Erro inesperado em %s %s: %v", request.Method, request.URL.Path, err)
		response.WriteTranslatedError(writer, request, handler.translator, http.StatusInternalServerError, "audit.errorFind")
		return
	}

	if record == nil {
		response.WriteTranslatedError(writer, request, handler.translator, http.StatusNotFound, "audit.notFound")
		return
	}

	response.WriteJSON(writer, http.StatusOK, record)
}
