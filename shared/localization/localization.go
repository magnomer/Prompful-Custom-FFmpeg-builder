package localization

import (
	"embed"
	"encoding/json"
	"strings"
)

//go:embed en.json ko.json
var LLocaleFiles embed.FS

var LLocaleEnglish = LLocaleLoad("en.json")
var LLocaleMap = map[string]map[string]string{
	"en": LLocaleEnglish,
	"ko": LLocaleLoad("ko.json"),
}

func LLocaleLoad(fileName string) map[string]string {
	bytes, err := LLocaleFiles.ReadFile(fileName)
	if err != nil {
		return map[string]string{}
	}
	values := map[string]string{}
	if err := json.Unmarshal(bytes, &values); err != nil {
		return map[string]string{}
	}
	return values
}

func LTextInterpolate(template string, values map[string]string) string {
	for name, value := range values {
		template = strings.ReplaceAll(template, "{"+name+"}", value)
	}
	return template
}

// LLocaleTextGet resolves a key in English. Used for backend log and error messages,
// which stay English regardless of UI language.
func LLocaleTextGet(key string, values map[string]string) string {
	template := LLocaleEnglish[key]
	if template == "" {
		return key
	}
	return LTextInterpolate(template, values)
}

// LLocaleTextForGet resolves a key in the given locale, falling back to English per
// key so an untranslated key still shows English text rather than a bare key.
// Used for the native confirmation dialog, the one backend-rendered surface that
// follows the UI language.
func LLocaleTextForGet(locale string, key string, values map[string]string) string {
	if dictionary, ok := LLocaleMap[locale]; ok {
		if template, ok := dictionary[key]; ok && template != "" {
			return LTextInterpolate(template, values)
		}
	}
	return LLocaleTextGet(key, values)
}
