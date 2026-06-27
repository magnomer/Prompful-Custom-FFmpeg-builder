package download

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
)

func TestIsRetryableStatusCode(t *testing.T) {
	cases := []struct {
		statusCode int
		want       bool
	}{
		{http.StatusInternalServerError, true}, // 500
		{http.StatusBadGateway, true},          // 502 - common from CDN mirrors
		{http.StatusServiceUnavailable, true},  // 503
		{http.StatusGatewayTimeout, true},      // 504
		{http.StatusTooManyRequests, true},     // 429
		{http.StatusRequestTimeout, true},      // 408
		{http.StatusNotFound, false},           // 404 - permanent
		{http.StatusForbidden, false},          // 403 - permanent
		{http.StatusOK, false},
	}
	for _, c := range cases {
		if got := isRetryableStatusCode(c.statusCode); got != c.want {
			t.Errorf("isRetryableStatusCode(%d) = %v, want %v", c.statusCode, got, c.want)
		}
	}
}

func TestActivityTrackingReaderCountsBytes(t *testing.T) {
	var bytesTransferred atomic.Int64
	reader := &activityTrackingReader{reader: strings.NewReader("payload"), bytesTransferred: &bytesTransferred}

	total := 0
	buffer := make([]byte, 4)
	for {
		n, err := reader.Read(buffer)
		total += n
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("Read returned error: %v", err)
		}
	}
	if got := bytesTransferred.Load(); got != int64(total) || got != int64(len("payload")) {
		t.Errorf("bytesTransferred = %d, want %d", got, len("payload"))
	}
}

func TestActivityTrackingReaderNoCountOnEOF(t *testing.T) {
	var bytesTransferred atomic.Int64
	reader := &activityTrackingReader{reader: strings.NewReader(""), bytesTransferred: &bytesTransferred}

	_, err := reader.Read(make([]byte, 4))
	if err != io.EOF {
		t.Fatalf("expected io.EOF, got %v", err)
	}
	if bytesTransferred.Load() != 0 {
		t.Error("expected no byte count on an empty (EOF) read")
	}
}

func TestDescribeStallError(t *testing.T) {
	base := io.ErrUnexpectedEOF
	if got := describeStallError(stallKindNone, base); got != base {
		t.Errorf("stallKindNone should return the base error unchanged, got %v", got)
	}
	zero := describeStallError(stallKindZeroData, base)
	if !errors.Is(zero, base) || !strings.Contains(zero.Error(), "no data received") {
		t.Errorf("stallKindZeroData error = %v, want it to wrap base and mention no data", zero)
	}
	slow := describeStallError(stallKindTooSlow, base)
	if !errors.Is(slow, base) || !strings.Contains(slow.Error(), "too slow") {
		t.Errorf("stallKindTooSlow error = %v, want it to wrap base and mention too slow", slow)
	}
}
