package planning

func LStringDefaultGet(record map[string]any, fieldName string, defaultValue string) string {
	value := LCatalogFieldGet(record, fieldName)
	if value == "" {
		return defaultValue
	}
	return value
}

func LCatalogBooleanGet(record map[string]any, fieldName string) bool {
	if record == nil {
		return false
	}
	value, ok := record[fieldName].(bool)
	return ok && value
}

func LArrayFieldGet(record map[string]any, fieldName string) []string {
	values, ok := record[fieldName].([]any)
	if !ok {
		return nil
	}
	result := []string{}
	for _, value := range values {
		stringValue, ok := value.(string)
		if ok && stringValue != "" {
			result = append(result, stringValue)
		}
	}
	return result
}

func LShellPackageRead(versionObject map[string]any, shellProfileName string) []string {
	if LShellProfileCheck(versionObject, shellProfileName) {
		return nil
	}
	packageNamesByShellProfile, ok := versionObject["packageNamesByShellProfile"].(map[string]any)
	if !ok {
		return nil
	}
	return LArrayFieldGet(packageNamesByShellProfile, shellProfileName)
}

func LShellProfileCheck(versionObject map[string]any, shellProfileName string) bool {
	for _, unavailableShellProfileName := range LArrayFieldGet(versionObject, "unavailableShellProfiles") {
		if unavailableShellProfileName == shellProfileName {
			return true
		}
	}
	return false
}

func LWorkIdentifierCreate(ffmpegVersion string, libraryId string) string {
	return "ffmpeg-" + ffmpegVersion + "-" + libraryId + "-work"
}
