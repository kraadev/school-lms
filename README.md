<div align="center">
  <h1>School Space</h1>
  <p><strong>Learning management system untuk menghubungkan sekolah, guru, dan siswa dalam satu ruang belajar digital.</strong></p>
  <p>
    <img alt="Go" src="https://img.shields.io/badge/Backend-Go-00ADD8?logo=go&logoColor=white">
    <img alt="Nuxt" src="https://img.shields.io/badge/Frontend-Nuxt-00DC82?logo=nuxt.js&logoColor=white">
    <img alt="Database" src="https://img.shields.io/badge/Database-PostgreSQL-4169E1?logo=postgresql&logoColor=white">
  </p>
</div>

School Space adalah aplikasi LMS berbasis web. Backend menyediakan REST API dan komunikasi real-time; frontend menyediakan antarmuka untuk administrasi sekolah dan kegiatan belajar.

## Fitur

- Manajemen akun dan peran admin, guru, serta siswa.
- Pengelolaan tahun ajaran, semester, kelas, mata pelajaran, dan agenda akademik.
- Materi, tugas, pengumpulan tugas, kuis, dan presensi.
- Notifikasi, percakapan, dan pusat bantuan.
- Pertemuan daring dengan LiveKit dan WebRTC.

## Teknologi

| Bagian | Teknologi |
| --- | --- |
| Frontend | Nuxt, Vue, TypeScript, Tailwind CSS |
| Backend | Go, Gin, GORM |
| Database | PostgreSQL |
| Komunikasi | WebSocket, LiveKit, WebRTC |

## Struktur proyek

```text
.
├── backend/
│   ├── cmd/server/       # Entry point API
│   ├── internal/         # Konfigurasi, model, handler, dan layanan
│   └── migrations/       # SQL schema/migration
├── frontend/
│   ├── app/              # Halaman dan komponen Nuxt
│   └── public/           # Aset statis aplikasi
└── go.work
```

## Menjalankan secara lokal

### Prasyarat

- Go versi yang tercantum di [`backend/go.mod`](backend/go.mod).
- Node.js dan npm yang kompatibel dengan versi Nuxt di [`frontend/package.json`](frontend/package.json).
- PostgreSQL yang sedang berjalan.
- LiveKit Server untuk menguji fitur pertemuan daring.

### 1. Siapkan konfigurasi lokal

Salin berkas contoh, lalu isi nilai yang sesuai untuk mesin Anda. Jangan commit berkas `.env`.

```powershell
Copy-Item backend\.env.example backend\.env
Copy-Item frontend\.env.example frontend\.env
```

Atur koneksi database di `backend/.env`, buat database yang namanya sesuai dengan `DB_NAME`, dan isi `JWT_SECRET` serta `ADMIN_INITIAL_PASSWORD` dengan nilai unik. Backend membuat tabel yang diperlukan saat pertama kali dijalankan. Akun awal `admin` hanya dibuat jika belum ada; kata sandinya berasal dari `ADMIN_INITIAL_PASSWORD`.

`frontend/.env` mengatur URL API dan konfigurasi ICE WebRTC. Konfigurasi ICE/TURN produksi harus disediakan oleh operator deployment.

### 2. Jalankan backend

Dari direktori `backend`:

```powershell
go run ./cmd/server
```

API menggunakan port `APP_PORT` (default `8080`).

### 3. Jalankan frontend

Buka terminal lain dari direktori `frontend`:

```powershell
npm ci
npm run dev
```

Buka URL yang ditampilkan oleh Nuxt di terminal.

### LiveKit lokal

Skrip `backend/start-livekit.ps1` menjalankan LiveKit dalam mode development untuk pengujian lokal. Mode ini memakai kredensial bawaan development dan bind ke semua interface; **jangan gunakan atau ekspos konfigurasi tersebut di jaringan publik/produksi**. Untuk deployment, siapkan LiveKit dan kredensial sendiri melalui environment.

## Pengembangan dan pengujian

Jalankan pemeriksaan backend dari `backend`:

```powershell
go test ./...
```

Perintah frontend yang tersedia:

```powershell
npm run build
```

## Keamanan dan data lokal

- Simpan rahasia hanya di `.env` lokal atau secret manager. Jangan masukkan token, kata sandi, private key, sertifikat lokal, maupun data pengguna ke commit.
- Unggahan di `backend/uploads/`, sertifikat di `frontend/certs/`, hasil build, executable, log, dan file environment lokal dikecualikan oleh `.gitignore`.
- Sebelum publikasi, periksa perubahan dan riwayat Git. `.gitignore` tidak menghapus file yang sudah ter-track atau rahasia dari commit sebelumnya; jika rahasia pernah dipublikasikan, cabut/rotasi rahasia tersebut dan bersihkan riwayat dengan prosedur yang sesuai.
- Pastikan semua gambar, video, dan aset lain yang memang akan dipublikasikan memiliki hak penggunaan yang sesuai.
- Laporkan dugaan kerentanan secara privat; lihat [`SECURITY.md`](SECURITY.md).

## Kontribusi

Kontribusi, laporan bug, dan usulan perbaikan sangat dihargai. Baca [`CONTRIBUTING.md`](CONTRIBUTING.md) sebelum mengirim perubahan.

## Lisensi

Lisensi pada tingkat repository belum ditetapkan secara eksplisit. [`backend/LICENSE`](backend/LICENSE) hanya berada di direktori backend; konfirmasikan hak dan lisensi untuk seluruh proyek beserta asetnya sebelum mendistribusikan atau menerima kontribusi publik.
