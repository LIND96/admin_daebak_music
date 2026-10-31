# Cara Menjalankan Aplikasi `admin_daebak_music`

Proyek ini terdiri dari dua bagian:

- **Backend**: Go (Gin) + PostgreSQL, berjalan di port `8080`.
- **Frontend**: Next.js (TailAdmin template), berjalan di port `3000`.

## Prasyarat

- Go 1.22+
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

## 4. Alur Penggunaan

1. Buka `http://localhost:3000` (landing page publik).
2. Register akun baru (`/signup`) atau login (`/signin`), atau klik **Get Started**.
3. Setelah login, masuk ke dashboard admin (`/dashboard`), kelola lewat sidebar.
4. Akses halaman **Users** (`/users`) untuk melihat daftar pengguna.
5. Akses halaman **Profile** (`/profile`) untuk melihat dan mengubah profil.

## Endpoint API Utama

| Method | Path | Keterangan |
|--------|------|------------|
| POST | `/api/auth/register` | Registrasi user |
| POST | `/api/auth/login` | Login, mengembalikan JWT |
| POST | `/api/auth/logout` | Logout (stateless) |
| GET | `/api/auth/me` | Profil user saat ini |
| GET | `/api/users` | Daftar semua user |
| GET | `/api/users/:id` | Detail user |
| PATCH | `/api/users/me` | Update profil sendiri |

Autentikasi menggunakan header `Authorization: Bearer <token>`.