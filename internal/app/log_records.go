package app

import (
	"bufio"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	"promptfulcustomffmpegbuilder/internal/workspace"
)

type LocalLogEntry struct {
	Level     string `json:"level"`
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
}

type LocalLogRecord struct {
	RunId             string          `json:"runId"`
	CreatedAt         string          `json:"createdAt"`
	DisplayTime       string          `json:"displayTime"`
	Kind              string          `json:"kind"`
	Status            string          `json:"status"`
	Directory         string          `json:"directory"`
	Entries           []LocalLogEntry `json:"entries"`
	RawText           string          `json:"rawText"`
	ErrorCount        int             `json:"errorCount"`
	WarnCount         int             `json:"warnCount"`
	HasStdoutLog      bool            `json:"hasStdoutLog"`
	HasStderrLog      bool            `json:"hasStderrLog"`
	HasSecurityEvents bool            `json:"hasSecurityEvents"`
}

var localLogRawFileNames = map[string]bool{
	"stdout.log":            true,
	"stderr.log":            true,
	"security-events.jsonl": true,
}

type localAuditEvent struct {
	RunId      string `json:"runId"`
	EventName  string `json:"eventName"`
	ActionName string `json:"actionName"`
	Level      string `json:"level"`
	Message    string `json:"message"`
	CreatedAt  string `json:"createdAt"`
}

func (app *App) ListLocalLogRecords(workspaceDirectory string) ([]LocalLogRecord, error) {
	if workspaceDirectory == "" {
		return []LocalLogRecord{}, nil
	}
	layout := workspace.WorkspaceLayoutFor(workspaceDirectory)
	if err := workspace.CheckRealPathInsideWorkspace(layout.WorkspaceDirectory, layout.WorkspaceDirectory); err != nil {
		return nil, err
	}
	if _, err := os.Stat(layout.LogsDirectory); errors.Is(err, os.ErrNotExist) {
		return []LocalLogRecord{}, nil
	} else if err != nil {
		return nil, err
	}
	if err := workspace.CheckRealPathInsideWorkspace(layout.WorkspaceDirectory, layout.LogsDirectory); err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(layout.LogsDirectory)
	if err != nil {
		return nil, err
	}
	records := []LocalLogRecord{}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		runId := entry.Name()
		if strings.Contains(runId, string(os.PathSeparator)) || strings.Contains(runId, "\x00") {
			continue
		}
		recordDirectory := filepath.Join(layout.LogsDirectory, runId)
		if err := workspace.CheckRealPathInsideWorkspace(layout.WorkspaceDirectory, recordDirectory); err != nil {
			continue
		}
		record, ok := readLocalLogRecord(recordDirectory, runId, false)
		if ok {
			records = append(records, record)
		}
	}
	sort.Slice(records, func(i, j int) bool {
		return records[i].CreatedAt > records[j].CreatedAt
	})
	return records, nil
}

func (app *App) GetLocalLogRecord(workspaceDirectory string, runId string) (LocalLogRecord, error) {
	if workspaceDirectory == "" || runId == "" {
		return LocalLogRecord{}, errors.New("workspace directory and log record id are required")
	}
	if strings.Contains(runId, string(os.PathSeparator)) || strings.Contains(runId, "/") || strings.Contains(runId, "\\") || strings.Contains(runId, "\x00") {
		return LocalLogRecord{}, errors.New("invalid log record id")
	}
	layout := workspace.WorkspaceLayoutFor(workspaceDirectory)
	recordDirectory := filepath.Join(layout.LogsDirectory, runId)
	if err := workspace.CheckRealPathInsideWorkspace(layout.WorkspaceDirectory, recordDirectory); err != nil {
		return LocalLogRecord{}, err
	}
	record, ok := readLocalLogRecord(recordDirectory, runId, true)
	if !ok {
		return LocalLogRecord{}, errors.New("log record was not found")
	}
	return record, nil
}

func (app *App) OpenLocalLogsFolder(workspaceDirectory string) error {
	if workspaceDirectory == "" {
		return errors.New("workspace directory is required")
	}
	layout := workspace.WorkspaceLayoutFor(workspaceDirectory)
	if err := workspace.CheckRealPathInsideWorkspace(layout.WorkspaceDirectory, layout.LogsDirectory); err != nil {
		return err
	}
	if info, err := os.Stat(layout.LogsDirectory); err != nil {
		return err
	} else if !info.IsDir() {
		return errors.New("logs folder is not a directory")
	}
	return openFolder(layout.LogsDirectory)
}

func (app *App) OpenLocalLogRecordFolder(workspaceDirectory string, runId string) error {
	if workspaceDirectory == "" || runId == "" {
		return errors.New("workspace directory and log record id are required")
	}
	if strings.Contains(runId, string(os.PathSeparator)) || strings.Contains(runId, "/") || strings.Contains(runId, "\\") || strings.Contains(runId, "\x00") {
		return errors.New("invalid log record id")
	}
	layout := workspace.WorkspaceLayoutFor(workspaceDirectory)
	recordDirectory := filepath.Join(layout.LogsDirectory, runId)
	if err := workspace.CheckRealPathInsideWorkspace(layout.WorkspaceDirectory, recordDirectory); err != nil {
		return err
	}
	if info, err := os.Stat(recordDirectory); err != nil {
		return err
	} else if !info.IsDir() {
		return errors.New("log record folder is not a directory")
	}
	return openFolder(recordDirectory)
}

func (app *App) OpenLocalLogRecordFile(workspaceDirectory string, runId string, fileName string) error {
	if workspaceDirectory == "" || runId == "" {
		return errors.New("workspace directory and log record id are required")
	}
	if strings.Contains(runId, string(os.PathSeparator)) || strings.Contains(runId, "/") || strings.Contains(runId, "\\") || strings.Contains(runId, "\x00") {
		return errors.New("invalid log record id")
	}
	if !localLogRawFileNames[fileName] {
		return errors.New("invalid log file")
	}
	layout := workspace.WorkspaceLayoutFor(workspaceDirectory)
	filePath := filepath.Join(layout.LogsDirectory, runId, fileName)
	if err := workspace.CheckRealPathInsideWorkspace(layout.WorkspaceDirectory, filePath); err != nil {
		return err
	}
	if info, err := os.Stat(filePath); err != nil {
		return err
	} else if info.IsDir() {
		return errors.New("log file is a directory")
	}
	return openPath(filePath)
}

func openFolder(directory string) error {
	switch runtime.GOOS {
	case "windows":
		return exec.Command("explorer.exe", directory).Start()
	case "darwin":
		return exec.Command("open", directory).Start()
	default:
		return exec.Command("xdg-open", directory).Start()
	}
}

func openPath(path string) error {
	switch runtime.GOOS {
	case "windows":
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", path).Start()
	case "darwin":
		return exec.Command("open", path).Start()
	default:
		return exec.Command("xdg-open", path).Start()
	}
}

func readLocalLogRecord(recordDirectory string, runId string, includeDetails bool) (LocalLogRecord, bool) {
	events := readLocalAuditEvents(filepath.Join(recordDirectory, "security-events.jsonl"))
	hasEvents := hasReadableTextFile(filepath.Join(recordDirectory, "security-events.jsonl"))
	hasStdout := hasReadableTextFile(filepath.Join(recordDirectory, "stdout.log"))
	hasStderr := hasReadableTextFile(filepath.Join(recordDirectory, "stderr.log"))
	if len(events) == 0 && !hasStdout && !hasStderr {
		return LocalLogRecord{}, false
	}
	stdoutText := ""
	stderrText := ""
	if includeDetails {
		stdoutText = readSmallTextFile(filepath.Join(recordDirectory, "stdout.log"), 256*1024)
		stderrText = readSmallTextFile(filepath.Join(recordDirectory, "stderr.log"), 256*1024)
	}

	record := LocalLogRecord{RunId: runId, Directory: recordDirectory, Kind: "unknown", Status: "unknown", Entries: []LocalLogEntry{}, HasStdoutLog: hasStdout, HasStderrLog: hasStderr, HasSecurityEvents: hasEvents}
	if parsedTime, err := time.Parse("20060102T150405Z", runId); err == nil {
		record.CreatedAt = parsedTime.Format(time.RFC3339)
		record.DisplayTime = parsedTime.Local().Format("2006-01-02 15:04")
	} else {
		record.CreatedAt = runId
		record.DisplayTime = runId
	}

	for _, event := range events {
		if event.RunId != "" {
			record.RunId = event.RunId
		}
		if event.CreatedAt != "" {
			record.CreatedAt = firstNonEmpty(record.CreatedAt, event.CreatedAt)
		}
		if record.Kind == "unknown" {
			record.Kind = inferLogKind(event.ActionName, event.Message)
		}
		if event.Level == "error" {
			record.ErrorCount++
		}
		if event.Level == "warn" {
			record.WarnCount++
		}
		if event.EventName == "action-completed" {
			record.Status = "completed"
		}
		if event.EventName == "action-failed" || event.Level == "error" && record.Status != "completed" {
			record.Status = "failed"
		}
		level := event.Level
		if level == "" {
			level = "info"
		}
		timestamp := event.CreatedAt
		if parsedTime, err := time.Parse(time.RFC3339, event.CreatedAt); err == nil {
			timestamp = parsedTime.Local().Format("15:04:05")
		}
		if includeDetails {
			record.Entries = append(record.Entries, LocalLogEntry{Level: level, Message: event.Message, Timestamp: timestamp})
		}
	}
	if record.Kind == "unknown" {
		record.Kind = inferLogKind("", readSmallTextFile(filepath.Join(recordDirectory, "stdout.log"), 16*1024)+"\n"+readSmallTextFile(filepath.Join(recordDirectory, "stderr.log"), 16*1024))
	}
	if record.Status == "unknown" && (len(events) > 0 || hasStdout || hasStderr) {
		record.Status = "recorded"
	}
	if includeDetails {
		record.RawText = buildRawLogText(events, stdoutText, stderrText)
	}
	return record, true
}

func readLocalAuditEvents(path string) []localAuditEvent {
	file, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	events := []localAuditEvent{}
	for scanner.Scan() {
		var event localAuditEvent
		if err := json.Unmarshal(scanner.Bytes(), &event); err == nil {
			events = append(events, event)
		}
	}
	return events
}

func readSmallTextFile(path string, limit int64) string {
	fileInfo, err := os.Stat(path)
	if err != nil || fileInfo.IsDir() || fileInfo.Size() == 0 {
		return ""
	}
	if fileInfo.Size() > limit {
		file, err := os.Open(path)
		if err != nil {
			return ""
		}
		defer file.Close()
		data := make([]byte, limit)
		readCount, _ := file.Read(data)
		return string(data[:readCount]) + "\n\n[Log truncated in the UI. Open the log folder to inspect the full file.]"
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return string(data)
}

func hasReadableTextFile(path string) bool {
	fileInfo, err := os.Stat(path)
	return err == nil && !fileInfo.IsDir() && fileInfo.Size() > 0
}

func buildRawLogText(events []localAuditEvent, stdoutText string, stderrText string) string {
	sections := []string{}
	if len(events) > 0 {
		lines := []string{"security-events.jsonl"}
		for _, event := range events {
			eventBytes, err := json.Marshal(event)
			if err == nil {
				lines = append(lines, string(eventBytes))
			}
		}
		sections = append(sections, strings.Join(lines, "\n"))
	}
	if stdoutText != "" {
		sections = append(sections, "stdout.log\n"+stdoutText)
	}
	if stderrText != "" {
		sections = append(sections, "stderr.log\n"+stderrText)
	}
	return strings.Join(sections, "\n\n")
}

func inferLogKind(actionName string, text string) string {
	value := strings.ToLower(actionName + "\n" + text)
	if strings.Contains(value, "build-ffmpeg") || strings.Contains(value, "ffmpeg build") || strings.Contains(value, "ffmpeg configure") || strings.Contains(value, "ffmpeg make") {
		return "ffmpeg"
	}
	if strings.Contains(value, "prepare-private-msys2") || strings.Contains(value, "toolchain") || strings.Contains(value, "msys2") || strings.Contains(value, "pacman package installation") {
		return "toolchain"
	}
	return "unknown"
}

func firstNonEmpty(current string, next string) string {
	if current != "" {
		return current
	}
	return next
}
