# Demo script for the TCP/UDP showcase. Launches a couple of CLI clients in
# separate PowerShell windows, then drives a publish through the gateway.
#
# Prereqs:
#   - docker compose up (or run gateway + tcp-server + udp-server manually)
#   - One user already seeded; pass the user id with -UserId
#   - One manga id; pass with -MangaId
#
# Usage:
#   .\scripts\demo-tcp-udp.ps1 -UserId 665a... -MangaId 67120... -Token "Bearer eyJ..."

param(
    [Parameter(Mandatory = $true)] [string] $UserId,
    [Parameter(Mandatory = $true)] [string] $MangaId,
    [Parameter(Mandatory = $true)] [string] $Token,
    [string] $BaseUrl = "http://localhost:8080",
    [string] $TcpAddr = "localhost:9000",
    [string] $UdpAddr = "localhost:9001"
)

$repo = Split-Path -Parent $PSScriptRoot

Write-Host "1) Spawning 2 TCP clients (subscribed to $UserId)..."
Start-Process powershell -ArgumentList @(
    "-NoExit", "-Command",
    "cd `"$repo`"; go run ./cmd/tcpclient -addr=$TcpAddr -user=$UserId"
)
Start-Process powershell -ArgumentList @(
    "-NoExit", "-Command",
    "cd `"$repo`"; go run ./cmd/tcpclient -addr=$TcpAddr -user=$UserId"
)

Write-Host "2) Spawning 2 UDP clients (REGISTERed, ACK enabled)..."
Start-Process powershell -ArgumentList @(
    "-NoExit", "-Command",
    "cd `"$repo`"; go run ./cmd/udpclient -server=$UdpAddr -id=desktop-01 -ack"
)
Start-Process powershell -ArgumentList @(
    "-NoExit", "-Command",
    "cd `"$repo`"; go run ./cmd/udpclient -server=$UdpAddr -id=desktop-02 -ack"
)

Start-Sleep -Seconds 3
Write-Host "3) Triggering progress update via HTTP -> should fan out via TCP"
$body = @{ status = "reading"; current_chapter = (Get-Random -Min 1 -Max 200) } | ConvertTo-Json
Invoke-RestMethod -Method Put `
    -Uri "$BaseUrl/me/reading/$MangaId" `
    -Headers @{ Authorization = $Token; "Content-Type" = "application/json" } `
    -Body $body | Out-Host

Start-Sleep -Seconds 2
Write-Host "4) Triggering admin notify -> should fan out via UDP"
$notifyBody = @{ user_id = $UserId; content = "Demo chapter notification!" } | ConvertTo-Json
Invoke-RestMethod -Method Post `
    -Uri "$BaseUrl/admin/notify" `
    -Headers @{ Authorization = $Token; "Content-Type" = "application/json"; "X-Admin-Token" = $env:ADMIN_TOKEN } `
    -Body $notifyBody | Out-Host

Write-Host ""
Write-Host "Done. Check the spawned windows — TCP clients should show progress_update, UDP clients should show new_chapter."
