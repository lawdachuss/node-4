package server

import (
	"log"
	"os"

	"github.com/HeapOfChaos/goondvr/entity"
	"github.com/HeapOfChaos/goondvr/supabase"
)

var Config *entity.Config
var SupabaseClient *supabase.Client

// InitSupabase initializes the Supabase client if URL and API key are available.
// It checks CLI flags first, then falls back to environment variables.
func InitSupabase() {
	url := Config.SupabaseURL
	key := Config.SupabaseAPIKey

	// Fall back to environment variables (for Docker / GitHub Actions)
	if url == "" {
		url = os.Getenv("SUPABASE_URL")
	}
	if key == "" {
		key = os.Getenv("SUPABASE_API_KEY")
	}

	// Persist back to config so the rest of the app sees them
	if url != "" {
		Config.SupabaseURL = url
	}
	if key != "" {
		Config.SupabaseAPIKey = key
	}

	if url != "" && key != "" {
		SupabaseClient = supabase.NewClient(url, key)
		Config.EnableSupabase = true
		log.Printf("[SUPABASE] ✓ client initialized (url: %s)", url)
	} else {
		log.Println("[SUPABASE] ⚠️  not configured (missing URL or API key)")
	}
}
