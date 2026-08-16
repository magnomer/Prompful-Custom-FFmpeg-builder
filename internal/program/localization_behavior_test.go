package program

import "testing"

func TestToolchainVerificationReturnsLocalizationKey(t *testing.T) {
	program := LProgramCreate()
	verification, err := program.LToolchainInstallVerify(t.TempDir(), "ucrt64")
	if err != nil {
		t.Fatalf("verify toolchain: %v", err)
	}
	if verification.MessageKey != "verify.toolchain.notInstalled" {
		t.Fatalf("message key = %q", verification.MessageKey)
	}
	if verification.Message != "" {
		t.Fatalf("unexpected frozen message %q", verification.Message)
	}
}

func TestBuildVerificationReturnsLocalizationKey(t *testing.T) {
	program := LProgramCreate()
	verification, err := program.LVerificationBuildRun(t.TempDir())
	if err != nil {
		t.Fatalf("verify build: %v", err)
	}
	if verification.MessageKey != "verify.build.ffmpegNotFound" {
		t.Fatalf("message key = %q", verification.MessageKey)
	}
	if verification.Message != "" {
		t.Fatalf("unexpected frozen message %q", verification.Message)
	}
}

func TestApprovalConfirmationResolveIsSingleUse(t *testing.T) {
	program := LProgramCreate()
	response := make(chan bool, 1)
	program.LConfirmationRequestId = "request-1"
	program.LConfirmationResponse = response

	if err := program.LApprovalConfirmationResolve("request-1", true); err != nil {
		t.Fatalf("resolve confirmation: %v", err)
	}
	if approved := <-response; !approved {
		t.Fatal("expected approved response")
	}
	if err := program.LApprovalConfirmationResolve("request-1", true); err == nil {
		t.Fatal("expected a consumed confirmation to be rejected")
	}
}
