package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/teacat/chaturbate-dvr/database"
)

// TestConnection tests the Supabase database connection and basic operations
func main() {
	// Get credentials from environment
	url := os.Getenv("SUPABASE_URL")
	apiKey := os.Getenv("SUPABASE_API_KEY")

	if url == "" || apiKey == "" {
		fmt.Println("❌ Error: SUPABASE_URL and SUPABASE_API_KEY must be set")
		fmt.Println("\nUsage:")
		fmt.Println("  export SUPABASE_URL=https://your-project.supabase.co")
		fmt.Println("  export SUPABASE_API_KEY=your_anon_key")
		fmt.Println("  go run database/test_connection.go")
		os.Exit(1)
	}

	fmt.Println("🔍 Testing Supabase Connection...")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("URL: %s\n", url)
	fmt.Printf("API Key: %s...\n\n", apiKey[:20])

	// Create client
	client := database.NewClient(url, apiKey)

	// Test 1: Health Check
	fmt.Println("📋 Test 1: Health Check")
	if err := client.HealthCheck(); err != nil {
		fmt.Printf("❌ Health check failed: %v\n", err)
		fmt.Println("\n💡 Make sure you've run the SQL migration in your Supabase SQL Editor")
		os.Exit(1)
	}
	fmt.Println("✅ Health check passed\n")

	// Test 2: Create a test channel
	fmt.Println("📋 Test 2: Create Test Channel")
	testChannel := &database.Channel{
		Username:    "test_model_" + fmt.Sprint(time.Now().Unix()),
		IsPaused:    false,
		Framerate:   30,
		Resolution:  1080,
		Pattern:     "videos/{{.Username}}_{{.Year}}-{{.Month}}-{{.Day}}",
		MaxDuration: 60,
		MaxFilesize: 2048,
		Compress:    true,
		CreatedAt:   time.Now().Unix(),
	}

	if err := client.SaveChannel(testChannel); err != nil {
		fmt.Printf("❌ Failed to create channel: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✅ Created channel: %s\n\n", testChannel.Username)

	// Test 3: Retrieve the channel
	fmt.Println("📋 Test 3: Retrieve Channel")
	retrieved, err := client.GetChannel(testChannel.Username)
	if err != nil {
		fmt.Printf("❌ Failed to retrieve channel: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✅ Retrieved channel: %s\n", retrieved.Username)
	fmt.Printf("   - Resolution: %dp\n", retrieved.Resolution)
	fmt.Printf("   - Framerate: %d fps\n", retrieved.Framerate)
	fmt.Printf("   - Paused: %v\n\n", retrieved.IsPaused)

	// Test 4: Create a test recording
	fmt.Println("📋 Test 4: Create Test Recording")
	testRecording := &database.Recording{
		Username:     testChannel.Username,
		Filename:     fmt.Sprintf("%s_2024-01-15_20-30-00.mp4", testChannel.Username),
		Timestamp:    time.Now().Format(time.RFC3339),
		RoomTitle:    "Test Stream Title",
		Tags:         []string{"test", "demo", "sample"},
		Viewers:      150,
		Resolution:   "1920x1080",
		Framerate:    30,
		Filesize:     1024 * 1024 * 500, // 500 MB
		Gender:       "female",
		ThumbnailURL: "https://example.com/thumb.jpg",
		SpriteURL:    "https://example.com/sprite.jpg",
		EmbedURL:     "https://example.com/embed/xxxxx",
	}

	if err := client.SaveRecording(testRecording); err != nil {
		fmt.Printf("❌ Failed to create recording: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✅ Created recording: %s\n\n", testRecording.Filename)

	// Test 5: Retrieve the recording
	fmt.Println("📋 Test 5: Retrieve Recording")
	recRetrieved, err := client.GetRecording(testRecording.Filename)
	if err != nil {
		fmt.Printf("❌ Failed to retrieve recording: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✅ Retrieved recording: %s\n", recRetrieved.Filename)
	fmt.Printf("   - Room Title: %s\n", recRetrieved.RoomTitle)
	fmt.Printf("   - Tags: %v\n", recRetrieved.Tags)
	fmt.Printf("   - Viewers: %d\n", recRetrieved.Viewers)
	fmt.Printf("   - Filesize: %.2f MB\n\n", float64(recRetrieved.Filesize)/1024/1024)

	// Test 6: Create upload links
	fmt.Println("📋 Test 6: Create Upload Links")
	uploadLinks := []database.UploadLink{
		{RecordingID: recRetrieved.ID, Host: "streamtape", URL: "https://streamtape.com/v/test123"},
		{RecordingID: recRetrieved.ID, Host: "voesx", URL: "https://voe.sx/e/test456"},
		{RecordingID: recRetrieved.ID, Host: "sendcm", URL: "https://send.cm/d/test789"},
	}

	for _, link := range uploadLinks {
		if err := client.SaveUploadLink(&link); err != nil {
			fmt.Printf("❌ Failed to create upload link for %s: %v\n", link.Host, err)
			os.Exit(1)
		}
		fmt.Printf("✅ Created upload link: %s\n", link.Host)
	}
	fmt.Println()

	// Test 7: Retrieve upload links
	fmt.Println("📋 Test 7: Retrieve Upload Links")
	links, err := client.GetUploadLinks(recRetrieved.ID)
	if err != nil {
		fmt.Printf("❌ Failed to retrieve upload links: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✅ Retrieved %d upload links:\n", len(links))
	for _, link := range links {
		fmt.Printf("   - %s: %s\n", link.Host, link.URL)
	}
	fmt.Println()

	// Test 8: Save app setting
	fmt.Println("📋 Test 8: Save App Setting")
	testSetting := map[string]string{
		"cookies":    "test_cookie_value",
		"user_agent": "Mozilla/5.0 Test",
	}
	if err := client.SaveSetting("test_settings", testSetting); err != nil {
		fmt.Printf("❌ Failed to save setting: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✅ Saved app setting\n")

	// Test 9: Retrieve app setting
	fmt.Println("📋 Test 9: Retrieve App Setting")
	var retrievedSetting map[string]string
	if err := client.GetSetting("test_settings", &retrievedSetting); err != nil {
		fmt.Printf("❌ Failed to retrieve setting: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✅ Retrieved app setting:")
	settingJSON, _ := json.MarshalIndent(retrievedSetting, "   ", "  ")
	fmt.Printf("   %s\n\n", string(settingJSON))

	// Test 10: Save preview image
	fmt.Println("📋 Test 10: Save Preview Image")
	previewImg := &database.PreviewImage{
		Filename:     testRecording.Filename,
		ThumbnailURL: "https://example.com/thumb_updated.jpg",
		SpriteURL:    "https://example.com/sprite_updated.jpg",
		GithubPath:   "previews/test_model/thumb.jpg",
	}
	if err := client.SavePreviewImage(previewImg); err != nil {
		fmt.Printf("❌ Failed to save preview image: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✅ Saved preview image metadata\n")

	// Test 11: Retrieve preview image
	fmt.Println("📋 Test 11: Retrieve Preview Image")
	imgRetrieved, err := client.GetPreviewImage(testRecording.Filename)
	if err != nil {
		fmt.Printf("❌ Failed to retrieve preview image: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✅ Retrieved preview image:\n")
	fmt.Printf("   - Thumbnail: %s\n", imgRetrieved.ThumbnailURL)
	fmt.Printf("   - Sprite: %s\n", imgRetrieved.SpriteURL)
	fmt.Printf("   - GitHub Path: %s\n\n", imgRetrieved.GithubPath)

	// Test 12: Get all channels
	fmt.Println("📋 Test 12: Get All Channels")
	allChannels, err := client.GetAllChannels()
	if err != nil {
		fmt.Printf("❌ Failed to get all channels: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✅ Retrieved %d channels\n\n", len(allChannels))

	// Test 13: Get recordings by username
	fmt.Println("📋 Test 13: Get Recordings by Username")
	userRecordings, err := client.GetRecordingsByUsername(testChannel.Username)
	if err != nil {
		fmt.Printf("❌ Failed to get recordings: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✅ Retrieved %d recordings for %s\n\n", len(userRecordings), testChannel.Username)

	// Cleanup
	fmt.Println("🧹 Cleaning up test data...")
	if err := client.DeleteChannel(testChannel.Username); err != nil {
		fmt.Printf("⚠️  Warning: Failed to delete test channel: %v\n", err)
	} else {
		fmt.Println("✅ Deleted test channel")
	}

	if err := client.DeleteRecording(testRecording.Filename); err != nil {
		fmt.Printf("⚠️  Warning: Failed to delete test recording: %v\n", err)
	} else {
		fmt.Println("✅ Deleted test recording")
	}

	// Summary
	fmt.Println("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("🎉 All tests passed successfully!")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("\n✨ Your Supabase database is properly configured and ready to use!")
	fmt.Println("\nNext steps:")
	fmt.Println("  1. Update your .env file with SUPABASE_URL and SUPABASE_API_KEY")
	fmt.Println("  2. Restart your application")
	fmt.Println("  3. Start recording channels!")
}
