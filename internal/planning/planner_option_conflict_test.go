package planning

import "testing"

func TestLOptionConflictValidateAllowsSharedAndLto(t *testing.T) {
	warnings, blocked := LOptionConflictValidate([]string{"--enable-shared", "--enable-lto"})
	if blocked {
		t.Fatal("did not expect shared + lto to be blocked")
	}
	for _, planWarning := range warnings {
		if planWarning.LRiskLevel == LRiskBlocked {
			t.Fatalf("unexpected blocking warning: %s", planWarning.MessageKey)
		}
	}
}
