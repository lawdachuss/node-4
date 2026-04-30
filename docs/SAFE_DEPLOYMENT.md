# Safe Deployment Guide

## What Happens During Deployment?

When you push code to GitHub and trigger a deployment, here's what happens to your recordings:

### 🔄 Deployment Process

1. **GitHub Actions triggers** → Deployment workflow starts
2. **Code is pulled** to EC2 instance
3. **Active recordings detected** → Workflow checks for in-progress recordings
4. **Graceful shutdown initiated** → Docker sends SIGTERM to container
5. **Application cleanup** (up to 5 minutes):
   - ✅ Current recording segments are closed
   - ✅ Files are synced to disk
   - ✅ Configuration is saved
   - ✅ Recordings are finalized (remux/transcode if enabled)
   - ⚠️ **Upload may not complete** if still in progress
6. **Container stops** → Old container removed
7. **New container starts** → Application resumes with saved state

### 📁 What Happens to Your Files?

#### ✅ **SAFE - Files are Preserved:**
- All recordings are stored in Docker **volumes** (`./videos`, `./database`, `./conf`)
- Volumes persist even when containers are rebuilt
- Your 5GB recording will be **saved** and **available** after deployment

#### ⚠️ **POTENTIAL ISSUES:**

1. **Incomplete Recordings:**
   - If a stream is actively recording when deployment happens
   - The current segment will be closed and saved
   - Recording will **resume automatically** when the new container starts
   - You'll have multiple files for the same stream session

2. **Upload Interruption:**
   - If a recording just finished and is uploading to GoFile
   - Upload might be interrupted if it takes > 5 minutes
   - File will remain in `videos/` directory
   - **Solution:** Orphaned file processor will upload it on next startup

3. **Disk Space:**
   - Recordings are NOT deleted during deployment
   - Disk space usage remains the same
   - Only Docker build cache is cleaned

### 🛡️ Protection Mechanisms

#### 1. **Graceful Shutdown (5 minutes)**
```yaml
stop_grace_period: 5m  # In docker-compose.yml
```
- Allows recordings to finalize properly
- Gives time for remux/transcode operations
- Prevents file corruption

#### 2. **Active Recording Detection**
```bash
# Deployment workflow checks for active recordings
ACTIVE_RECORDINGS=$(ls -1 /usr/src/app/videos/*.ts *.mp4 2>/dev/null | wc -l)
```
- Warns you if recordings are in progress
- Adds 30-second delay before shutdown
- Allows current segments to finish writing

#### 3. **Orphaned File Recovery**
- On startup, application scans for unprocessed recordings
- Automatically uploads any files that weren't uploaded
- Ensures no recordings are lost

#### 4. **Persistent Volumes**
```yaml
volumes:
  - ./videos:/usr/src/app/videos      # Recordings persist
  - ./database:/usr/src/app/database  # Upload logs persist
  - ./conf:/usr/src/app/conf          # Settings persist
```

### 📊 Deployment Scenarios

#### Scenario 1: No Active Recordings ✅
```
Push to GitHub → Build (2-3 min) → Deploy → Resume monitoring
```
**Impact:** None - Safe to deploy anytime

#### Scenario 2: Recording in Progress (< 5GB) ⚠️
```
Push to GitHub → Detect recording → Wait 30s → Graceful stop (closes file) 
→ Build → Deploy → Resume recording (new file)
```
**Impact:** 
- Current recording file is saved
- New file created when stream resumes
- You'll have 2+ files for the same stream session

#### Scenario 3: Large Recording + Upload in Progress ⚠️⚠️
```
Push to GitHub → Detect recording → Wait 30s → Graceful stop (5 min timeout)
→ If upload not done in 5 min → Force stop → File saved but not uploaded
→ Build → Deploy → Orphaned file processor uploads on startup
```
**Impact:**
- Recording is saved locally
- Upload might be delayed until next startup
- Disk space temporarily higher

### 🎯 Best Practices

#### 1. **Use Manual Deployment Trigger**
Instead of auto-deploying on every push, use manual workflow dispatch:
```bash
# In GitHub Actions UI, click "Run workflow"
# Choose "Skip rebuild" if only config changes
```

#### 2. **Deploy During Low Activity**
- Check your channels' streaming schedules
- Deploy when most channels are offline
- Monitor the dashboard before deploying

#### 3. **Use Skip Rebuild Option**
For configuration-only changes:
```yaml
# In workflow dispatch
skip_build: true  # Just restarts, no rebuild (faster, less disruptive)
```

#### 4. **Monitor Disk Space**
Before deploying, check disk usage:
```bash
ssh ubuntu@your-ec2-ip
df -h
docker system df
```

#### 5. **Check Active Recordings**
```bash
# SSH into EC2
docker exec goondvr ls -lh /usr/src/app/videos/

# Check for .ts or .mp4 files (active recordings)
```

### 🚨 Emergency: Stop Deployment

If you accidentally trigger deployment during important recording:

1. **Cancel GitHub Actions workflow** (click "Cancel workflow" in GitHub UI)
2. **Or SSH into EC2 and prevent container stop:**
   ```bash
   # This won't help if workflow already started, but you can monitor
   docker logs -f goondvr
   ```

### 📈 Disk Space Management

#### What Uses Disk Space:
1. **Active recordings** (`./videos/*.ts`, `*.mp4`)
2. **Completed recordings** (`./videos/completed/`)
3. **Docker images** (old builds)
4. **Docker build cache**

#### Automatic Cleanup:
- ✅ Docker build cache cleaned after each deployment
- ✅ Unused Docker images removed
- ✅ Recordings uploaded to GoFile are deleted locally
- ❌ Failed uploads remain in `videos/` (manual cleanup needed)

#### Manual Cleanup:
```bash
# SSH into EC2
cd /home/ubuntu/goondvr

# Check disk usage
df -h
du -sh videos/

# Remove old completed recordings (if not uploaded)
rm -rf videos/completed/*

# Clean Docker completely (CAUTION: stops all containers)
docker system prune -a --volumes
```

### 🔧 Configuration Changes Without Rebuild

For settings-only changes (cookies, API keys, etc.):

1. **Edit `conf/settings.json` directly on EC2:**
   ```bash
   ssh ubuntu@your-ec2-ip
   cd /home/ubuntu/goondvr
   nano conf/settings.json
   ```

2. **Restart without rebuild:**
   ```bash
   docker compose restart
   ```

3. **Or use workflow dispatch with `skip_build: true`**

### 📝 Summary

| Situation | Files Safe? | Upload Safe? | Disk Space Impact |
|-----------|-------------|--------------|-------------------|
| No active recordings | ✅ Yes | ✅ Yes | None |
| Recording in progress | ✅ Yes | ⚠️ Interrupted | None (file saved) |
| Upload in progress | ✅ Yes | ⚠️ May fail | Temporary increase |
| Large recording (>5GB) | ✅ Yes | ⚠️ May timeout | Temporary increase |

**Bottom Line:** Your recordings are **always saved** due to Docker volumes, but uploads might be interrupted. The orphaned file processor will handle any missed uploads on the next startup.

### 🎬 Recommended Workflow

1. **Make code changes locally**
2. **Test locally if possible** (`go build && ./goondvr`)
3. **Check dashboard** for active recordings
4. **Push to GitHub** when channels are mostly offline
5. **Monitor deployment** in GitHub Actions
6. **Verify** application is running after deployment
7. **Check logs** for any orphaned file uploads

### 🆘 Troubleshooting

#### "Recording was interrupted during deployment"
- ✅ File is saved in `videos/`
- ✅ Will be uploaded on next startup
- ✅ Recording resumes automatically

#### "Upload failed during deployment"
- ✅ File is saved locally
- ✅ Orphaned file processor will retry
- ⚠️ Check disk space

#### "Disk space full after deployment"
- Check for failed uploads: `ls -lh videos/`
- Manually upload: Use web UI "Upload Completed" button
- Or delete old recordings: `rm videos/completed/*`

### 📞 Need Help?

Check the logs:
```bash
# Application logs
docker logs -f goondvr

# Deployment logs
# Check GitHub Actions workflow run
```
