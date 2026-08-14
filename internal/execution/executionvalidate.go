package execution

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"promptfulcustomffmpegbuilder/internal/workspace"
)

func LPlanCommandValidate(commandPlan LPlanCommand) error {
	if commandPlan.ExecutablePath == "" {
		return errors.New("approved command executable path is empty")
	}
	if commandPlan.WorkingDirectory == "" || commandPlan.WorkspaceDirectory == "" {
		return errors.New("approved command directories are empty")
	}
	if err := workspace.LPathWorkspaceCheck(commandPlan.WorkspaceDirectory, commandPlan.WorkingDirectory); err != nil {
		return err
	}
	if err := workspace.LPathRealCheck(commandPlan.WorkspaceDirectory, commandPlan.WorkingDirectory); err != nil {
		return err
	}
	if err := workspace.LPathWorkspaceCheck(commandPlan.WorkspaceDirectory, commandPlan.ExecutablePath); err != nil {
		return err
	}
	if err := workspace.LPathRealCheck(commandPlan.WorkspaceDirectory, commandPlan.ExecutablePath); err != nil {
		return err
	}
	if len(commandPlan.AllowedExecutableBasenames) == 0 {
		return errors.New("approved command must include executable basename allowlist")
	}
	if !LProgramAllowedCheck(filepath.Base(commandPlan.ExecutablePath), commandPlan.AllowedExecutableBasenames) {
		return fmt.Errorf("approved command executable is not allowlisted: %s", filepath.Base(commandPlan.ExecutablePath))
	}
	if commandPlan.Msys2RootDirectory != "" {
		if err := workspace.LPathWorkspaceCheck(commandPlan.WorkspaceDirectory, commandPlan.Msys2RootDirectory); err != nil {
			return err
		}
		if err := workspace.LPathRealCheck(commandPlan.WorkspaceDirectory, commandPlan.Msys2RootDirectory); err != nil {
			return err
		}
	}
	if commandPlan.RunLAuditDirectoryGet != "" {
		if err := workspace.LPathWorkspaceCheck(commandPlan.WorkspaceDirectory, commandPlan.RunLAuditDirectoryGet); err != nil {
			return err
		}
		if err := workspace.LPathRealCheck(commandPlan.WorkspaceDirectory, filepath.Dir(commandPlan.RunLAuditDirectoryGet)); err != nil {
			return err
		}
	}
	if LTextShellCheck(commandPlan.ExecutablePath) {
		return errors.New("approved command executable contains shell metacharacters")
	}
	for _, argumentValue := range commandPlan.ArgumentValues {
		if strings.Contains(argumentValue, "\x00") {
			return errors.New("approved command argument contains null byte")
		}
	}
	if commandPlan.ApprovedScriptFilePath != "" {
		if commandPlan.LScriptKind == "" {
			return errors.New("approved script kind is empty")
		}
		if err := LHashScriptCheck(commandPlan); err != nil {
			return err
		}
	}
	return nil
}

func LHashScriptCheck(commandPlan LPlanCommand) error {
	_, _, err := LScriptStdinPrepare(commandPlan)
	return err
}

func LScriptStdinPrepare(commandPlan LPlanCommand) ([]byte, []string, error) {
	if commandPlan.ApprovedScriptSha256Hash == "" {
		return nil, nil, errors.New("approved script hash is empty")
	}
	if err := workspace.LPathWorkspaceCheck(commandPlan.WorkspaceDirectory, commandPlan.ApprovedScriptFilePath); err != nil {
		return nil, nil, err
	}
	if err := workspace.LPathRealCheck(commandPlan.WorkspaceDirectory, commandPlan.ApprovedScriptFilePath); err != nil {
		return nil, nil, err
	}
	updatedArgumentValues, foundScriptArgument := LStdinFlagReplace(commandPlan.ArgumentValues, commandPlan.ApprovedScriptFilePath)
	if !foundScriptArgument {
		return nil, nil, errors.New("approved script path is not present in command arguments")
	}
	file, err := os.Open(commandPlan.ApprovedScriptFilePath)
	if err != nil {
		return nil, nil, err
	}
	defer file.Close()
	scriptBytes, err := io.ReadAll(file)
	if err != nil {
		return nil, nil, err
	}
	hash := sha256.Sum256(scriptBytes)
	actualScriptSha256Hash := hex.EncodeToString(hash[:])
	if !strings.EqualFold(actualScriptSha256Hash, commandPlan.ApprovedScriptSha256Hash) {
		return nil, nil, errors.New("approved script hash does not match script content")
	}
	return scriptBytes, updatedArgumentValues, nil
}

func LTextShellCheck(value string) bool {
	return strings.ContainsAny(value, ";&|><`$\n\r")
}

func LProgramAllowedCheck(executableBasename string, allowedExecutableBasenames []string) bool {
	for _, allowedExecutableBasename := range allowedExecutableBasenames {
		if strings.EqualFold(executableBasename, allowedExecutableBasename) {
			return true
		}
	}
	return false
}

func LStdinFlagReplace(argumentValues []string, scriptFilePath string) ([]string, bool) {
	updatedArgumentValues := make([]string, len(argumentValues))
	copy(updatedArgumentValues, argumentValues)
	for index, argumentValue := range updatedArgumentValues {
		if argumentValue == scriptFilePath || argumentValue == filepath.ToSlash(scriptFilePath) {
			updatedArgumentValues[index] = "-s"
			return updatedArgumentValues, true
		}
	}
	return updatedArgumentValues, false
}
