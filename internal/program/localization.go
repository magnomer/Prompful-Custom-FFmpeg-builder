package program

import sharedlocalization "promptfulcustomffmpegbuilder/shared/localization"

func LLocaleTextGetInternal(key string, values map[string]string) string {
	return sharedlocalization.LLocaleTextGet(key, values)
}

func LLocaleTextForGet(locale string, key string, values map[string]string) string {
	return sharedlocalization.LLocaleTextForGet(locale, key, values)
}
