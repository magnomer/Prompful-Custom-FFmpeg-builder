package execution

import (
	"fmt"
	"regexp"
	"strings"
	"sync"

	"promptfulcustomffmpegbuilder/internal/scripting"
)

// LErrorNetworkStalled marks a command that exhausted its retry budget while
// still hitting transient network failures. It is distinct from a genuine
// build/config error so the run can halt in a retryable "stalled" state rather
// than fail outright, and it records which addresses were tried (the authored
// mirror bases plus any specific hosts parsed from the streamed failures). See
// the resume rationale on LCommandAttemptMax above.
type LErrorNetworkStalled struct {
	LNetworkAddresses []string
	LNetworkCause     error
}

func (stalled *LErrorNetworkStalled) Error() string {
	if len(stalled.LNetworkAddresses) == 0 {
		return fmt.Sprintf("network stalled after %d attempts: %v", LCommandAttemptMax, stalled.LNetworkCause)
	}
	return fmt.Sprintf("network stalled after %d attempts (tried %s): %v", LCommandAttemptMax, strings.Join(stalled.LNetworkAddresses, ", "), stalled.LNetworkCause)
}

func (stalled *LErrorNetworkStalled) Unwrap() error { return stalled.LNetworkCause }

// LNetworkAddressCollector accumulates the distinct hosts named by transient
// "failed retrieving file … from <host>" lines across every retry attempt of one
// command. It is written from the two per-attempt stream-copy goroutines, so its
// mutations are mutex-guarded.
type LNetworkAddressCollector struct {
	LMutex        sync.Mutex
	LNetworkSeen  map[string]bool
	LNetworkHosts []string
}

func (collector *LNetworkAddressCollector) LNetworkHostAdd(host string) {
	if collector == nil || host == "" {
		return
	}
	collector.LMutex.Lock()
	defer collector.LMutex.Unlock()
	if collector.LNetworkSeen == nil {
		collector.LNetworkSeen = map[string]bool{}
	}
	if collector.LNetworkSeen[host] {
		return
	}
	collector.LNetworkSeen[host] = true
	collector.LNetworkHosts = append(collector.LNetworkHosts, host)
}

func (collector *LNetworkAddressCollector) LNetworkHostListGet() []string {
	if collector == nil {
		return nil
	}
	collector.LMutex.Lock()
	defer collector.LMutex.Unlock()
	return append([]string{}, collector.LNetworkHosts...)
}

// LNetworkHostPattern extracts the server named by a pacman/curl transfer failure
// such as "error: failed retrieving file 'zlib.pkg.tar.zst' from repo.msys2.org : …".
// The first whitespace-delimited token after "from" is the host that stalled.
var LNetworkHostPattern = regexp.MustCompile(`(?i)failed retrieving file\s+.+?\s+from\s+(\S+)`)

// LNetworkHostParse returns the stalled host named by a transient-failure line, or
// an empty string when the line does not carry one.
func LNetworkHostParse(line string) string {
	match := LNetworkHostPattern.FindStringSubmatch(line)
	if match == nil {
		return ""
	}
	return strings.Trim(match[1], "'\"")
}

// LNetworkStalledCreate wraps an exhausted-retry error as LErrorNetworkStalled,
// attaching the authored mirror bases (the servers pacman was configured to try in
// order) followed by any distinct hosts parsed from the streamed failures.
func LNetworkStalledCreate(cause error, addressCollector *LNetworkAddressCollector) error {
	addresses := append([]string{}, scripting.LMSYSMirrorCatalog...)
	seen := map[string]bool{}
	for _, address := range addresses {
		seen[address] = true
	}
	for _, host := range addressCollector.LNetworkHostListGet() {
		if !seen[host] {
			seen[host] = true
			addresses = append(addresses, host)
		}
	}
	return &LErrorNetworkStalled{LNetworkAddresses: addresses, LNetworkCause: cause}
}

// LErrorNetworkMarkers are substrings (matched case-insensitively)
// that signal a download or connection failed for a transient reason rather
// than a real build/install error: a stalled transfer, a dropped or refused
// connection, DNS failure, or a 5xx from a mirror. A line carrying any of these
// makes the whole command eligible for retry. Markers are kept specific so a
// genuine compile/link error is never mistaken for a network blip.
var LErrorNetworkMarkers = []string{
	"operation too slow",
	"failed retrieving file",
	"could not resolve host",
	"name or service not known",
	"temporary failure in name resolution",
	"connection timed out",
	"connection refused",
	"connection reset",
	"network is unreachable",
	"transfer closed",
	"timeout was reached",
	"failed to commit transaction (unexpected error)",
	"rpc failed",
	"early eof",
	"the remote end hung up unexpectedly",
	"gnutls recv error",
	"ssl_read",
	"unexpected disconnect",
}

// LLogNetworkCheck reports whether a streamed output line indicates
// a transient network failure that warrants retrying the command.
func LLogNetworkCheck(line string) bool {
	lower := strings.ToLower(line)
	for _, marker := range LErrorNetworkMarkers {
		if strings.Contains(lower, marker) {
			return true
		}
	}
	return false
}
