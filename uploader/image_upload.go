package uploader

import (
	"fmt"
	"time"
)

// pixhostSem limits concurrent Pixhost.to uploads to avoid rate limiting.
var pixhostSem = make(chan struct{}, 5)

// MultiImageUploader uploads thumbnails/sprites to Pixhost.to.
type MultiImageUploader struct {
	pixhost *ThumbnailUploader
}

// NewMultiImageUploader creates a new uploader backed by Pixhost.to.
func NewMultiImageUploader() *MultiImageUploader {
	return &MultiImageUploader{pixhost: NewThumbnailUploader("")}
}

// Upload uploads a file to Pixhost.to and returns the URL.
func (m *MultiImageUploader) Upload(filePath string) (url, host string, err error) {
	pixhostSem <- struct{}{}
	defer func() { <-pixhostSem }()

	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(1<<attempt) * time.Second)
		}
		u, err := m.pixhost.Upload(filePath)
		if err == nil {
			return u, "Pixhost", nil
		}
		lastErr = err
	}
	return "", "", fmt.Errorf("pixhost: %w", lastErr)
}
