package program

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// LReporterWails is the GUI reporting.LReporter implementation. It emits the
// Wails events the frontend listens for. It reads program.LContext lazily
// because the context only becomes valid after LProgramStart.
type LReporterWails struct {
	program *LProgram
}

func (reporter LReporterWails) LReporterStatusEmit(status string) {
	if reporter.program.LContext != nil {
		wailsRuntime.EventsEmit(reporter.program.LContext, "approved-action-status", map[string]string{"status": status})
	}
}

func (reporter LReporterWails) LReporterLogEmit(level string, message string) {
	if reporter.program.LContext != nil {
		wailsRuntime.EventsEmit(reporter.program.LContext, "security-log", map[string]string{"level": level, "message": message})
	}
}

func (reporter LReporterWails) LReporterStalledEmit(addresses []string) {
	if reporter.program.LContext != nil {
		wailsRuntime.EventsEmit(reporter.program.LContext, "approved-action-stalled", map[string][]string{"addresses": addresses})
	}
}

// LConfirmerWails is the GUI reporting.LConfirmer implementation. The backend
// owns a one-time request and waits for the application-localized modal to
// resolve it. This avoids Windows MessageBox buttons following the OS language
// instead of the application's selected language.
type LConfirmerWails struct {
	program *LProgram
}

func (confirmer LConfirmerWails) LConfirmerApprovalGet(actionName string, planHash string) (bool, error) {
	program := confirmer.program
	if program.LContext == nil {
		return false, errors.New("program context is not ready for approval confirmation")
	}
	requestBytes := make([]byte, 16)
	if _, err := rand.Read(requestBytes); err != nil {
		return false, err
	}
	requestId := hex.EncodeToString(requestBytes)
	response := make(chan bool, 1)
	program.LMutexConfirmation.Lock()
	if program.LConfirmationRequestId != "" {
		program.LMutexConfirmation.Unlock()
		return false, errors.New("another approval confirmation is already pending")
	}
	program.LConfirmationRequestId = requestId
	program.LConfirmationResponse = response
	program.LMutexConfirmation.Unlock()

	defer program.lApprovalConfirmationClear(requestId)
	wailsRuntime.EventsEmit(program.LContext, "approval-confirmation-request", map[string]string{
		"requestId":  requestId,
		"actionName": actionName,
		"planHash":   planHash,
	})

	select {
	case approved := <-response:
		return approved, nil
	case <-program.LContext.Done():
		return false, program.LContext.Err()
	case <-time.After(5 * time.Minute):
		return false, errors.New("approval confirmation timed out")
	}
}

func (program *LProgram) LApprovalConfirmationResolve(requestId string, approved bool) error {
	program.LMutexConfirmation.Lock()
	if requestId == "" || requestId != program.LConfirmationRequestId || program.LConfirmationResponse == nil {
		program.LMutexConfirmation.Unlock()
		return errors.New("approval confirmation request is not pending")
	}
	response := program.LConfirmationResponse
	program.LConfirmationRequestId = ""
	program.LConfirmationResponse = nil
	program.LMutexConfirmation.Unlock()
	response <- approved
	return nil
}

func (program *LProgram) lApprovalConfirmationClear(requestId string) {
	program.LMutexConfirmation.Lock()
	defer program.LMutexConfirmation.Unlock()
	if program.LConfirmationRequestId == requestId {
		program.LConfirmationRequestId = ""
		program.LConfirmationResponse = nil
	}
	if program.LContext != nil {
		wailsRuntime.EventsEmit(program.LContext, "approval-confirmation-closed", map[string]string{"requestId": requestId})
	}
}

func (program *LProgram) lApprovalConfirmationCancel() {
	program.LMutexConfirmation.Lock()
	if program.LConfirmationResponse == nil {
		program.LMutexConfirmation.Unlock()
		return
	}
	response := program.LConfirmationResponse
	program.LConfirmationRequestId = ""
	program.LConfirmationResponse = nil
	program.LMutexConfirmation.Unlock()
	response <- false
}
