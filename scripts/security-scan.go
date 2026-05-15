//go:build ignore

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type sourceRule struct {
	needle          string
	allowedPrefixes []string
	message         string
}

func main() {
	violations := make([]string, 0)
	rules := []sourceRule{
		{needle: "exec.Command", allowedPrefixes: []string{"internal/execution/"}, message: "exec.Command usage is allowed only in internal/execution"},
		{needle: "os.StartProcess", allowedPrefixes: []string{}, message: "os.StartProcess is forbidden"},
		{needle: "syscall.Exec", allowedPrefixes: []string{}, message: "syscall.Exec is forbidden"},
		{needle: "http.NewRequest", allowedPrefixes: []string{"internal/download/"}, message: "HTTP request creation is allowed only in internal/download"},
		{needle: "http.Client", allowedPrefixes: []string{"internal/download/"}, message: "HTTP clients are allowed only in internal/download"},
		{needle: ".Do(request)", allowedPrefixes: []string{"internal/download/"}, message: "HTTP request execution is allowed only in internal/download"},
		{needle: "net.Dial", allowedPrefixes: []string{}, message: "raw network dialing is forbidden"},
		{needle: "os.RemoveAll", allowedPrefixes: []string{"internal/workspace/"}, message: "recursive deletion is allowed only in internal/workspace"},
		{needle: "os.Remove(", allowedPrefixes: []string{"internal/download/", "internal/workspace/"}, message: "file deletion is allowed only in controlled packages"},
		{needle: "os.Rename", allowedPrefixes: []string{"internal/download/", "internal/workspace/"}, message: "file renaming is allowed only in controlled packages"},
		{needle: "os.WriteFile", allowedPrefixes: []string{"internal/scripting/", "internal/audit/", "app.go"}, message: "direct file writes are allowed only in scripting/audit/reporting code"},
		{needle: "\"-lc\"", allowedPrefixes: []string{}, message: "bash -lc shell-string execution is forbidden"},
		{needle: "unsafe.", allowedPrefixes: []string{}, message: "unsafe package usage is forbidden"},
	}
	err := filepath.WalkDir(".", func(path string, directoryEntry os.DirEntry, walkError error) error {
		if walkError != nil {
			return walkError
		}
		if directoryEntry.IsDir() || filepath.Ext(path) != ".go" {
			return nil
		}
		normalizedPath := filepath.ToSlash(path)
		if strings.HasPrefix(normalizedPath, "scripts/") {
			return nil
		}
		contentBytes, readError := os.ReadFile(path)
		if readError != nil {
			return readError
		}
		contentText := string(contentBytes)
		for _, rule := range rules {
			if strings.Contains(contentText, rule.needle) && !isAllowedPath(normalizedPath, rule.allowedPrefixes) {
				violations = append(violations, path+": "+rule.message)
			}
		}
		return nil
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if len(violations) > 0 {
		fmt.Fprintln(os.Stderr, "security scan failed:")
		for _, violation := range violations {
			fmt.Fprintln(os.Stderr, " - "+violation)
		}
		os.Exit(1)
	}
	fmt.Println("security scan passed")
}

func isAllowedPath(path string, allowedPrefixes []string) bool {
	for _, allowedPrefix := range allowedPrefixes {
		if path == allowedPrefix || strings.HasPrefix(path, allowedPrefix) {
			return true
		}
	}
	return false
}
