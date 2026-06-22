package app

import sharedlocalization "customffmpegbuilder/shared/localization"

func localize(key string, values map[string]string) string {
	return sharedlocalization.Localize(key, values)
}
