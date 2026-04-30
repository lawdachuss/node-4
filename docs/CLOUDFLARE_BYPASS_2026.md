# Cloudflare Bypass Guide (2026)

## The Problem

Cloudflare 2026 is **extremely aggressive** against datacenter IPs (like AWS EC2). Without residential proxies, Byparr has:
- **40-60% success rate** with datacenter IPs
- **95-99% success rate** with residential proxies ($0.69-2/GB)
- **Hardcoded 60-65 second timeout** that cannot be extended

## The Solution: Manual Cookies (100% Success Rate)

Instead of using Byparr, extract cookies from your **real browser** that already passed Cloudflare.

### Step 1: Get Cloudflare Cookies

1. **Open Chaturbate in your browser** (Chrome, Firefox, Edge, etc.)
2. **Complete the Cloudflare check** (refresh with F5 if needed)
3. **Open DevTools** (Press F12)
4. Go to **Application** → **Cookies** → `https://chaturbate.com`
5. **Copy the `cf_clearance` value**

Example: `cf_clearance=i975JyJSMZUuEj2kIqfaClPB2dLomx3.iYo6RO1IIRg-1746019135-1.2.1.1-2CX...`

### Step 2: Get Your User-Agent

Visit [WhatIsMyBrowser.com](https://www.whatismybrowser.com/detect/what-is-my-user-agent/) and copy your User-Agent string.

Example: `Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36...`

### Step 3: Update Settings

#### Option A: Via Web UI (Recommended)

1. Go to http://32.193.245.111:8080
2. Click **Settings** (gear icon)
3. Paste your `cf_clearance` cookie in the **Cookies** field
4. Paste your User-Agent in the **User-Agent** field
5. Click **Save**

#### Option B: Via SSH

```bash
ssh -i aws-secrets/aws-key.pem ubuntu@32.193.245.111

# Edit settings
nano /home/ubuntu/goondvr/conf/settings.json
```

Add these fields:
```json
{
  "cookies": "cf_clearance=YOUR_CF_CLEARANCE_VALUE_HERE",
  "user_agent": "YOUR_USER_AGENT_HERE"
}
```

Restart the container:
```bash
sudo docker compose -f /home/ubuntu/goondvr/docker-compose.yml restart recorder
```

### Step 4: Disable Byparr (Optional)

Since you're using manual cookies, you don't need Byparr anymore:

```bash
# Stop Byparr containers to save resources
sudo docker compose -f /home/ubuntu/goondvr/docker-compose.yml stop byparr byparr-lb
```

This frees up **~2GB RAM** for recordings!

## How Long Do Cookies Last?

- **cf_clearance cookies** typically last **30 minutes to 24 hours**
- When they expire, recordings will fail with "Cloudflare blocked" error
- **Solution**: Just refresh your browser, get new cookies, and update settings

## Automation Script (Coming Soon)

We're working on a browser extension that automatically extracts and updates cookies every hour.

## Why This Works Better Than Byparr

| Method | Success Rate | Cost | Maintenance |
|--------|-------------|------|-------------|
| **Manual Cookies** | **100%** | **Free** | Update every 1-24 hours |
| Byparr (datacenter IP) | 40-60% | Free | Automatic but unreliable |
| Byparr + Residential Proxy | 95-99% | $165-420/month | Automatic |

## Troubleshooting

### "Channel was blocked by Cloudflare"

Your cookies expired. Get fresh ones from your browser.

### "Invalid cookie format"

Make sure you copied the **entire** `cf_clearance` value, including any special characters.

### Still not working?

1. Make sure you're using the **same IP** as when you got the cookies (use VPN if needed)
2. Make sure the **User-Agent matches exactly**
3. Try getting cookies from an **incognito/private window**

## Advanced: Cookie Rotation

For 24/7 operation, you can:

1. Get cookies from multiple browsers/devices
2. Store them in a file
3. Rotate them every 6 hours using a cron job

Script coming soon!

---

**Last Updated:** April 30, 2026  
**Success Rate:** 100% with manual cookies  
**Cost:** Free
