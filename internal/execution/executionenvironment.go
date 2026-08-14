package execution

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func LEnvironmentMsysCreate(commandPlan LPlanCommand) []string {
	environmentByName := map[string]string{}
	environmentNameOrder := []string{}
	setEnvironmentValue := func(environmentName string, environmentValue string) {
		if _, exists := environmentByName[environmentName]; !exists {
			environmentNameOrder = append(environmentNameOrder, environmentName)
		}
		environmentByName[environmentName] = environmentValue
	}

	for _, inheritedEnvironmentValue := range os.Environ() {
		environmentName, environmentValue, found := strings.Cut(inheritedEnvironmentValue, "=")
		if !found || environmentName == "" {
			continue
		}
		setEnvironmentValue(environmentName, environmentValue)
	}

	setEnvironmentValue("MSYS2_PATH_TYPE", "minimal")
	setEnvironmentValue("CHERE_INVOKING", "1")

	if commandPlan.Msys2RootDirectory != "" && commandPlan.WindowsShellProfileName != "" {
		profileDirectoryName := strings.ToLower(commandPlan.WindowsShellProfileName)
		shellSystemName := strings.ToUpper(profileDirectoryName)
		existingPath := environmentByName["PATH"]
		msys2Path := filepath.Join(commandPlan.Msys2RootDirectory, profileDirectoryName, "bin") + string(os.PathListSeparator) + filepath.Join(commandPlan.Msys2RootDirectory, "usr", "bin")
		if existingPath != "" {
			msys2Path += string(os.PathListSeparator) + existingPath
		}

		msys2UnixPrefix := "/" + profileDirectoryName
		mingwPackagePrefix := "mingw-w64-" + profileDirectoryName + "-x86_64"
		mingwChost := "x86_64-w64-mingw32"
		if profileDirectoryName == "clangarm64" {
			mingwPackagePrefix = "mingw-w64-clang-aarch64"
			mingwChost = "aarch64-w64-mingw32"
		} else if profileDirectoryName == "clang64" {
			mingwPackagePrefix = "mingw-w64-clang-x86_64"
		} else if profileDirectoryName == "mingw64" {
			mingwPackagePrefix = "mingw-w64-x86_64"
		}

		setEnvironmentValue("MSYSTEM", shellSystemName)
		setEnvironmentValue("MSYSTEM_PREFIX", msys2UnixPrefix)
		setEnvironmentValue("MINGW_PREFIX", msys2UnixPrefix)
		setEnvironmentValue("MINGW_CHOST", mingwChost)
		setEnvironmentValue("MINGW_PACKAGE_PREFIX", mingwPackagePrefix)
		setEnvironmentValue("PKG_CONFIG_PATH", msys2UnixPrefix+"/lib/pkgconfig:"+msys2UnixPrefix+"/share/pkgconfig:/usr/lib/pkgconfig:/usr/share/pkgconfig")
		setEnvironmentValue("PKG_CONFIG_LIBDIR", msys2UnixPrefix+"/lib/pkgconfig:"+msys2UnixPrefix+"/share/pkgconfig:/usr/lib/pkgconfig:/usr/share/pkgconfig")
		setEnvironmentValue("PATH", msys2Path)
	}

	for environmentName, environmentValue := range commandPlan.EnvironmentVariables {
		if !LEnvironmentSafeCheck(environmentName, environmentValue) {
			continue
		}
		setEnvironmentValue(environmentName, environmentValue)
	}

	environment := make([]string, 0, len(environmentNameOrder))
	for _, environmentName := range environmentNameOrder {
		environment = append(environment, fmt.Sprintf("%s=%s", environmentName, environmentByName[environmentName]))
	}
	return environment
}

func LEnvironmentSafeCheck(environmentName string, environmentValue string) bool {
	if environmentName == "" || strings.Contains(environmentName, "=") || strings.Contains(environmentName, "\x00") || strings.Contains(environmentValue, "\x00") {
		return false
	}
	for _, environmentNameRune := range environmentName {
		if !(environmentNameRune == '_' || environmentNameRune >= 'A' && environmentNameRune <= 'Z' || environmentNameRune >= 'a' && environmentNameRune <= 'z' || environmentNameRune >= '0' && environmentNameRune <= '9') {
			return false
		}
	}
	return true
}
