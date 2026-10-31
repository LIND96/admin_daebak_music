# Konfigurasi Docker

## File Terkait

- `docker-compose.yml` (root) — orkestrasi service `db`, `backend`, `frontend`.
- `backend/Dockerfile` — build image backend Go.
- `frontend/Dockerfile` — build image frontend Next.js.

## Menjalankan dengan Docker Compose

Dari root proyek:

```bash
docker compose up --build
```

Service yang dijalankan:

- `db`: PostgreSQL 16 (port 5432)
- `backend`: API Go (port 8080); migrasi dijalankan manual via `docker compose exec backend /migrate up`
- `frontend`: Next.js (port 3000, standalone output)

Akses aplikasi di `http://localhost:3000`.

## Menghentikan

```bash
docker compose down
```

Untuk menghapus volume database:

```bash
docker compose down -v
```

## Variabel Lingkungan

### Backend

| Variabel | Default | Keterangan |
|----------|---------|------------|
| `PORT` | 8080 | Port server |
| `DB_HOST` | db | Host database |
| `DB_PORT` | 5432 | Port database |
| `DB_USER` | postgres | User database |
| `DB_PASSWORD` | postgres | Password database |
| `DB_NAME` | admin_daebak_music | Nama database |
| `DB_SSLMODE` | disable | SSL mode |
| `JWT_SECRET` | change-me-in-production | Secret JWT |
| `JWT_EXPIRY` | 24h | Masa berlaku token |
| `FRONTEND_URL` | http://localhost:3000 | Origin frontend (CORS) |
| `COOKIE_SECURE` | false | `true` di produksi (wajib HTTPS) |

### Frontend

| Variabel | Default | Keterangan |
|----------|---------|------------|
| `NEXT_PUBLIC_API_URL` | http://localhost:8080 | URL backend API |

## Catatan

- `docker-compose.yml` meng-override command backend menjadi hanya `/server`, jadi migrasi TIDAK otomatis: jalankan `docker compose exec backend /migrate up` setelah service menyala. (Image `backend/Dockerfile` sendiri defaultnya `/migrate up && /server`.)
- Pastikan port 3000, 8080, dan 5432 tidak digunakan proses lain.
- Untuk produksi, ganti `JWT_SECRET` dengan nilai yang kuat dan aman.