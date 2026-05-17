package uploader

import "fmt"

type imageHost struct {
	name   string
	upload func(string) (string, error)
}

// MultiImageUploader uploads thumbnails/sprites with durable fallbacks:
// Pixhost (NSFW API) → Catbox (permanent) → Freeimage.
type MultiImageUploader struct {
	hosts []imageHost
}

// NewMultiImageUploader creates the default thumbnail upload chain.
func NewMultiImageUploader() *MultiImageUploader {
	pixhost := NewThumbnailUploader("")
	catbox := NewCatboxUploader()
	freeimage := NewFreeimageUploader()
	return &MultiImageUploader{
		hosts: []imageHost{
			{name: "Pixhost", upload: pixhost.Upload},
			{name: "Catbox", upload: catbox.Upload},
			{name: "Freeimage", upload: freeimage.Upload},
		},
	}
}

// NewSpriteUploader creates an upload chain that prefers Catbox (file host, no
// recompression) over Pixhost (image host, may downscale wide sprites).
func NewSpriteUploader() *MultiImageUploader {
	catbox := NewCatboxUploader()
	pixhost := NewThumbnailUploader("")
	freeimage := NewFreeimageUploader()
	return &MultiImageUploader{
		hosts: []imageHost{
			{name: "Catbox", upload: catbox.Upload},
			{name: "Pixhost", upload: pixhost.Upload},
			{name: "Freeimage", upload: freeimage.Upload},
		},
	}
}

// Upload tries each host in order until one succeeds.
func (m *MultiImageUploader) Upload(filePath string) (url, host string, err error) {
	var lastErr error
	for _, h := range m.hosts {
		url, err = h.upload(filePath)
		if err == nil {
			return url, h.name, nil
		}
		lastErr = fmt.Errorf("%s: %w", h.name, err)
	}
	return "", "", fmt.Errorf("all image hosts failed: %w", lastErr)
}
