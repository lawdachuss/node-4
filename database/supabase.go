package database

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client represents a Supabase database client
type Client struct {
	URL    string
	APIKey string
	client *http.Client
}

// NewClient creates a new Supabase client
func NewClient(url, apiKey string) *Client {
	return &Client{
		URL:    url,
		APIKey: apiKey,
		client: &http.Client{Timeout: 15 * time.Second},
	}
}

// ============================================================================
// HTTP HELPERS
// ============================================================================

func (c *Client) request(method, path string, body interface{}) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonData)
	}

	req, err := http.NewRequest(method, c.URL+"/rest/v1"+path, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("apikey", c.APIKey)
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
		// Use resolution=merge-duplicates for upsert behavior
		req.Header.Set("Prefer", "resolution=merge-duplicates,return=representation")
	}

	return c.client.Do(req)
}

func (c *Client) get(path string, result interface{}) error {
	resp, err := c.request("GET", path, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	return json.NewDecoder(resp.Body).Decode(result)
}

func (c *Client) post(path string, body interface{}, result interface{}) error {
	resp, err := c.request("POST", path, body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(bodyBytes))
	}

	if result != nil && resp.StatusCode != http.StatusNoContent {
		return json.NewDecoder(resp.Body).Decode(result)
	}
	return nil
}

func (c *Client) patch(path string, body interface{}) error {
	resp, err := c.request("PATCH", path, body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(bodyBytes))
	}
	return nil
}

func (c *Client) delete(path string) error {
	resp, err := c.request("DELETE", path, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(bodyBytes))
	}
	return nil
}

// ============================================================================
// CHANNELS
// ============================================================================

type Channel struct {
	ID          string `json:"id,omitempty"`
	Username    string `json:"username"`
	IsPaused    bool   `json:"is_paused"`
	Framerate   int    `json:"framerate"`
	Resolution  int    `json:"resolution"`
	Pattern     string `json:"pattern"`
	MaxDuration int    `json:"max_duration"`
	MaxFilesize int    `json:"max_filesize"`
	Compress    bool   `json:"compress"`
	CreatedAt   int64  `json:"created_at"`
	UpdatedAt   string `json:"updated_at,omitempty"`
}

// SaveChannel creates or updates a channel using Supabase's upsert functionality
func (c *Client) SaveChannel(ch *Channel) error {
	// First, try to get existing channel
	existing, err := c.GetChannel(ch.Username)
	if err == nil && existing != nil {
		// Channel exists, update it
		return c.patch(fmt.Sprintf("/channels?username=eq.%s", ch.Username), ch)
	}

	// Channel doesn't exist, insert it
	resp, err := c.request("POST", "/channels", ch)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(bodyBytes))
	}
	return nil
}

// GetChannel retrieves a channel by username
func (c *Client) GetChannel(username string) (*Channel, error) {
	var channels []Channel
	err := c.get(fmt.Sprintf("/channels?username=eq.%s&limit=1", username), &channels)
	if err != nil {
		return nil, err
	}
	if len(channels) == 0 {
		return nil, fmt.Errorf("channel not found")
	}
	return &channels[0], nil
}

// GetAllChannels retrieves all channels
func (c *Client) GetAllChannels() ([]Channel, error) {
	var channels []Channel
	err := c.get("/channels?order=created_at.desc", &channels)
	return channels, err
}

// DeleteChannel removes a channel
func (c *Client) DeleteChannel(username string) error {
	return c.delete(fmt.Sprintf("/channels?username=eq.%s", username))
}

// ============================================================================
// RECORDINGS
// ============================================================================

type Recording struct {
	ID           string   `json:"id,omitempty"`
	ChannelID    string   `json:"channel_id,omitempty"`
	Username     string   `json:"username"`
	Filename     string   `json:"filename"`
	Timestamp    string   `json:"timestamp"`
	RoomTitle    string   `json:"room_title,omitempty"`
	Tags         []string `json:"tags,omitempty"`
	Viewers      int      `json:"viewers"`
	Resolution   string   `json:"resolution,omitempty"`
	Framerate    int      `json:"framerate"`
	Filesize     int64    `json:"filesize"`
	Gender       string   `json:"gender,omitempty"`
	ThumbnailURL string   `json:"thumbnail_url,omitempty"`
	SpriteURL    string   `json:"sprite_url,omitempty"`
	EmbedURL     string   `json:"embed_url,omitempty"`
	CreatedAt    string   `json:"created_at,omitempty"`
	UpdatedAt    string   `json:"updated_at,omitempty"`
}

// SaveRecording creates or updates a recording using Supabase's upsert functionality
func (c *Client) SaveRecording(rec *Recording) error {
	// First, try to get existing recording
	existing, err := c.GetRecording(rec.Filename)
	if err == nil && existing != nil {
		// Recording exists, update it
		return c.patch(fmt.Sprintf("/recordings?filename=eq.%s", rec.Filename), rec)
	}

	// Recording doesn't exist, insert it
	resp, err := c.request("POST", "/recordings", rec)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(bodyBytes))
	}
	return nil
}

// GetRecording retrieves a recording by filename
func (c *Client) GetRecording(filename string) (*Recording, error) {
	var recordings []Recording
	err := c.get(fmt.Sprintf("/recordings?filename=eq.%s&limit=1", filename), &recordings)
	if err != nil {
		return nil, err
	}
	if len(recordings) == 0 {
		return nil, fmt.Errorf("recording not found")
	}
	return &recordings[0], nil
}

// GetRecordingsByUsername retrieves all recordings for a username
func (c *Client) GetRecordingsByUsername(username string) ([]Recording, error) {
	var recordings []Recording
	err := c.get(fmt.Sprintf("/recordings?username=eq.%s&order=timestamp.desc", username), &recordings)
	return recordings, err
}

// GetAllRecordings retrieves all recordings
func (c *Client) GetAllRecordings() ([]Recording, error) {
	var recordings []Recording
	err := c.get("/recordings?order=timestamp.desc", &recordings)
	return recordings, err
}

// DeleteRecording removes a recording
func (c *Client) DeleteRecording(filename string) error {
	return c.delete(fmt.Sprintf("/recordings?filename=eq.%s", filename))
}

// ============================================================================
// UPLOAD LINKS
// ============================================================================

type UploadLink struct {
	ID          string `json:"id,omitempty"`
	RecordingID string `json:"recording_id"`
	Host        string `json:"host"`
	URL         string `json:"url"`
	UploadedAt  string `json:"uploaded_at,omitempty"`
}

// SaveUploadLink creates a new upload link
func (c *Client) SaveUploadLink(link *UploadLink) error {
	var result []UploadLink
	return c.post("/upload_links", link, &result)
}

// GetUploadLinks retrieves all upload links for a recording
func (c *Client) GetUploadLinks(recordingID string) ([]UploadLink, error) {
	var links []UploadLink
	err := c.get(fmt.Sprintf("/upload_links?recording_id=eq.%s", recordingID), &links)
	return links, err
}

// ============================================================================
// APP SETTINGS
// ============================================================================

type AppSetting struct {
	Key       string          `json:"key"`
	Value     json.RawMessage `json:"value"`
	UpdatedAt string          `json:"updated_at,omitempty"`
}

// SaveSetting creates or updates an app setting
func (c *Client) SaveSetting(key string, value interface{}) error {
	jsonValue, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("marshal value: %w", err)
	}

	setting := &AppSetting{
		Key:   key,
		Value: jsonValue,
	}

	// Upsert using Prefer header
	resp, err := c.request("POST", "/app_settings", setting)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

// GetSetting retrieves an app setting
func (c *Client) GetSetting(key string, result interface{}) error {
	var settings []AppSetting
	err := c.get(fmt.Sprintf("/app_settings?key=eq.%s&limit=1", key), &settings)
	if err != nil {
		return err
	}
	if len(settings) == 0 {
		return fmt.Errorf("setting not found")
	}

	return json.Unmarshal(settings[0].Value, result)
}

// ============================================================================
// TUNNELS
// ============================================================================

type Tunnel struct {
	ID        string `json:"id,omitempty"`
	URL       string `json:"url"`
	RunID     int    `json:"run_id"`
	IsActive  bool   `json:"is_active"`
	CreatedAt string `json:"created_at,omitempty"`
	ExpiresAt string `json:"expires_at,omitempty"`
}

// SaveTunnel creates a new tunnel
func (c *Client) SaveTunnel(tunnel *Tunnel) error {
	var result []Tunnel
	return c.post("/tunnels", tunnel, &result)
}

// GetActiveTunnel retrieves the most recent active tunnel
func (c *Client) GetActiveTunnel() (*Tunnel, error) {
	var tunnels []Tunnel
	err := c.get("/tunnels?is_active=eq.true&order=created_at.desc&limit=1", &tunnels)
	if err != nil {
		return nil, err
	}
	if len(tunnels) == 0 {
		return nil, fmt.Errorf("no active tunnel found")
	}
	return &tunnels[0], nil
}

// DeactivateOldTunnels marks all tunnels as inactive
func (c *Client) DeactivateOldTunnels() error {
	return c.patch("/tunnels?is_active=eq.true", map[string]interface{}{
		"is_active": false,
	})
}

// ============================================================================
// CHANNEL LOGS
// ============================================================================

type ChannelLog struct {
	ID        string `json:"id,omitempty"`
	ChannelID string `json:"channel_id,omitempty"`
	Username  string `json:"username"`
	LogLevel  string `json:"log_level"`
	Message   string `json:"message"`
	CreatedAt string `json:"created_at,omitempty"`
}

// SaveLog creates a new log entry
func (c *Client) SaveLog(log *ChannelLog) error {
	var result []ChannelLog
	return c.post("/channel_logs", log, &result)
}

// GetLogs retrieves logs for a channel
func (c *Client) GetLogs(username string, limit int) ([]ChannelLog, error) {
	var logs []ChannelLog
	err := c.get(fmt.Sprintf("/channel_logs?username=eq.%s&order=created_at.desc&limit=%d", username, limit), &logs)
	return logs, err
}

// ============================================================================
// PREVIEW IMAGES
// ============================================================================

type PreviewImage struct {
	ID           string `json:"id,omitempty"`
	RecordingID  string `json:"recording_id,omitempty"`
	Filename     string `json:"filename"`
	ThumbnailURL string `json:"thumbnail_url,omitempty"`
	SpriteURL    string `json:"sprite_url,omitempty"`
	GithubPath   string `json:"github_path,omitempty"`
	UploadedAt   string `json:"uploaded_at,omitempty"`
}

// SavePreviewImage creates or updates preview image metadata using Supabase's upsert functionality
func (c *Client) SavePreviewImage(img *PreviewImage) error {
	// First, try to get existing preview image
	existing, err := c.GetPreviewImage(img.Filename)
	if err == nil && existing != nil {
		// Preview image exists, update it
		return c.patch(fmt.Sprintf("/preview_images?filename=eq.%s", img.Filename), img)
	}

	// Preview image doesn't exist, insert it
	resp, err := c.request("POST", "/preview_images", img)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(bodyBytes))
	}
	return nil
}

// GetPreviewImage retrieves preview image metadata
func (c *Client) GetPreviewImage(filename string) (*PreviewImage, error) {
	var images []PreviewImage
	err := c.get(fmt.Sprintf("/preview_images?filename=eq.%s&limit=1", filename), &images)
	if err != nil {
		return nil, err
	}
	if len(images) == 0 {
		return nil, fmt.Errorf("preview image not found")
	}
	return &images[0], nil
}

// ============================================================================
// DISK USAGE
// ============================================================================

type DiskUsage struct {
	ID           string `json:"id,omitempty"`
	TotalBytes   int64  `json:"total_bytes"`
	UsedBytes    int64  `json:"used_bytes"`
	FreeBytes    int64  `json:"free_bytes"`
	PercentUsed  int    `json:"percent_used"`
	RecordedAt   string `json:"recorded_at,omitempty"`
}

// SaveDiskUsage records current disk usage
func (c *Client) SaveDiskUsage(usage *DiskUsage) error {
	var result []DiskUsage
	return c.post("/disk_usage", usage, &result)
}

// GetLatestDiskUsage retrieves the most recent disk usage record
func (c *Client) GetLatestDiskUsage() (*DiskUsage, error) {
	var usages []DiskUsage
	err := c.get("/disk_usage?order=recorded_at.desc&limit=1", &usages)
	if err != nil {
		return nil, err
	}
	if len(usages) == 0 {
		return nil, fmt.Errorf("no disk usage records found")
	}
	return &usages[0], nil
}

// ============================================================================
// HEALTH CHECK
// ============================================================================

// HealthCheck verifies the database connection
func (c *Client) HealthCheck() error {
	resp, err := c.request("GET", "/app_settings?key=eq.__healthcheck__&select=key&limit=1", nil)
	if err != nil {
		return fmt.Errorf("health check request failed: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case 200:
		return nil
	case 404:
		return fmt.Errorf("app_settings table not found (HTTP 404) — run the SQL migration first")
	case 401, 403:
		return fmt.Errorf("authentication failed (HTTP %d) — check SUPABASE_API_KEY and RLS policies", resp.StatusCode)
	default:
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected response (HTTP %d): %s", resp.StatusCode, string(body))
	}
}
