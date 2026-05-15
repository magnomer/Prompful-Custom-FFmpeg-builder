package audit

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Writer struct {
	runId        string
	logDirectory string
	mutex        sync.Mutex
}

type Event struct {
	RunId      string `json:"runId"`
	EventName  string `json:"eventName"`
	ActionName string `json:"actionName,omitempty"`
	PlanHash   string `json:"planHash,omitempty"`
	Level      string `json:"level,omitempty"`
	Message    string `json:"message"`
	CreatedAt  string `json:"createdAt"`
}

func NewWriter(workspaceLogsDirectory string, runId string) (*Writer, error) {
	logDirectory := filepath.Join(workspaceLogsDirectory, runId)
	if err := os.MkdirAll(logDirectory, 0o700); err != nil {
		return nil, err
	}
	fileInfo, err := os.Lstat(logDirectory)
	if err != nil {
		return nil, err
	}
	if fileInfo.Mode()&os.ModeSymlink != 0 {
		return nil, os.ErrPermission
	}
	return &Writer{runId: runId, logDirectory: logDirectory}, nil
}

func (writer *Writer) LogDirectory() string {
	if writer == nil {
		return ""
	}
	return writer.logDirectory
}

func (writer *Writer) WriteEvent(eventName string, actionName string, planHash string, level string, message string) error {
	if writer == nil {
		return nil
	}
	writer.mutex.Lock()
	defer writer.mutex.Unlock()
	event := Event{RunId: writer.runId, EventName: eventName, ActionName: actionName, PlanHash: planHash, Level: level, Message: message, CreatedAt: time.Now().UTC().Format(time.RFC3339)}
	eventBytes, err := json.Marshal(event)
	if err != nil {
		return err
	}
	filePath := filepath.Join(writer.logDirectory, "security-events.jsonl")
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
