package execution

import (
	"errors"
	"strings"
	"sync/atomic"
	"testing"

	"promptfulcustomffmpegbuilder/internal/scripting"
)

func TestLLogLineClassifyDemotesGarbledStripNoSectionLine(t *testing.T) {
	line := "D:/tempc/build/prep/xavs2/source/common/x86/mc-a2.asm:402: warning: improperly calling multi-lD:\\tempc\\toolchains\\msys2-ucrt64\\ucrt64\\bin\\strip.exe:ine macr o `SETUP_STACK_POINTER' with 0error: the input file 'common/x86/pixel-32.o' has no section spa"
	if got := LLogLineGet("warn", line); got != "info" {
		t.Fatalf("expected strip no-section noise to be info, got %q", got)
	}
}

func TestLLogLineClassifyPromotesArgumentListTooLong(t *testing.T) {
	line := "/bin/sh: line 1: /ucrt64/bin/gcc: Argument list too long"
	if got := LLogLineGet("warn", line); got != "error" {
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
		nil,
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

func TestLNetworkHostParseExtractsStalledHost(t *testing.T) {
	line := "error: failed retrieving file 'zlib-1.3.1-1-any.pkg.tar.zst' from repo.msys2.org : Operation too slow"
	if got := LNetworkHostParse(line); got != "repo.msys2.org" {
		t.Fatalf("expected host repo.msys2.org, got %q", got)
	}
	if got := LNetworkHostParse("gcc: error: undefined reference to foo"); got != "" {
		t.Fatalf("expected no host from a compile error, got %q", got)
	}
}

func TestLNetworkStalledCreateMergesCatalogAndParsedHosts(t *testing.T) {
	collector := &LNetworkAddressCollector{}
	collector.LNetworkHostAdd("mirror.example.net")
	collector.LNetworkHostAdd("mirror.example.net") // repeat is deduped

	cause := errors.New("exit status 1")
	stalledErr := LNetworkStalledCreate(cause, collector)

	var stalled *LErrorNetworkStalled
	if !errors.As(stalledErr, &stalled) {
		t.Fatalf("expected an LErrorNetworkStalled, got %T", stalledErr)
	}
	if !errors.Is(stalledErr, cause) {
		t.Fatal("expected the stalled error to unwrap to its cause")
	}
	// Every authored mirror base is listed first, in order.
	for index, mirror := range scripting.LMSYSMirrorCatalog {
		if stalled.LNetworkAddresses[index] != mirror {
			t.Fatalf("expected authored mirror %q at position %d, got %q", mirror, index, stalled.LNetworkAddresses[index])
		}
	}
	joined := strings.Join(stalled.LNetworkAddresses, ",")
	if strings.Count(joined, "mirror.example.net") != 1 {
		t.Fatalf("expected the parsed host once, got %q", joined)
	}
}
