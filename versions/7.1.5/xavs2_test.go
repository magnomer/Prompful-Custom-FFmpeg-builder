package version715

import (
	"testing"

	"promptfulcustomffmpegbuilder/versions/shared"
)

func TestLLibraryXavs2PrepareDisablesStripDuringMake(t *testing.T) {
	plan := shared.LPreparationPlanCreate("7.1.5", "xavs2", "versions/7.1.5/xavs2.go")
	LXavs2Prepare(plan)

	for _, makeVariable := range plan.MakeVariables {
		if makeVariable == "STRIP=" {
			return
		}
	}
	t.Fatalf("expected xavs2 make variables to include STRIP=, got %#v", plan.MakeVariables)
}
