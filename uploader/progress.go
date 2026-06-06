package uploader

import (
	"io"
	"sync"
)

// ProgressFunc is called with the host name, bytes read so far, and total bytes.
type ProgressFunc func(host string, current, total int64)

var (
	progressMu sync.Mutex
	progressFn ProgressFunc
)

// SetProgressCallback sets the global progress callback for upload operations.
func SetProgressCallback(fn ProgressFunc) {
	progressMu.Lock()
	progressFn = fn
	progressMu.Unlock()
}

// ClearProgressCallback clears the global progress callback.
func ClearProgressCallback() {
	progressMu.Lock()
	progressFn = nil
	progressMu.Unlock()
}

// reportProgress calls the global callback if set.
func reportProgress(host string, current, total int64) {
	progressMu.Lock()
	fn := progressFn
	progressMu.Unlock()
	if fn != nil {
		fn(host, current, total)
	}
}

// ProgressReader wraps an io.Reader and reports read progress via a callback.
type ProgressReader struct {
	reader  io.Reader
	total   int64
	read    int64
	host    string
	called  bool // whether we've already sent the initial (0, total) report
}

// NewProgressReader creates a ProgressReader that reports progress to the
// global callback. The host parameter identifies which upload host is reading.
func NewProgressReader(r io.Reader, total int64, host string) *ProgressReader {
	return &ProgressReader{reader: r, total: total, host: host}
}

func (pr *ProgressReader) Read(p []byte) (int, error) {
	if !pr.called {
		pr.called = true
		reportProgress(pr.host, 0, pr.total)
	}
	n, err := pr.reader.Read(p)
	pr.read += int64(n)
	if n > 0 {
		reportProgress(pr.host, pr.read, pr.total)
	}
	return n, err
}
