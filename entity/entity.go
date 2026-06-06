package entity

import (
	"regexp"
	"strings"
	"sync/atomic"
	"time"
)

// Event represents the type of event for the channel.
type Event = string

const (
	EventUpdate Event = "update"
	EventLog    Event = "log"
)

// ChannelConfig represents the configuration for a channel.
type ChannelConfig struct {
	IsPaused                  atomic.Bool `json:"-"`
	Site                      string      `json:"site"`       // "chaturbate" or "stripchat"
	Username                  string      `json:"username"`
	Framerate                 int         `json:"framerate"`
	Resolution                int         `json:"resolution"`
	Pattern                   string      `json:"pattern"`
	MaxDuration               int         `json:"max_duration"`
	MaxFilesize               int         `json:"max_filesize"`
	Compress                  bool        `json:"compress"`
	MinDurationBeforeUpload   int         `json:"min_duration_before_upload"` // seconds; 0 = disabled
	CreatedAt                 int64       `json:"created_at"`
}

func (c *ChannelConfig) Sanitize() {
	c.Username = regexp.MustCompile(`[^a-zA-Z0-9_-]`).ReplaceAllString(c.Username, "")
	c.Username = strings.TrimSpace(c.Username)
	if c.Site == "" {
		c.Site = "chaturbate"
	}
	if c.Resolution == 0 {
		c.Resolution = 2160
	}
	if c.MaxDuration == 0 {
		c.MaxDuration = 60
	}
	if c.Pattern == "" {
		c.Pattern = "videos/{{.Username}}_{{.Year}}-{{.Month}}-{{.Day}}_{{.Hour}}-{{.Minute}}-{{.Second}}{{if .Sequence}}_{{.Sequence}}{{end}}"
	}
	if c.Framerate == 0 {
		c.Framerate = 60
	}
}

// ChannelInfo represents the information about a channel,
// mostly used for the template rendering.
type ChannelInfo struct {
	IsOnline      bool
	IsConnecting  bool
	IsPaused      bool
	IsCompressing bool
	RoomStatus    string // public, private, group, away, offline, hidden
	Username      string
	Site          string // "chaturbate" or "stripchat"
	SiteDomain    string // domain for channel link, e.g. "https://chaturbate.com/"
	LiveThumbURL  string // live-updating thumbnail; empty = use platform default
	Duration      string
	Filesize      string
	Filename      string
	StreamedAt    string
	MaxDuration   string
	MaxFilesize   string
	CreatedAt     int64
	Logs           []string
	GlobalConfig   *Config // for nested template to access $.Config
	UploadStatus   string  // human-readable upload status (empty = idle)
	UploadProgress float64 // 0–100 upload progress estimate
	UploadFilename string  // file currently being uploaded
}

// DiskInfo holds disk usage information for the UI.
type DiskInfo struct {
	Total   string // formatted, e.g. "256.00 GB"
	Used    string // formatted, e.g. "120.50 GB"
	Free    string // formatted, e.g. "135.50 GB"
	Percent int    // 0-100
	UsedGB  float64
	TotalGB float64
}

// Config holds the configuration for the application.
type Config struct {
	Version       string
	Username      string
	AdminUsername string
	AdminPassword string
	Framerate     int
	Resolution    int
	Pattern       string
	MaxDuration   int
	MaxFilesize   int
	Compress      bool
	Port          string
	Interval      int
	Cookies       string
	SessionID     string
	Csrftoken     string
	CfClearance   string
	UserAgent     string
	Domain        string
	ProxyURL      string
	ProxyUsername string
	ProxyPassword string

	OutputDir               string
	PerModelFolder          bool
	DeleteLocalAfterUpload  bool
	MinDurationBeforeUpload  int  // seconds; 0 = disabled; videos shorter than this are deferred for merge
	OrphanCleanupInterval   int  // minutes between periodic orphan/thumbnail sweeps (0 = disabled)
	DiskWarningPercent      int  // log warning when disk usage exceeds this (0 = disabled, default 80)
	DiskCriticalPercent     int  // auto-delete oldest recordings when disk exceeds this (0 = disabled, default 90)
	MaxLocalAgeDays         int  // delete local files older than N days if uploaded (0 = disabled)

	VoeSXAPIKey       string
	StreamtapeLogin   string
	StreamtapeKey     string
	MixdropEmail      string
	MixdropToken      string
	PixelDrainToken   string

	SupabaseURL    string
	SupabaseAPIKey string

	StripchatPDKey string

	FFmpegPath string

	SessionDuration string // recording session length (e.g. "5h20m0s"); empty = disabled (continuous recording)
	SessionDurationParsed time.Duration // parsed from SessionDuration; 0 = disabled
}
