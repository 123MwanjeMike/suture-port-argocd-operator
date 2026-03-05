package common

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

// ── Latency logger ────────────────────────────────────────────────────────────

var (
	latencyOnce   sync.Once
	latencyWriter *csv.Writer
	latencyMu     sync.Mutex
	requestCounter atomic.Uint64
)

func initLatencyLog() {
	latencyOnce.Do(func() {
		f, err := os.OpenFile("/tmp/operator-latency.csv", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[latency] failed to open log file: %v\n", err)
			return
		}
		latencyWriter = csv.NewWriter(f)
		_ = latencyWriter.Write([]string{
			"request_id", "timestamp_unix_ns", "verb", "path", "duration_ms", "status_code",
		})
		latencyWriter.Flush()
	})
}

func recordLatency(verb, path string, durationMs float64, statusCode int) {
	if latencyWriter == nil {
		return
	}
	id := requestCounter.Add(1)
	latencyMu.Lock()
	defer latencyMu.Unlock()
	_ = latencyWriter.Write([]string{
		strconv.FormatUint(id, 10),
		strconv.FormatInt(time.Now().UnixNano(), 10),
		verb,
		path,
		strconv.FormatFloat(durationMs, 'f', 4, 64),
		strconv.Itoa(statusCode),
	})
	latencyWriter.Flush()
}

// ── Transport ─────────────────────────────────────────────────────────────────

// SutureIDTransport wraps an http.RoundTripper, adds the Suture_ID header to
// every request, and records per-request latency to /tmp/operator-latency.csv.
type SutureIDTransport struct {
	Base     http.RoundTripper
	SutureID string
}

// RoundTrip executes a single HTTP transaction.
func (s *SutureIDTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if s.SutureID != "" {
		req.Header.Set("Suture_ID", s.SutureID)
	}

	// Skip long-lived requests with no individual round-trips.
	isWatch := req.URL.Query().Get("watch") == "true"

	var start time.Time
	if !isWatch {
		start = time.Now()
	}

	resp, err := s.Base.RoundTrip(req)

	if !isWatch {
		durationMs := float64(time.Since(start).Microseconds()) / 1000.0
		statusCode := 0
		if resp != nil {
			statusCode = resp.StatusCode
		}
		recordLatency(req.Method, req.URL.Path, durationMs, statusCode)
	}

	return resp, err
}

// WrapTransportWithSutureID wraps the provided transport with SutureIDTransport
// if the SUTURE_ID environment variable is set.
func WrapTransportWithSutureID(base http.RoundTripper) http.RoundTripper {
	initLatencyLog()
	sutureID := os.Getenv("SUTURE_ID")
	if sutureID == "" {
		// Still wrap so latency is recorded even in baseline runs.
		return &SutureIDTransport{Base: base, SutureID: ""}
	}
	return &SutureIDTransport{
		Base:     base,
		SutureID: sutureID,
	}
}
