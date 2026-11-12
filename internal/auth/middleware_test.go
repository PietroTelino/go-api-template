package auth

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/PietroTelino/go-api-template/internal/i18n"
)

func newTestTranslator(t *testing.T) *i18n.Translator {
	t.Helper()

	translator, err := i18n.New("pt-BR", []string{"pt-BR", "en"})
	if err != nil {
		t.Fatalf("erro ao carregar as traduções: %v", err)
	}

	return translator
}

func responseMessage(t *testing.T, recorder *httptest.ResponseRecorder) string {
	t.Helper()

	var body struct {
		Message string `json:"message"`
	}

	if err := json.NewDecoder(recorder.Body).Decode(&body); err != nil {
		t.Fatalf("resposta não é um JSON válido: %v", err)
	}

	return body.Message
}

func TestRequireAuthentication(t *testing.T) {
	tokenManager := newTestTokenManager(900)
	translator := newTestTranslator(t)
	validToken, _ := tokenManager.GenerateAccessToken("id-do-usuario", "administrator")

	var receivedUser *AuthenticatedUser
	protectedHandler := RequireAuthentication(tokenManager, translator)(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		receivedUser, _ = AuthenticatedUserFromContext(request.Context())
		writer.WriteHeader(http.StatusOK)
	}))

	testCases := []struct {
		name          string
		authorization string
		wantStatus    int
		wantMessage   string
	}{
		{"sem header", "", http.StatusUnauthorized, "Token não fornecido"},
		{"sem Bearer", "Basic abc", http.StatusUnauthorized, "Formato de token inválido"},
		{"Bearer vazio", "Bearer ", http.StatusUnauthorized, "Formato de token inválido"},
		{"token inválido", "Bearer abc", http.StatusUnauthorized, "Token inválido ou expirado"},
		{"token válido", "Bearer " + validToken, http.StatusOK, ""},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			receivedUser = nil

			request := httptest.NewRequest(http.MethodGet, "/protegida", nil)
			if testCase.authorization != "" {
				request.Header.Set("Authorization", testCase.authorization)
			}

			recorder := httptest.NewRecorder()
			protectedHandler.ServeHTTP(recorder, request)

			if recorder.Code != testCase.wantStatus {
				t.Fatalf("status = %d, esperado %d", recorder.Code, testCase.wantStatus)
			}

			if testCase.wantStatus != http.StatusOK {
				if message := responseMessage(t, recorder); message != testCase.wantMessage {
					t.Errorf("mensagem = %q, esperado %q", message, testCase.wantMessage)
				}
				return
			}

			if receivedUser == nil || receivedUser.ID != "id-do-usuario" || receivedUser.Role != "administrator" {
				t.Errorf("usuário no contexto = %+v", receivedUser)
			}
		})
	}
}

func TestRequiredRoles(t *testing.T) {
	translator := newTestTranslator(t)

	adminOnlyHandler := RequiredRoles(translator, "administrator", "god")(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusOK)
	}))

	testCases := []struct {
		name       string
		user       *AuthenticatedUser
		wantStatus int
	}{
		{"sem usuário no contexto", nil, http.StatusUnauthorized},
		{"user comum", &AuthenticatedUser{ID: "1", Role: "user"}, http.StatusForbidden},
		{"administrator", &AuthenticatedUser{ID: "1", Role: "administrator"}, http.StatusOK},
		{"god", &AuthenticatedUser{ID: "1", Role: "god"}, http.StatusOK},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, "/admin", nil)
			if testCase.user != nil {
				request = request.WithContext(WithAuthenticatedUser(request.Context(), testCase.user))
			}

			recorder := httptest.NewRecorder()
			adminOnlyHandler.ServeHTTP(recorder, request)

			if recorder.Code != testCase.wantStatus {
				t.Errorf("status = %d, esperado %d", recorder.Code, testCase.wantStatus)
			}
		})
	}
}
