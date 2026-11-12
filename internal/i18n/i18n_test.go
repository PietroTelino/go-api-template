package i18n

import "testing"

func newTestTranslator(t *testing.T) *Translator {
	t.Helper()

	translator, err := New("pt-BR", []string{"pt-BR", "en"})
	if err != nil {
		t.Fatalf("erro ao carregar as traduções: %v", err)
	}

	return translator
}

func TestTranslate(t *testing.T) {
	translator := newTestTranslator(t)

	testCases := []struct {
		name     string
		language string
		key      string
		want     string
	}{
		{"português", "pt-BR", "user.notFound", "Usuário não encontrado"},
		{"inglês", "en", "user.notFound", "User not found"},
		{"idioma não suportado usa o padrão", "es", "user.notFound", "Usuário não encontrado"},
		{"idioma vazio usa o padrão", "", "user.notFound", "Usuário não encontrado"},
		{"chave inexistente devolve a própria chave", "pt-BR", "nao.existe", "nao.existe"},
		{"chave que aponta para um grupo devolve a chave", "pt-BR", "user", "user"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			if got := translator.Translate(testCase.language, testCase.key); got != testCase.want {
				t.Errorf("Translate(%q, %q) = %q, esperado %q", testCase.language, testCase.key, got, testCase.want)
			}
		})
	}
}

func TestTranslationFilesHaveSameKeys(t *testing.T) {
	translator := newTestTranslator(t)

	portugueseKeys := flattenKeys(translator.translationsByLanguage["pt-BR"], "")
	englishKeys := flattenKeys(translator.translationsByLanguage["en"], "")

	for key := range portugueseKeys {
		if !englishKeys[key] {
			t.Errorf("a chave %q existe em pt-BR, mas não em en", key)
		}
	}

	for key := range englishKeys {
		if !portugueseKeys[key] {
			t.Errorf("a chave %q existe em en, mas não em pt-BR", key)
		}
	}
}

func flattenKeys(translations map[string]any, prefix string) map[string]bool {
	keys := map[string]bool{}

	for key, value := range translations {
		if nested, isMap := value.(map[string]any); isMap {
			for nestedKey := range flattenKeys(nested, prefix+key+".") {
				keys[nestedKey] = true
			}
			continue
		}

		keys[prefix+key] = true
	}

	return keys
}

func TestResolveLanguageFromHeader(t *testing.T) {
	supportedLanguages := []string{"pt-BR", "en"}

	testCases := []struct {
		header string
		want   string
	}{
		{"", "pt-BR"},
		{"en", "en"},
		{"EN", "en"},
		{"en-US,en;q=0.9", "en"},
		{"pt", "pt-BR"},
		{"fr-FR,en;q=0.8", "en"},
		{"es", "pt-BR"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.header, func(t *testing.T) {
			if got := resolveLanguageFromHeader(testCase.header, supportedLanguages, "pt-BR"); got != testCase.want {
				t.Errorf("resolveLanguageFromHeader(%q) = %q, esperado %q", testCase.header, got, testCase.want)
			}
		})
	}
}
