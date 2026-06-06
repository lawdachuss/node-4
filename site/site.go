package site

import (
	"context"

	"github.com/teacat/chaturbate-dvr/internal"
)

// Room status constants.
const (
	StatusPublic  = "public"
	StatusPrivate = "private"
	StatusAway    = "away"
	StatusOffline = "offline"
)

// StreamInfo holds the result of fetching a stream for a model.
type StreamInfo struct {
	HLSSource    string
	RoomStatus   string
	RoomTitle    string
	Tags         []string
	NumUsers     int
	Gender       string
	LiveThumbURL string // live-updating thumbnail URL; empty = use site default
}

// Site is the interface that each live cam site must implement.
type Site interface {
	// FetchStream retrieves the HLS stream URL and room metadata.
	// Returns a non-nil StreamInfo even on error so callers can always
	// read RoomStatus.  The error indicates whether a stream URL was
	// obtained; RoomStatus reflects the current state of the room.
	FetchStream(ctx context.Context, req *internal.Req, username string) (*StreamInfo, error)

	// GetRoomStatus returns the room status string (public, private, away, offline, etc.).
	GetRoomStatus(ctx context.Context, req *internal.Req, username string) (string, error)
}
