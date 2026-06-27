package localization

import (
	"embed"
	"encoding/json"
	"strings"
)

//go:embed en.json ko.json
var localizationFiles embed.FS

var englishLocalization = loadLocalization("en.json")
var localizationsByLocale = map[string]map[string]string{
	"en": englishLocalization,
	"ko": loadLocalization("ko.json"),
}

func loadLocalization(fileName string) map[string]string {
	bytes, err := localizationFiles.ReadFile(fileName)
	if err != nil {
		return map[string]string{}
	}
	values := map[string]string{}
	if err := json.Unmarshal(bytes, &values); err != nil {
		return map[string]string{}
	}
	return values
}

func interpolate(template string, values map[string]string) string {
	for name, value := range values {
		template = strings.ReplaceAll(template, "{"+name+"}", value)
	}
	return template
}

// Localize resolves a key in English. Used for backend log and error messages,
// which stay English regardless of UI language.
func Localize(key string, values map[string]string) string {
	template := englishLocalization[key]
	if template == "" {
		return key
	}
	return interpolate(template, values)
}

// LocalizeFor resolves a key in the given locale, falling back to English per
// key so an untranslated key still shows English text rather than a bare key.
// Used for the native confirmation dialog, the one backend-rendered surface that
// follows the UI language.
func LocalizeFor(locale string, key string, values map[string]string) string {
	if dictionary, ok := localizationsByLocale[locale]; ok {
		if template, ok := dictionary[key]; ok && template != "" {
			return interpolate(template, values)
		}
	}
	return Localize(key, values)
}
