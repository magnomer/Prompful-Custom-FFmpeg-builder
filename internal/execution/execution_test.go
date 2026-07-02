package execution

import (
	"strings"
	"sync/atomic"
	"testing"
)

func TestLLogLineClassifyDemotesGarbledStripNoSectionLine(t *testing.T) {
	line := "D:/tempc/build/prep/xavs2/source/common/x86/mc-a2.asm:402: warning: improperly calling multi-lD:\\tempc\\toolchains\\msys2-ucrt64\\ucrt64\\bin\\strip.exe:ine macr o `SETUP_STACK_POINTER' with 0error: the input file 'common/x86/pixel-32.o' has no section spa"
	if got := LLogLineClassify("warn", line); got != "info" {
		t.Fatalf("expected strip no-section noise to be info, got %q", got)
	}
}

func TestLLogLineClassifyPromotesArgumentListTooLong(t *testing.T) {
	line := "/bin/sh: line 1: /ucrt64/bin/gcc: Argument list too long"
	if got := LLogLineClassify("warn", line); got != "error" {
		t.Fatalf("expected argument-list overflow to be error, got %q", got)
	}
}

func TestLLogCommandCopyKeepsSpecificErrorBeforeGenericMakeFailure(t *testing.T) {
	var lastErrorLine atomic.Pointer[string]
	doneChannel := make(chan struct{}, 1)
	LLogCommandCopy(
		strings.NewReader("/bin/sh: line 1: /ucrt64/bin/gcc: Argument list too long\nmake: *** [ffbuild/library.mak:119: libavcodec/avcodec-61.dll] Error 126\n"),
		nil,
		"warn",
		nil,
		nil,
		&lastErrorLine,
		doneChannel,
	)
	<-doneChannel

	got := lastErrorLine.Load()
	if got == nil {
		t.Fatal("expected an error line to be captured")
	}
	want := "/bin/sh: line 1: /ucrt64/bin/gcc: Argument list too long"
	if *got != want {
		t.Fatalf("expected specific failure %q, got %q", want, *got)
	}
}
