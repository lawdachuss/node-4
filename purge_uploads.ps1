#Requires -Version 5.1

param(
    [switch]$Wait
)

$ErrorActionPreference = "Stop"

# Load .env
$envFile = Join-Path $PSScriptRoot ".env"
$vars = @{}
Get-Content $envFile | ForEach-Object {
    if ($_ -match '^\s*([^#=]+)=\s*"(.+)"\s*$') {
        $vars[$matches[1]] = $matches[2]
    }
}

$baseUrl = $vars["SUPABASE_URL"].TrimEnd('/')
$apiKey = $vars["SUPABASE_SERVICE_ROLE_KEY"]
$headers = @{
    "apikey"        = $apiKey
    "Authorization" = "Bearer $apiKey"
    "Content-Type"  = "application/json"
}

$tables = @(
    @{ Table = "upload_links";     Filter = "id=neq.00000000-0000-0000-0000-000000000000" }
    @{ Table = "preview_images";   Filter = "id=neq.00000000-0000-0000-0000-000000000000" }
    @{ Table = "pipeline_states";  Filter = "file_hash=neq." }
    @{ Table = "upload_journal";   Filter = "id=neq.00000000-0000-0000-0000-000000000000" }
    @{ Table = "recordings";       Filter = "id=neq.00000000-0000-0000-0000-000000000000" }
)

$now = (Get-Date).ToString("yyyy-MM-dd HH:mm:ss")
Write-Host "[$now] Purging all upload data from Supabase..."

foreach ($t in $tables) {
    $url = "$baseUrl/rest/v1/$($t.Table)?$($t.Filter)"
    Write-Host "  -> DELETE $($t.Table)... " -NoNewline
    try {
        $resp = Invoke-RestMethod -Uri $url -Method Delete -Headers $headers -SkipHttpErrorCheck -ContentType "application/json"
        Write-Host "OK"
    } catch {
        Write-Host "FAILED: $_"
    }
    if ($Wait -and $t.Table -ne $tables[-1].Table) {
        Start-Sleep -Seconds 1
    }
}

$done = (Get-Date).ToString("yyyy-MM-dd HH:mm:ss")
Write-Host "[$done] Purge complete."
