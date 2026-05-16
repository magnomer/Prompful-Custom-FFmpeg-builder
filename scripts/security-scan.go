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
	description     string
}

func main() {
	violations := make([]string, 0)
	observedDescriptions := make(map[string]bool)

	// This scanner documents the source boundaries used by this app.
	// It is intentionally small and text-based: it looks for sensitive Go
	// primitives and checks whether they appear in the parts of the app that
	// currently own those responsibilities.
	rules := []sourceRule{
		{
			needle:          "exec.Command",
			allowedPrefixes: []string{"app.go", "internal/execution/"},
			description:     "processes are started either by the app shell integration or by the managed execution layer",
		},
		{
			needle:          "os.StartProcess",
			allowedPrefixes: []string{},
			description:     "low-level process startup is not part of this app",
		},
		{
			needle:          "syscall.Exec",
			allowedPrefixes: []string{},
			description:     "process replacement is not part of this app",
		},
		{
			needle:          "http.NewRequest",
			allowedPrefixes: []string{"internal/download/"},
			description:     "network downloads are implemented by the download layer",
		},
		{
			needle:          "http.Client",
			allowedPrefixes: []string{"internal/download/"},
			description:     "HTTP clients are owned by the download layer",
		},
		{
			needle:          ".Do(request)",
			allowedPrefixes: []string{"internal/download/"},
			description:     "HTTP request execution is owned by the download layer",
		},
		{
			needle:          "net.Dial",
			allowedPrefixes: []string{},
			description:     "raw network dialing is not part of this app",
		},
		{
			needle:          "os.RemoveAll",
			allowedPrefixes: []string{"app.go", "internal/workspace/"},
			description:     "recursive deletion is limited to workspace cleanup code paths",
		},
		{
			needle:          "os.Remove(",
			allowedPrefixes: []string{"app.go", "internal/download/", "internal/scripting/", "internal/workspace/"},
			description:     "single-file deletion is used for workspace cleanup, download replacement, and generated script replacement",
		},
		{
			needle:          "os.Rename",
			allowedPrefixes: []string{"app.go", "internal/download/", "internal/workspace/"},
			description:     "file and directory moves are used for workspace setup and atomic download replacement",
		},
		{
			needle:          "os.WriteFile",
			allowedPrefixes: []string{"app.go", "internal/audit/", "internal/scripting/"},
			description:     "direct file writes are used for generated scripts, audit/report files, and app-owned result files",
		},
		{
			needle:          "\"-lc\"",
			allowedPrefixes: []string{"app.go"},
			description:     "shell-string execution is limited to app-owned MSYS2 maintenance commands",
		},
		{
			needle:          "unsafe.",
			allowedPrefixes: []string{},
			description:     "unsafe Go operations are not part of this app",
		},
	}

	err := filepath.WalkDir(".", func(path string, directoryEntry os.DirEntry, walkError error) error {
		if walkError != nil {
			return walkError
		}
		if directoryEntry.IsDir() {
			return nil
		}
		if filepath.Ext(path) != ".go" || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		normalizedPath := filepath.ToSlash(strings.TrimPrefix(path, "./"))
		if strings.HasPrefix(normalizedPath, "scripts/") {
			return nil
		}

		contentBytes, readError := os.ReadFile(path)
		if readError != nil {
			return readError
		}
		contentText := string(contentBytes)

		for _, rule := range rules {
			if strings.Contains(contentText, rule.needle) {
				observedDescriptions[rule.description] = true
				if !isAllowedPath(normalizedPath, rule.allowedPrefixes) {
					violations = append(violations, normalizedPath+": "+rule.description)
				}
			}
		}
		return nil
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if len(violations) > 0 {
		fmt.Fprintln(os.Stderr, "app source boundary scan found mismatches:")
		for _, violation := range violations {
			fmt.Fprintln(os.Stderr, " - "+violation)
		}
		os.Exit(1)
	}

	fmt.Println("app source boundary scan passed")
	fmt.Println("observed app behavior:")
	for _, rule := range rules {
		if observedDescriptions[rule.description] {
			fmt.Println(" - " + rule.description)
		}
	}
}

func isAllowedPath(path string, allowedPrefixes []string) bool {
	for _, allowedPrefix := range allowedPrefixes {
		if path == allowedPrefix || strings.HasPrefix(path, allowedPrefix) {
			return true
		}
	}
	return false
}
