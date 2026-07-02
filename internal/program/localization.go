package program

import "promptfulcustomffmpegbuilder/localization"

func LLocaleTextGetInternal(key string, values map[string]string) string {
	return localization.LLocaleTextGet(key, values)
}

func LLocaleTextForGet(locale string, key string, values map[string]string) string {
	return localization.LLocaleTextForGet(locale, key, values)
}
