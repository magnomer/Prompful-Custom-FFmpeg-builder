package localization

import (
	"embed"
	"encoding/json"
	"strings"
)

//go:embed en.json
var localizationFiles embed.FS

var englishLocalization = loadEnglishLocalization()

func loadEnglishLocalization() map[string]string {
	bytes, err := localizationFiles.ReadFile("en.json")
	if err != nil {
		return map[string]string{}
	}
	values := map[string]string{}
	if err := json.Unmarshal(bytes, &values); err != nil {
		return map[string]string{}
	}
	return values
}

func Localize(key string, values map[string]string) string {
	template := englishLocalization[key]
	if template == "" {
		return key
	}
	for name, value := range values {
		template = strings.ReplaceAll(template, "{"+name+"}", value)
	}
	return template
}
