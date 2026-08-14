package program

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

type LLogLocalEntry struct {
	Level     string `json:"level"`
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
}

type LRecordLog struct {
	RunId                   string           `json:"runId"`
	CreatedAt               string           `json:"createdAt"`
	DisplayTime             string           `json:"displayTime"`
	Kind                    string           `json:"kind"`
	Status                  string           `json:"status"`
	Directory               string           `json:"directory"`
	Entries                 []LLogLocalEntry `json:"entries"`
	RawText                 string           `json:"rawText"`
	ErrorCount              int              `json:"errorCount"`
	WarnCount               int              `json:"warnCount"`
	HasStdoutLog            bool             `json:"hasStdoutLog"`
	HasStderrLog            bool             `json:"hasStderrLog"`
	HasSecurityLAuditEvents bool             `json:"hasSecurityLAuditEvents"`
}

var LLogRawNames = map[string]bool{
	"stdout.log":            true,
	"stderr.log":            true,
	"security-events.jsonl": true,
}

type LAuditLocalEvent struct {
	RunId           string `json:"runId"`
	LAuditEventName string `json:"eventName"`
	ActionName      string `json:"actionName"`
	Level           string `json:"level"`
	Message         string `json:"message"`
	CreatedAt       string `json:"createdAt"`
}

func (program *LProgram) LLogRecordList(workspaceDirectory string) ([]LRecordLog, error) {
	if workspaceDirectory == "" {
		return []LRecordLog{}, nil
	}
	layout := workspace.LWorkspaceLayoutResolve(workspaceDirectory)
	if err := workspace.LPathRealCheck(layout.WorkspaceDirectory, layout.WorkspaceDirectory); err != nil {
		return nil, err
	}
	if _, err := os.Stat(layout.LogsDirectory); errors.Is(err, os.ErrNotExist) {
		return []LRecordLog{}, nil
	} else if err != nil {
		return nil, err
	}
	if err := workspace.LPathRealCheck(layout.WorkspaceDirectory, layout.LogsDirectory); err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(layout.LogsDirectory)
	if err != nil {
		return nil, err
	}
	records := []LRecordLog{}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		LRunId := entry.Name()
		if strings.Contains(LRunId, string(os.PathSeparator)) || strings.Contains(LRunId, "\x00") {
			continue
		}
		recordDirectory := filepath.Join(layout.LogsDirectory, LRunId)
		if err := workspace.LPathRealCheck(layout.WorkspaceDirectory, recordDirectory); err != nil {
			continue
		}
		record, ok := LRecordLocalRead(recordDirectory, LRunId, false)
		if ok {
			records = append(records, record)
		}
	}
	sort.Slice(records, func(i, j int) bool {
		return records[i].CreatedAt > records[j].CreatedAt
	})
	return records, nil
}

func (program *LProgram) LRecordLogGet(workspaceDirectory string, LRunId string) (LRecordLog, error) {
	if workspaceDirectory == "" || LRunId == "" {
		return LRecordLog{}, errors.New("workspace directory and log record id are required")
	}
	if strings.Contains(LRunId, string(os.PathSeparator)) || strings.Contains(LRunId, "/") || strings.Contains(LRunId, "\\") || strings.Contains(LRunId, "\x00") {
		return LRecordLog{}, errors.New("invalid log record id")
	}
	layout := workspace.LWorkspaceLayoutResolve(workspaceDirectory)
	recordDirectory := filepath.Join(layout.LogsDirectory, LRunId)
	if err := workspace.LPathRealCheck(layout.WorkspaceDirectory, recordDirectory); err != nil {
		return LRecordLog{}, err
	}
	record, ok := LRecordLocalRead(recordDirectory, LRunId, true)
	if !ok {
		return LRecordLog{}, errors.New("log record was not found")
	}
	return record, nil
}

func (program *LProgram) LFolderLogOpen(workspaceDirectory string) error {
	if workspaceDirectory == "" {
		return errors.New("workspace directory is required")
	}
	layout := workspace.LWorkspaceLayoutResolve(workspaceDirectory)
	if err := workspace.LPathRealCheck(layout.WorkspaceDirectory, layout.LogsDirectory); err != nil {
		return err
	}
	if info, err := os.Stat(layout.LogsDirectory); err != nil {
		return err
	} else if !info.IsDir() {
		return errors.New("logs folder is not a directory")
	}
	return LDirectoryOpen(layout.LogsDirectory)
}

func (program *LProgram) LFolderRecordOpen(workspaceDirectory string, LRunId string) error {
	if workspaceDirectory == "" || LRunId == "" {
		return errors.New("workspace directory and log record id are required")
	}
	if strings.Contains(LRunId, string(os.PathSeparator)) || strings.Contains(LRunId, "/") || strings.Contains(LRunId, "\\") || strings.Contains(LRunId, "\x00") {
		return errors.New("invalid log record id")
	}
	layout := workspace.LWorkspaceLayoutResolve(workspaceDirectory)
	recordDirectory := filepath.Join(layout.LogsDirectory, LRunId)
	if err := workspace.LPathRealCheck(layout.WorkspaceDirectory, recordDirectory); err != nil {
		return err
	}
	if info, err := os.Stat(recordDirectory); err != nil {
		return err
	} else if !info.IsDir() {
		return errors.New("log record folder is not a directory")
	}
	return LDirectoryOpen(recordDirectory)
}

func (program *LProgram) LFileRecordOpen(workspaceDirectory string, LRunId string, fileName string) error {
	if workspaceDirectory == "" || LRunId == "" {
		return errors.New("workspace directory and log record id are required")
	}
	if strings.Contains(LRunId, string(os.PathSeparator)) || strings.Contains(LRunId, "/") || strings.Contains(LRunId, "\\") || strings.Contains(LRunId, "\x00") {
		return errors.New("invalid log record id")
	}
	if !LLogRawNames[fileName] {
		return errors.New("invalid log file")
	}
	layout := workspace.LWorkspaceLayoutResolve(workspaceDirectory)
	filePath := filepath.Join(layout.LogsDirectory, LRunId, fileName)
	if err := workspace.LPathRealCheck(layout.WorkspaceDirectory, filePath); err != nil {
		return err
	}
	if info, err := os.Stat(filePath); err != nil {
		return err
	} else if info.IsDir() {
		return errors.New("log file is a directory")
	}
	return LPathOpen(filePath)
}

func LDirectoryOpen(directory string) error {
	switch runtime.GOOS {
	case "windows":
		return exec.Command("explorer.exe", directory).Start()
	case "darwin":
		return exec.Command("open", directory).Start()
	default:
		return exec.Command("xdg-open", directory).Start()
	}
}

func LPathOpen(path string) error {
	switch runtime.GOOS {
	case "windows":
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", path).Start()
	case "darwin":
		return exec.Command("open", path).Start()
	default:
		return exec.Command("xdg-open", path).Start()
	}
}

func LRecordLocalRead(recordDirectory string, LRunId string, includeDetails bool) (LRecordLog, bool) {
	events := LAuditLocalRead(filepath.Join(recordDirectory, "security-events.jsonl"))
	hasLAuditEvents := LTextReadableCheck(filepath.Join(recordDirectory, "security-events.jsonl"))
	hasStdout := LTextReadableCheck(filepath.Join(recordDirectory, "stdout.log"))
	hasStderr := LTextReadableCheck(filepath.Join(recordDirectory, "stderr.log"))
	if len(events) == 0 && !hasStdout && !hasStderr {
		return LRecordLog{}, false
	}
	stdoutText := ""
	stderrText := ""
	if includeDetails {
		stdoutText = LTextSmallRead(filepath.Join(recordDirectory, "stdout.log"), 256*1024)
		stderrText = LTextSmallRead(filepath.Join(recordDirectory, "stderr.log"), 256*1024)
	}

	record := LRecordLog{RunId: LRunId, Directory: recordDirectory, Kind: "unknown", Status: "unknown", Entries: []LLogLocalEntry{}, HasStdoutLog: hasStdout, HasStderrLog: hasStderr, HasSecurityLAuditEvents: hasLAuditEvents}
	if parsedTime, err := time.Parse("20060102T150405Z", LRunId); err == nil {
		record.CreatedAt = parsedTime.Format(time.RFC3339)
		record.DisplayTime = parsedTime.Local().Format("2006-01-02 15:04")
	} else {
		record.CreatedAt = LRunId
		record.DisplayTime = LRunId
	}

	for _, event := range events {
		if event.RunId != "" {
			record.RunId = event.RunId
		}
		if event.CreatedAt != "" {
			record.CreatedAt = LTextFirstGet(record.CreatedAt, event.CreatedAt)
		}
		if record.Kind == "unknown" {
			record.Kind = LLogKindGet(event.ActionName, event.Message)
		}
		if event.Level == "error" {
			record.ErrorCount++
		}
		if event.Level == "warn" {
			record.WarnCount++
		}
		if event.LAuditEventName == "action-completed" {
			record.Status = "completed"
		}
		// A transient-network stall halts the run in its own retryable state. Its
		// terminal event is written at warn level (so it does not raise ErrorCount)
		// and is the last event in the record, so it overrides any "failed" that
		// earlier error-level output set.
		if event.LAuditEventName == "action-stalled" {
			record.Status = "stalled"
		}
		if event.LAuditEventName == "action-failed" || event.Level == "error" && record.Status != "completed" && record.Status != "stalled" {
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
			record.Entries = append(record.Entries, LLogLocalEntry{Level: level, Message: event.Message, Timestamp: timestamp})
		}
	}
	if record.Kind == "unknown" {
		record.Kind = LLogKindGet("", LTextSmallRead(filepath.Join(recordDirectory, "stdout.log"), 16*1024)+"\n"+LTextSmallRead(filepath.Join(recordDirectory, "stderr.log"), 16*1024))
	}
	if record.Status == "unknown" && (len(events) > 0 || hasStdout || hasStderr) {
		record.Status = "recorded"
	}
	if includeDetails {
		record.RawText = LLogRawCreate(events, stdoutText, stderrText)
	}
	return record, true
}

func LAuditLocalRead(path string) []LAuditLocalEvent {
	file, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	events := []LAuditLocalEvent{}
	for scanner.Scan() {
		var event LAuditLocalEvent
		if err := json.Unmarshal(scanner.Bytes(), &event); err == nil {
			events = append(events, event)
		}
	}
	_ = scanner.Err()
	return events
}

func LTextSmallRead(path string, limit int64) string {
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

func LTextReadableCheck(path string) bool {
	fileInfo, err := os.Stat(path)
	return err == nil && !fileInfo.IsDir() && fileInfo.Size() > 0
}

func LLogRawCreate(events []LAuditLocalEvent, stdoutText string, stderrText string) string {
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

func LLogKindGet(actionName string, text string) string {
	value := strings.ToLower(actionName + "\n" + text)
	if strings.Contains(value, "build-ffmpeg") || strings.Contains(value, "ffmpeg build") || strings.Contains(value, "ffmpeg configure") || strings.Contains(value, "ffmpeg make") {
		return "ffmpeg"
	}
	if strings.Contains(value, "prepare-private-msys2") || strings.Contains(value, "toolchain") || strings.Contains(value, "msys2") || strings.Contains(value, "pacman package installation") {
		return "toolchain"
	}
	return "unknown"
}

func LTextFirstGet(current string, next string) string {
	if current != "" {
		return current
	}
	return next
}
