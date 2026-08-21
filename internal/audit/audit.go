package audit

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type LAuditWriter struct {
	LIdentifierRun       string
	LIdentifierReviewRun string
	LDirectoryLog        string
	LMutex               sync.Mutex
}

type LAuditEvent struct {
	RunId           string `json:"runId"`
	ReviewSessionId string `json:"reviewSessionId,omitempty"`
	LAuditEventName string `json:"eventName"`
	ActionName      string `json:"actionName,omitempty"`
	PlanHash        string `json:"planHash,omitempty"`
	Level           string `json:"level,omitempty"`
	Message         string `json:"message"`
	CreatedAt       string `json:"createdAt"`
}

func LAuditWriterCreate(workspaceLogsDirectory string, LRunId string, reviewSessionId string) (*LAuditWriter, error) {
	LDirectoryLog := filepath.Join(workspaceLogsDirectory, LRunId)
	if err := os.MkdirAll(LDirectoryLog, 0o700); err != nil {
		return nil, err
	}
	fileInfo, err := os.Lstat(LDirectoryLog)
	if err != nil {
		return nil, err
	}
	if fileInfo.Mode()&os.ModeSymlink != 0 {
		return nil, os.ErrPermission
	}
	return &LAuditWriter{LIdentifierRun: LRunId, LIdentifierReviewRun: reviewSessionId, LDirectoryLog: LDirectoryLog}, nil
}

func (writer *LAuditWriter) LAuditDirectoryGet() string {
	if writer == nil {
		return ""
	}
	return writer.LDirectoryLog
}

func (writer *LAuditWriter) LAuditEventWrite(eventName string, actionName string, planHash string, level string, message string) error {
	if writer == nil {
		return nil
	}
	writer.LMutex.Lock()
	defer writer.LMutex.Unlock()
	event := LAuditEvent{RunId: writer.LIdentifierRun, ReviewSessionId: writer.LIdentifierReviewRun, LAuditEventName: eventName, ActionName: actionName, PlanHash: planHash, Level: level, Message: message, CreatedAt: time.Now().UTC().Format(time.RFC3339)}
	eventBytes, err := json.Marshal(event)
	if err != nil {
		return err
	}
	filePath := filepath.Join(writer.LDirectoryLog, "security-events.jsonl")
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	defer file.Close()
	if _, err := file.Write(append(eventBytes, '\n')); err != nil {
		return err
	}
	return nil
}
