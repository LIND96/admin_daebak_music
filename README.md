# admin_daebak_music

Web application (Next.js + Gin) with cookie-based JWT auth and PostgreSQL.

- **Backend**: Go (Gin) API + PostgreSQL, JWT via httpOnly cookie, port `8080`.
- **Frontend**: Next.js 16 (TailAdmin dashboard template) with public landing page, port `3000`.

## Structure

- `backend/` — Go (Gin) API, `golang-migrate` migrations
- `frontend/` — Next.js (TailAdmin), `src/proxy.ts` route guard
- `doc/` — Documentation (Indonesian)
- `docker-compose.yml` — Docker orchestration (`db`, `backend`, `frontend`)
- `.gitignore` — root ignore rules (secrets, build output, dependencies)

## Routes (frontend)

| Path | Access | Description |
|------|--------|-------------|
| `/` | public | Landing page |
| `/signin`, `/signup` | guest | Auth pages (`/login`, `/register` redirect here) |
| `/dashboard`, `/users`, `/profile` | login | Admin area |

Auth uses an httpOnly `token` cookie set by the backend (`credentials: "include"` on every fetch). See `doc/flow-aplikasi.md`.

## Quick Start

```bash
# Docker (from repo root)
docker compose up --build
docker compose exec backend /migrate up   # compose does NOT auto-migrate

# Local dev (two terminals)
cd backend && go run ./cmd/migrate up && go run ./cmd/server   # :8080
cd frontend && npm install && npm run dev                        # :3000
```

Details: `doc/cara-menjalankan.md`, `doc/migrasi-database.md`, `doc/konfigurasi-docker.md`.
