package i18n

import (
	"embed"
	"encoding/json"
	"fmt"
	"strings"
)

var embeddedTranslationFiles embed.FS

type Translator struct {
	translationsByLanguage map[string]map[string]any
	fallbackLanguage       string
	supportedLanguages     []string
}

func New(fallbackLanguage string, supportedLanguages []string) (*Translator, error) {
	translationsByLanguage := make(map[string]map[string]any, len(supportedLanguages))

	for _, language := range supportedLanguages {
		filepath := fmt.Sprintf("locales/%s/translation.json", language)

		fileBytes, err := embeddedTranslationFiles.ReadFile(filepath)
		if err != nil {
			return nil, fmt.Errorf("Erro ao ler o arquivo de tradução do idioma %q: %w", language, err)
		}

		var parsedTranslations map[string]any
		if err := json.Unmarshal(fileBytes, &parsedTranslations); err != nil {
			return nil, fmt.Errorf("Erro ao decodificar o arquivo de tradução do idioma %q: %w", language, err)
		}

		translationsByLanguage[language] = parsedTranslations
	}

	return &Translator{
		translationsByLanguage: translationsByLanguage,
		fallbackLanguage:       fallbackLanguage,
		supportedLanguages:     supportedLanguages,
	}, nil
}

func (translator *Translator) Translate(language string, translationKey string) string {
	if translations, languageIsSupported := translator.translationsByLanguage[language]; languageIsSupported {
		if message, keyWasFound := lookupNestedKey(translations, translationKey); keyWasFound {
			return message
		}
	}

	if language != translator.fallbackLanguage {
		if fallbackTranslations, ok := translator.translationsByLanguage[translator.fallbackLanguage]; ok {
			if message, keyWasFound := lookupNestedKey(fallbackTranslations, translationKey); keyWasFound {
				return message
			}
		}
	}

	return translationKey
}

func (translator *Translator) SupportedLanguages() []string {
	return translator.supportedLanguages
}

func (translator *Translator) FallbackLanguage() string {
	return translator.fallbackLanguage
}

func lookupNestedKey(translations map[string]any, dotSeparatedKey string) (string, bool) {
	keyParts := strings.Split(dotSeparatedKey, ".")

	var currentValue any = translations
	for _, keyPart := range keyParts {
		currentMap, isMap := currentValue.(map[string]any)
		if !isMap {
			return "", false
		}

		nextValue, exists := currentMap[keyPart]
		if !exists {
			return "", false
		}

		currentValue = nextValue
	}

	translatedMessage, isString := currentValue.(string)
	return translatedMessage, isString
}
