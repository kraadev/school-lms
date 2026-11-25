$ErrorActionPreference = "Stop"
$Version = "v1.13.6"
$ZipUrl = "https://github.com/livekit/livekit/releases/download/$Version/livekit_$( $Version.Substring(1) )_windows_amd64.zip"
$ZipFile = "livekit-server.zip"
$ExeFile = "livekit-server.exe"

if (-not (Test-Path $ExeFile)) {
    Write-Host "Mengunduh LiveKit Server $Version..."
    Invoke-WebRequest -Uri $ZipUrl -OutFile $ZipFile
    Write-Host "Mengekstrak..."
    Expand-Archive -Path $ZipFile -DestinationPath . -Force
    Remove-Item -Path $ZipFile
    Write-Host "LiveKit Server berhasil diunduh!"
}

# Dapatkan IP lokal Wi-Fi secara otomatis
$localIP = (Get-NetIPAddress -AddressFamily IPv4 | Where-Object {
    $_.InterfaceAlias -like '*Wi-Fi*' -or $_.InterfaceAlias -like '*WLAN*'
} | Select-Object -First 1).IPAddress

if (-not $localIP) {
    $localIP = "localhost"
}

Write-Host "Menjalankan LiveKit Server dalam mode Dev..."
Write-Host "Bind     : 0.0.0.0:7880 (semua interface)"
Write-Host "URL LAN  : ws://${localIP}:7880"
Write-Host "API Key  : devkey"
Write-Host "API Secret: secret"
Write-Host "Tekan Ctrl+C untuk berhenti."

# Jalankan server dengan bind ke semua interface agar bisa diakses dari LAN/HP
.\livekit-server.exe --dev --bind 0.0.0.0
