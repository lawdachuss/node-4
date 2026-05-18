package channel

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/teacat/chaturbate-dvr/uploader"
)

const (
	thumbWidth   = 640
	thumbHeight  = 360
	spriteFrames = 20  // Extract 20 frames for smooth preview
	spriteWidth  = 320 // Preview frame width
	spriteHeight = 180 // Preview frame height
)

// generateThumbnail is the channel-scoped wrapper — logs go to the channel log.
func (ch *Channel) generateThumbnail(videoPath string) (thumbURL, spriteURL string) {
	return generateThumbnailForFile(videoPath,
		func(f string, a ...interface{}) { ch.Info(f, a...) },
		func(f string, a ...interface{}) { ch.Error(f, a...) },
	)
}

// GenerateThumbnailForFile is a standalone thumbnail generator that can be
// called outside of a channel context (e.g. for pre-existing video files).
func GenerateThumbnailForFile(videoPath string) (thumbURL, spriteURL string) {
	return generateThumbnailForFile(videoPath,
		func(f string, a ...interface{}) { log.Printf("[thumb] "+f, a...) },
		func(f string, a ...interface{}) { log.Printf("[thumb:err] "+f, a...) },
	)
}

// generateThumbnailForFile creates thumbnail and sprite preview, uploads them
// to remote image hosts, and returns the remote URLs. Returns empty strings
// for any upload that fails. Always cleans up local JPG files.
func generateThumbnailForFile(videoPath string, info, errFn func(string, ...interface{})) (thumbURL, spriteURL string) {
	ext := strings.ToLower(filepath.Ext(videoPath))
	if ext != ".mp4" && ext != ".mkv" {
		return "", ""
	}

	baseName := filepath.Base(videoPath)

	// Generate both in parallel
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	thumbDone := make(chan string, 1)
	spriteDone := make(chan string, 1)

	// Thumbnail generation
	go func() {
		thumbJPG := videoPath + ".thumb.jpg"

		// Determine safe seek position (handles videos < 3s)
		seekPos := "00:00:03"
		probeOut, probeErr := exec.CommandContext(ctx, "ffprobe",
			"-v", "error",
			"-show_entries", "format=duration",
			"-of", "default=noprint_wrappers=1:nokey=1",
			videoPath,
		).Output()
		if probeErr == nil {
			dur, _ := strconv.ParseFloat(strings.TrimSpace(string(probeOut)), 64)
			if dur > 0 && dur < 3 {
				seekPos = fmt.Sprintf("%.2f", dur*0.5)
			}
		}

		err := exec.CommandContext(ctx, "ffmpeg",
			"-y",
			"-ss", seekPos,
			"-i", videoPath,
			"-vframes", "1",
			"-vf", fmt.Sprintf("scale=%d:%d:force_original_aspect_ratio=decrease,pad=%d:%d:(ow-iw)/2:(oh-ih)/2",
				thumbWidth, thumbHeight, thumbWidth, thumbHeight),
			"-q:v", "3",
			thumbJPG,
		).Run()

		if err != nil {
			errFn("thumb: failed for %s: %v", baseName, err)
			thumbDone <- ""
			return
		}

		// Upload to remote host
		imgUploader := uploader.NewMultiImageUploader()
		if remoteURL, _, uploadErr := imgUploader.Upload(thumbJPG); uploadErr == nil {
			info("thumb: ✓ %s", baseName)
			thumbDone <- remoteURL
		} else {
			errFn("thumb: upload failed for %s: %v", baseName, uploadErr)
			thumbDone <- ""
		}
	}()

	// Sprite generation
	go func() {
		spriteJPG := videoPath + ".sprite.jpg"

		// Probe video duration for even frame distribution
		var vf string
		probeOut, probeErr := exec.CommandContext(ctx, "ffprobe",
			"-v", "error",
			"-show_entries", "format=duration",
			"-of", "default=noprint_wrappers=1:nokey=1",
			videoPath,
		).Output()
		if probeErr == nil {
			dur, _ := strconv.ParseFloat(strings.TrimSpace(string(probeOut)), 64)
			if dur > 0 {
				fps := float64(spriteFrames) / dur
				vf = fmt.Sprintf("fps=%.8f,scale=%d:%d:flags=lanczos,tile=%dx1",
					fps, spriteWidth, spriteHeight, spriteFrames)
			}
		}
		if vf == "" {
			// Fallback: evenly sample every Nth frame
			vf = fmt.Sprintf("select='not(mod(n\\,%d))',scale=%d:%d:flags=lanczos,tile=%dx1",
				10, spriteWidth, spriteHeight, spriteFrames)
		}

		err := exec.CommandContext(ctx, "ffmpeg",
			"-y",
			"-i", videoPath,
			"-vf", vf,
			"-frames:v", fmt.Sprintf("%d", spriteFrames),
			"-q:v", "2",
			spriteJPG,
		).Run()

		if err != nil {
			errFn("sprite: failed for %s: %v", baseName, err)
			spriteDone <- ""
			return
		}

		// Upload to remote host
		imgUploader := uploader.NewMultiImageUploader()
		if remoteURL, _, uploadErr := imgUploader.Upload(spriteJPG); uploadErr == nil {
			info("sprite: ✓ %s (20 frames)", baseName)
			spriteDone <- remoteURL
		} else {
			errFn("sprite: upload failed for %s: %v", baseName, uploadErr)
			spriteDone <- ""
		}
	}()

	// Collect results and clean up
	thumbURL = <-thumbDone
	spriteURL = <-spriteDone

	// Always clean up local JPG files
	os.Remove(videoPath + ".thumb.jpg")
	os.Remove(videoPath + ".sprite.jpg")

	return thumbURL, spriteURL
}
