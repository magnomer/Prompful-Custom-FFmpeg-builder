package main

import (
	"fmt"
	"os"
	"strings"

	"promptfulcustomffmpegbuilder/internal/consent"
	"promptfulcustomffmpegbuilder/internal/program"
)

// buildReporter wraps the console reporter and records the last status so the
// build command can map the final outcome to an exit code after the synchronous
// build returns. A pointer receiver keeps the captured status.
type buildReporter struct {
	console LReporterConsole
	status  string
}

func (r *buildReporter) LReporterStatusEmit(status string) {
	r.status = status
	r.console.LReporterStatusEmit(status)
}

func (r *buildReporter) LReporterLogEmit(level string, message string) {
	r.console.LReporterLogEmit(level, message)
}

// cmdBuild resolves the plan, confirms, and runs the build synchronously,
// mapping the outcome to an exit code. It assumes the workspace's MSYS2
// toolchain is already prepared (there is no CLI `setup` yet); an unprepared
// workspace fails the toolchain-readiness check with a clear message.
func cmdBuild(args []string) int {
	parsed, err := argsParse(args)
	if err != nil {
		return exitFor(err)
	}
	settings, err := settingsResolve(parsed)
	if err != nil {
		return exitFor(err)
	}

	driver := program.LProgramCreate()
	reporter := &buildReporter{}
	driver.LReporter = reporter
	driver.LConfirmer = LConfirmerConsole{assumeYes: parsed.yes, noInput: parsed.noInput}

	review, err := driver.LPlanFFmpegRequest(settings)
	if err != nil {
		return exitFor(unsupported("could not resolve plan: %v", err))
	}
	printPlan(review.Plan)
	if !review.Plan.IsExecutable {
		fmt.Fprintln(os.Stderr, "promptfulx: plan is blocked by the warnings above; cannot build")
		return 4
	}

	approval := consent.LRequestApproval{
		ApprovedActionName: review.Plan.ActionName,
		ApprovedPlanHash:   review.Plan.PlanHash,
		LConsentText:       review.ExpectedLConsentText,
	}
	if _, err := driver.LPlanFFmpegApproveSync(review.ReviewSessionId, approval); err != nil {
		fmt.Fprintln(os.Stderr, "promptfulx:", err.Error())
		if strings.Contains(err.Error(), "rejected") {
			return 10 // user cancelled at the confirmation gate
		}
		return 1
	}

	if reporter.status == "completed" {
		return 0
	}
	return 8 // build ran but did not complete successfully
}
