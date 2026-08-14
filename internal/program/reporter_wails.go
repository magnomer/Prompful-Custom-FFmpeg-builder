package program

import (
	"errors"

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

// LConfirmerWails is the GUI reporting.LConfirmer implementation. It presents
// the backend-owned approval as a localized native question dialog.
type LConfirmerWails struct {
	program *LProgram
}

func (confirmer LConfirmerWails) LConfirmerApprovalGet(actionName string, planHash string) (bool, error) {
	program := confirmer.program
	if program.LContext == nil {
		return false, errors.New("program context is not ready for native approval dialog")
	}
	locale := program.lLocaleCurrentGet()
	message := LLocaleTextForGet(locale, "native.approval.message", map[string]string{"action": LLocaleTextForGet(locale, "approval.action."+actionName, nil), "planHash": planHash})
	noButtonLabel := LLocaleTextForGet(locale, "native.approval.no", nil)
	yesButtonLabel := LLocaleTextForGet(locale, "native.approval.yes", nil)
	choice, err := wailsRuntime.MessageDialog(program.LContext, wailsRuntime.MessageDialogOptions{
		Type:          wailsRuntime.QuestionDialog,
		Title:         LLocaleTextForGet(locale, "native.approval.title", nil),
		Message:       message,
		Buttons:       []string{noButtonLabel, yesButtonLabel},
		DefaultButton: noButtonLabel,
		CancelButton:  noButtonLabel,
	})
	if err != nil {
		return false, err
	}
	// On Windows, Wails' QuestionDialog ignores custom button labels and returns
	// the native "Yes"/"No" strings, so a localized yes label would never match.
	// Accept the localized label or the native English "Yes".
	return choice == yesButtonLabel || choice == "Yes", nil
}
