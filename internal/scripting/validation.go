package scripting

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

var LPackageSafePattern = regexp.MustCompile(`^[A-Za-z0-9_+.-]+$`)
var LFlagSafePattern = regexp.MustCompile(`^--[A-Za-z0-9][A-Za-z0-9_+./:=,-]*$`)

func LPackageMsysValidate(packageName string) error {
	if packageName == "" {
		return errors.New("MSYS2 package name is empty")
	}
	if !LPackageSafePattern.MatchString(packageName) {
		return fmt.Errorf("MSYS2 package name contains unsafe characters: %s", packageName)
	}
	return nil
}

func LFlagConfigureValidate(configureFlag string) error {
	if configureFlag == "" {
		return errors.New("configure flag is empty")
	}
	if !LFlagSafePattern.MatchString(configureFlag) {
		return fmt.Errorf("configure flag contains unsafe characters: %s", configureFlag)
	}
	if strings.Contains(configureFlag, "..") {
		return fmt.Errorf("configure flag contains unsafe path traversal marker: %s", configureFlag)
	}
	return nil
}
