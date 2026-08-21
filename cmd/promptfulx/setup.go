package main

import (
	"fmt"
	"os"
	"strings"

	"promptfulcustomffmpegbuilder/internal/consent"
	"promptfulcustomffmpegbuilder/internal/planning"
	"promptfulcustomffmpegbuilder/internal/program"
)

// LCommandSetupRun prepares the MSYS2 build environment (toolchain) in a workspace so a
// later `build` can run. It runs synchronously and maps the outcome to an exit
// code. Reuses LReporterBuild to capture the final status.
func LCommandSetupRun(args []string) int {
	parsed, err := LArgumentParse(args)
	if err != nil {
		return LCommandExitGet(err)
	}
	settings, err := LSettingsToolchainResolve(parsed)
	if err != nil {
		return LCommandExitGet(err)
	}

	driver := program.LProgramCreate()
	reporter := &LReporterBuild{}
	driver.LReporter = reporter
	driver.LConfirmer = LConfirmerConsole{assumeYes: parsed.yes, noInput: parsed.noInput}

	review, err := driver.LPlanToolchainRequest(settings)
	if err != nil {
		return LCommandExitGet(LErrorSupportCreate("could not resolve setup plan: %v", err))
	}
	LCommandToolchainPrint(review.Plan)
	if !review.Plan.IsExecutable {
		fmt.Fprintln(os.Stderr, "promptfulx: setup plan is blocked by the warnings above; cannot prepare")
		return 4
	}

	approval := consent.LRequestApproval{
		ApprovedActionName: review.Plan.ActionName,
		ApprovedPlanHash:   review.Plan.PlanHash,
		LConsentText:       review.ExpectedLConsentText,
	}
	stopSignalCancel := lActionSignalCancel(driver)
	defer stopSignalCancel()
	if _, err := driver.LToolchainApproveSync(review.ReviewSessionId, approval); err != nil {
		fmt.Fprintln(os.Stderr, "promptfulx:", err.Error())
		if strings.Contains(err.Error(), "rejected") {
			return 10 // user cancelled at the confirmation gate
		}
		return 6 // setup failure
	}

	if reporter.status == "completed" {
		return 0
	}
	return 6
}

func LCommandToolchainPrint(plan planning.LPlanToolchain) {
	fmt.Println("MSYS2 toolchain setup plan")
	fmt.Println("  workspace:  ", plan.WorkspaceDirectory)
	fmt.Println("  profile:    ", plan.WindowsShellProfileName)
	fmt.Println("  msys2:      ", plan.Msys2ArchiveUrl)
	fmt.Println("  msys2 root: ", plan.Msys2RootDirectory)
	fmt.Println("  reuse msys2:", plan.WillUseExistingMsys2)
	fmt.Println("  packages:   ", len(plan.Msys2PackageNames))
	fmt.Println("  executable: ", plan.IsExecutable)

	if len(plan.Warnings) > 0 {
		fmt.Printf("\nWarnings (%d):\n", len(plan.Warnings))
		for _, warning := range plan.Warnings {
			fmt.Printf("  [%s] %s\n", warning.LRiskLevel, warning.Message)
		}
	}
}
