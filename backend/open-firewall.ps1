# Jalankan script ini sebagai Administrator
# Klik kanan open-firewall.ps1 → Run with PowerShell (as Administrator)

Write-Host "Membuka port firewall untuk School LMS..." -ForegroundColor Cyan

# Hapus rule lama jika ada
netsh advfirewall firewall delete rule name="LiveKit-7880" 2>$null
netsh advfirewall firewall delete rule name="Backend-8080" 2>$null
netsh advfirewall firewall delete rule name="Nuxt-3000" 2>$null

# Tambah rule baru
netsh advfirewall firewall add rule name="LiveKit-7880" dir=in action=allow protocol=TCP localport=7880
netsh advfirewall firewall add rule name="Backend-8080" dir=in action=allow protocol=TCP localport=8080
netsh advfirewall firewall add rule name="Nuxt-3000" dir=in action=allow protocol=TCP localport=3000

Write-Host ""
Write-Host "Selesai! Port berikut sudah dibuka:" -ForegroundColor Green
Write-Host "  - 7880 (LiveKit WebSocket)" -ForegroundColor Green
Write-Host "  - 8080 (Backend API)"       -ForegroundColor Green
Write-Host "  - 3000 (Frontend Nuxt)"     -ForegroundColor Green
Write-Host ""
Write-Host "Tekan Enter untuk menutup..."
Read-Host
