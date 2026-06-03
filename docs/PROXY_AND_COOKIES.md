# Proxy and Cookie Configuration Guide

## Problem: "Stream URL unavailable (check cookies)"

This error occurs when using a proxy/VPN to change your region, but your cookies were obtained from a different region. Chaturbate validates that your cookies match your request IP/region.

## Why Some Channels Work and Others Don't

1. **Geo-restrictions vary per model** - Some models allow all regions, others restrict specific countries
2. **Cookie expiration** - Cloudflare cookies (`cf_clearance`, `__cf_bm`) expire quickly (minutes to hours)
3. **IP/Cookie mismatch** - Your cookies are from one region, but requests come from another (via proxy/VPN)

## Solutions

### Option 1: Get Fresh Cookies Through Your Proxy (Recommended)

1. **Connect to your proxy/VPN first**
2. **Open your browser with proxy configured**
3. **Visit chaturbate.com and log in**
4. **Get fresh cookies** from the browser (F12 → Application → Cookies)
5. **Update your `.env` file** with the new cookies

This ensures your cookies match the proxy's region.

### Option 2: Configure Proxy in .env

If you're using an HTTP/SOCKS proxy, configure it in `.env`:

```env
PROXY_URL="http://proxy.example.com:8080"
PROXY_USERNAME="your_username"  # Optional
PROXY_PASSWORD="your_password"  # Optional
```

Then get fresh cookies through that proxy (see Option 1).

### Option 3: Disable Proxy for Recording

If you don't need region changes:
- Disconnect your VPN/proxy
- Get fresh cookies from your actual region
- Update `.env` and restart the app

## How to Get Fresh Cookies

### Chrome/Edge:
1. Visit https://chaturbate.com and log in
2. Press F12 to open DevTools
3. Go to **Application** tab → **Cookies** → `https://chaturbate.com`
4. Copy all cookie values
5. Format as: `name1=value1; name2=value2; ...`

### Required Cookies:
- `csrftoken` - Required for API requests
- `cf_clearance` - Cloudflare clearance token (expires quickly!)
- `__cf_bm` - Cloudflare bot management (expires in ~30 minutes)
- Session cookies (`sbr`, `_iidt`, `_vid_t`, etc.)

## Troubleshooting

### All channels show "stream URL unavailable":
→ Your cookies expired or don't match your IP region
→ Get fresh cookies (see above)

### Some channels work, others don't:
→ Working channels allow your region, others have geo-restrictions
→ Try a different proxy region or contact the model about geo-restrictions

### Cookies expire frequently:
→ Cloudflare's `__cf_bm` expires every 30 minutes
→ Consider refreshing cookies periodically or use a browser automation tool

## Cookie Lifetime

- `cf_clearance`: 1-24 hours (varies)
- `__cf_bm`: ~30 minutes
- `csrftoken`: Days to weeks
- Session cookies: Until logout or browser close

**Best practice**: Refresh cookies every few hours for reliable recording.

## Advanced: Automatic Cookie Refresh

For production setups, consider:
1. Browser automation (Puppeteer/Playwright) to refresh cookies
2. Cookie rotation service
3. Multiple cookie sets for different regions
4. Webhook notifications when cookies expire

## Debugging

The app now shows debug logs:
```
[DEBUG] username POST API response: status=public url=https://...
[WARN] username: POST API returned empty URL, trying GET API fallback (check cookies if this persists)
[DEBUG] POST API 403 response for username: {"error": "..."}
```

Check these logs to diagnose cookie issues.
