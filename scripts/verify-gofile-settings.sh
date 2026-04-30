#!/bin/bash

# Script to verify and fix GoFile upload settings
# Run this on EC2 to ensure settings are correct

set -e

echo "🔍 Verifying GoFile Upload Settings..."
echo ""

# Check if conf directory exists
if [ ! -d "conf" ]; then
    echo "❌ conf/ directory not found!"
    echo "Creating conf/ directory..."
    mkdir -p conf
fi

# Check if settings.json exists
if [ ! -f "conf/settings.json" ]; then
    echo "⚠️  conf/settings.json not found!"
    echo "Creating default settings.json with GoFile enabled..."
    
    cat > conf/settings.json <<'EOF'
{
  "cookies": "",
  "user_agent": "",
  "enable_gofile_upload": true,
  "enable_supabase": true,
  "turboviplay_api_key": "",
  "voesx_api_key": "",
  "streamtape_login": "",
  "streamtape_api_key": "",
  "supabase_url": "",
  "supabase_api_key": "",
  "discord_webhook_url": "",
  "finalize_mode": "remux",
  "ffmpeg_container": "mp4",
  "ffmpeg_encoder": "libx264",
  "ffmpeg_quality": 18,
  "ffmpeg_preset": "slow",
  "disk_warning_percent": 80,
  "disk_critical_percent": 90
}
EOF
    echo "✅ Created conf/settings.json with GoFile enabled"
else
    echo "✅ conf/settings.json exists"
    
    # Check if enable_gofile_upload is present
    if grep -q "enable_gofile_upload" conf/settings.json; then
        # Check if it's set to true
        if grep -q '"enable_gofile_upload": true' conf/settings.json; then
            echo "✅ GoFile upload is ENABLED in settings.json"
        else
            echo "⚠️  GoFile upload is DISABLED in settings.json"
            echo "Enabling GoFile upload..."
            
            # Use jq if available, otherwise use sed
            if command -v jq &> /dev/null; then
                jq '.enable_gofile_upload = true' conf/settings.json > conf/settings.json.tmp
                mv conf/settings.json.tmp conf/settings.json
            else
                sed -i 's/"enable_gofile_upload": false/"enable_gofile_upload": true/g' conf/settings.json
            fi
            
            echo "✅ GoFile upload enabled"
        fi
    else
        echo "⚠️  enable_gofile_upload field missing from settings.json"
        echo "Adding enable_gofile_upload field..."
        
        # Add the field using jq if available
        if command -v jq &> /dev/null; then
            jq '. + {enable_gofile_upload: true}' conf/settings.json > conf/settings.json.tmp
            mv conf/settings.json.tmp conf/settings.json
        else
            # Fallback: add manually before the last closing brace
            sed -i 's/}$/,\n  "enable_gofile_upload": true\n}/' conf/settings.json
        fi
        
        echo "✅ Added enable_gofile_upload field"
    fi
fi

echo ""
echo "📄 Current settings.json content:"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
cat conf/settings.json
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

echo ""
echo "🔄 Restarting application to apply settings..."
if command -v docker &> /dev/null; then
    if docker ps | grep -q goondvr; then
        docker restart goondvr
        echo "✅ Application restarted"
        
        echo ""
        echo "⏳ Waiting for application to start..."
        sleep 5
        
        echo ""
        echo "📊 Checking application logs..."
        docker logs --tail 20 goondvr
    else
        echo "⚠️  goondvr container not running"
        echo "Start it with: docker compose up -d"
    fi
else
    echo "⚠️  Docker not found, please restart manually"
fi

echo ""
echo "✅ Verification complete!"
echo ""
echo "Next steps:"
echo "1. Open web UI: http://$(curl -s http://169.254.169.254/latest/meta-data/public-ipv4 2>/dev/null || echo 'localhost'):8080"
echo "2. Go to Settings (gear icon)"
echo "3. Verify 'Enable Multi-Host Upload' is checked"
echo "4. If not checked, check it and click Save"
