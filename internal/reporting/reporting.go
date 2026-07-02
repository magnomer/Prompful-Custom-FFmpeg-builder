// Package reporting defines the output and approval seams that let the same
// backend build logic drive either the Wails GUI or the PromptfulX CLI. The
// GUI implementations emit Wails events and show a native dialog; the CLI
// implementations print to stdout/stderr and prompt (or honor --yes). The
// interfaces carry no GUI types, so the CLI binary does not import Wails.
package reporting

// LReporter receives build status changes and log lines from the backend.
type LReporter interface {
	// LReporterStatusEmit reports a coarse action status (for example
	// "building-ffmpeg", "completed", "failed").
	LReporterStatusEmit(status string)
	// LReporterLogEmit reports one log line at the given level
	// ("info", "warn", "error").
	LReporterLogEmit(level string, message string)
}

// LConfirmer answers the single backend-owned approval gate that must be
// satisfied before an approved action runs. It returns true when the user
// approves. Each implementation owns its own presentation of actionName and
// planHash.
type LConfirmer interface {
	LConfirmerApprovalGet(actionName string, planHash string) (bool, error)
}
