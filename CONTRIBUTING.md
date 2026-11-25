# Panduan kontribusi

Terima kasih telah mempertimbangkan kontribusi untuk School Space.

## Sebelum mengirim perubahan

1. Buat perubahan yang fokus dan jelaskan masalah atau kebutuhan yang diselesaikan.
2. Jangan sertakan file `.env`, kredensial, sertifikat/private key, unggahan pengguna, atau data pribadi.
3. Jalankan pemeriksaan yang relevan:
   - Backend: `go test ./...` dari direktori `backend`.
   - Frontend: `npm run build` dari direktori `frontend`.
4. Sertakan ringkasan perubahan, langkah pengujian, dan tangkapan layar untuk perubahan UI bila relevan.

## Pull request

- Buka pull request dengan deskripsi yang jelas dan perubahan sekecil mungkin.
- Tandai perubahan yang memengaruhi konfigurasi, database, atau perilaku API.
- Jangan menambahkan dependency atau mengubah lisensi/aset tanpa alasan dan hak penggunaan yang jelas.
- Untuk kerentanan keamanan, ikuti [`SECURITY.md`](SECURITY.md); jangan membukanya sebagai issue publik.
