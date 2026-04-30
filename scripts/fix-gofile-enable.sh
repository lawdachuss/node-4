#!/bin/bash

# Emergency fix script for GoFile upload settings
# Run this directly on EC2 if GoFile upload is showing as disabled

echo "🚨 Emergency GoFile Upload Fix Script"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

cd /home/ubuntu/goondvr || {
    echo "❌ Cannot find /home/ubuntu/goondvr directory!"
    exit 1
}

echo "📍 Current directory: $(pwd)"
echo ""

# Step 1: Check if conf directory exists
echo "Step 1: Checking conf/ directory..."
if [ ! -d "conf" ]; then
    echo "⚠️  conf/ directory missing - creating..."
    mkdir -p conf
    echo "✅ Created conf/ directory"
else
    echo "✅ conf/ directory exists"
fi

# Step 2: Backup existing settings if they exist
if [ -f "conf/settings.json" ]; then
    echo ""
    echo "Step 2: Backing up existing settings..."
    cp conf/settings.json conf/settings.json.backup.$(date +%Y%m%d_%H%M%S)
    echo "✅ Backup created"
    
    echo ""
    echo "Current settings.json content:"
    cat conf/settings.json
else
    echo ""
    echo "Step 2: No existing settings.json found"
fi

# Step 3: Create/Update settings.json with GoFile enabled
echo ""
echo "Step 3: Ensuring GoFile upload is enabled..."

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

echo "✅ Created fresh settings.json with GoFile enabled"

# Step 4: Verify the file
echo ""
echo "Step 4: Verifying settings.json..."
if grep -q '"enable_gofile_upload": true' conf/settings.json; then
    echo "✅ GoFile upload is ENABLED in settings.json"
else
    echo "❌ ERROR: GoFile upload still not enabled!"
    exit 1
fi

# Step 5: Check Docker container
echo ""
echo "Step 5: Checking Docker container..."
if docker ps | grep -q goondvr; then
    echo "✅ goondvr container is running"
    
    # Check if conf is mounted
    echo ""
    echo "Checking volume mounts..."
    docker inspect goondvr | grep -A 5 "Mounts" || echo "Cannot inspect mounts"
    
    # Restart container
    echo ""
    echo "Step 6: Restarting container to apply settings..."
    docker restart goondvr
    
    echo "⏳ Waiting for container to start..."
    sleep 10
    
    echo ""
    echo "Step 7: Checking application logs..."
    docker logs --tail 30 goondvr
    
    echo ""
    echo "✅ Container restarted successfully"
else
    echo "⚠️  goondvr container is not running!"
    echo "Starting with docker compose..."
    docker compose up -d
    sleep 10
fi

# Step 8: Final verification
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "✅ FIX COMPLETE!"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "📄 Final settings.json:"
cat conf/settings.json
echo ""
echo "🌐 Access your application:"
PUBLIC_IP=$(curl -s http://169.254.169.254/latest/meta-data/public-ipv4 2>/dev/null || echo "localhost")
echo "   http://${PUBLIC_IP}:8080"
echo ""
echo "📋 Next steps:"
echo "1. Open the web UI"
echo "2. Go to Settings (gear icon)"
echo "3. Verify 'Enable Multi-Host Upload' checkbox is CHECKED"
echo "4. If it's unchecked, CHECK IT and click Save"
echo "5. The checkbox should now stay checked"
echo ""
echo "If the checkbox is still unchecked after this:"
echo "- The web UI might be caching old settings"
echo "- Try hard refresh: Ctrl+Shift+R (Windows) or Cmd+Shift+R (Mac)"
echo "- Or open in incognito/private window"
