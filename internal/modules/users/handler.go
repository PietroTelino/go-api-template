package users

import (
	"encoding/json"
	"errors"
	"net/http"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (handler *Handler) Register(writer http.ResponseWriter, request *http.Request) {
	var body struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
		writeError(writer, http.StatusBadRequest, "Corpo da requisição inválido")
		return
	}

	if body.Name == "" || body.Email == "" || body.Password == "" {
		writeError(writer, http.StatusBadRequest, "Name, email e password são obrigatórios")
		return
	}

	createdUser, err := handler.service.Register(request.Context(), RegisterInput{
		Name:     body.Name,
		Email:    body.Email,
		Password: body.Password,
	})

	if errors.Is(err, ErrEmailAlreadyInUse) {
		writeError(writer, http.StatusConflict, "Email já está em uso")
		return
	}

	if err != nil {
		writeError(writer, http.StatusInternalServerError, "Erro ao registrar usuário")
		return
	}

	writeJSON(writer, http.StatusCreated, createdUser)
}

func writeJSON(writer http.ResponseWriter, status int, data any) {
	writer.Header().Set("Content-type", "application/json")
	writer.WriteHeader(status)
	json.NewEncoder(writer).Encode(data)
}

func writeError(writer http.ResponseWriter, status int, message string) {
	writeJSON(writer, status, map[string]string{"message": message})
}
