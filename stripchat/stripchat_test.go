package stripchat

import (
	"encoding/json"
	"testing"
)

func TestCamResponseUnmarshal(t *testing.T) {
	// Case 1: "cam" is an object (online/active)
	jsonDataActive := []byte(`{
		"cam": {
			"streamName": "test_stream",
			"isCamActive": true,
			"viewServers": {"flashphoner-hls": "server1"},
			"broadcastSettings": {"broadcastType": "public"},
			"topic": "hello world"
		},
		"user": {
			"user": {
				"id": 12345,
				"username": "test_user",
				"isOnline": true,
				"isLive": true,
				"status": "public",
				"broadcastGender": "female",
				"previewUrlThumbBig": "https://thumb.com/pic.jpg",
				"snapshotTimestamp": 1600000000
			}
		}
	}`)

	var respActive camResponse
	if err := json.Unmarshal(jsonDataActive, &respActive); err != nil {
		t.Fatalf("Failed to unmarshal active cam JSON: %v", err)
	}

	if respActive.Cam.StreamName != "test_stream" {
		t.Errorf("Expected StreamName 'test_stream', got %q", respActive.Cam.StreamName)
	}
	if !respActive.Cam.IsCamActive {
		t.Errorf("Expected IsCamActive true")
	}
	if respActive.User.User.Username != "test_user" {
		t.Errorf("Expected Username 'test_user', got %q", respActive.User.User.Username)
	}

	// Case 2: "cam" is an empty array "[]" (offline/inactive/unmarshaling issue)
	jsonDataInactive := []byte(`{
		"cam": [],
		"user": {
			"user": {
				"id": 12345,
				"username": "test_user",
				"isOnline": false,
				"isLive": false,
				"status": "offline",
				"broadcastGender": "female",
				"previewUrlThumbBig": "https://thumb.com/pic.jpg",
				"snapshotTimestamp": 1600000000
			}
		}
	}`)

	var respInactive camResponse
	if err := json.Unmarshal(jsonDataInactive, &respInactive); err != nil {
		t.Fatalf("Failed to unmarshal inactive cam JSON: %v", err)
	}

	// Cam should remain zero-valued
	if respInactive.Cam.StreamName != "" {
		t.Errorf("Expected empty StreamName, got %q", respInactive.Cam.StreamName)
	}
	if respInactive.Cam.IsCamActive {
		t.Errorf("Expected IsCamActive false")
	}
	if respInactive.User.User.Username != "test_user" {
		t.Errorf("Expected Username 'test_user', got %q", respInactive.User.User.Username)
	}
}
