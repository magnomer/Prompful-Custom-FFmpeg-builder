package app

import sharedlocalization "promptfulcustomffmpegbuilder/shared/localization"

func localize(key string, values map[string]string) string {
	return sharedlocalization.Localize(key, values)
}

func localizeForLocale(locale string, key string, values map[string]string) string {
	return sharedlocalization.LocalizeFor(locale, key, values)
}
