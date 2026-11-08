# Cara Menjalankan Aplikasi `admin_daebak_music`

Proyek ini terdiri dari dua bagian:

- **Backend**: Go (Gin) + PostgreSQL, berjalan di port `8080`.
- **Frontend**: Next.js (TailAdmin template), berjalan di port `3000`.

## Prasyarat

- Go 1.26+
- Node.js 20+
- PostgreSQL 16+
- npm

## 1. Menjalankan Database PostgreSQL

Pastikan PostgreSQL berjalan. Buat database:

```bash
createdb admin_daebak_music
```

## 2. Menjalankan Backend

```bash
cd backend
cp .env.example .env
# Sesuaikan nilai pada .env (DB_USER, DB_PASSWORD, dll.)
go mod download
go run ./cmd/migrate up      # jalankan migrasi
go run ./cmd/server          # jalankan server
```

Backend akan berjalan di `http://localhost:8080`.

## 3. Menjalankan Frontend

```bash
cd frontend
cp .env.local.example .env.local
npm install
npm run dev
```

Frontend akan berjalan di `http://localhost:3000`.

> Dengan Docker + Caddy, akses utama adalah `http://localhost` (port 80):
> `/api/*` diteruskan ke backend, sisanya ke frontend (lihat `doc/konfigurasi-docker.md`).
> Produksi: `DOMAIN=app.example.com docker compose up --build` untuk HTTPS otomatis.

## 4. Alur Penggunaan

1. Bootstrap admin pertama (sekali saja, butuh `BOOTSTRAP_SECRET` dari `.env`):
   `POST /api/auth/bootstrap-admin` dengan `{ email, username, password, setup_key }`.
2. Buka `http://localhost:3000` dan login sebagai admin (`/signin`). Register publik dimatikan.
3. Setelah login, masuk ke dashboard admin (`/dashboard`), kelola lewat sidebar.
4. Akses halaman **Users** (`/users`) untuk melihat/menambah pengguna (admin only).
5. Akses halaman **Profile** (`/profile`) untuk melihat dan mengubah profil.
6. Customer mobile login via `POST /api/auth/google` dengan `id_token` dari Google SDK di HP.

## Endpoint API Utama

| Method | Path | Keterangan |
|--------|------|------------|
| POST | `/api/auth/bootstrap-admin` | Buat admin pertama (sekali pakai, pakai `setup_key`) |
| POST | `/api/auth/login` | Login admin web (email+password, `role=admin` only) |
| POST | `/api/auth/google` | Login customer mobile (Google `id_token`) |
| POST | `/api/auth/refresh` | Rotasi refresh token (admin 24h, customer 90 hari) |
| POST | `/api/auth/logout` | Logout + revoke refresh, hapus cookie |
| GET | `/api/auth/me` | Profil user saat ini |
| GET | `/api/users` | Daftar semua user (admin only) |
| GET | `/api/users/:id` | Detail user (admin only) |
| POST | `/api/users` | Buat user baru (admin only, pengganti register) |
| PATCH | `/api/users/me` | Update profil sendiri |
| PATCH | `/api/users/me/password` | Ganti password sendiri (perlu password lama, revoke semua refresh token) |

Autentikasi memakai JWT access pendek + refresh token terpisah (rotasi). Web memakai cookie httpOnly `token` + `refresh_token` (dikirim otomatis via `credentials: "include"`); mobile memakai header `Authorization: Bearer <access_token>`.