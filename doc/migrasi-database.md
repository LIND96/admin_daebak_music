# Migrasi Database & Setup Database

## Teknologi

- Database: PostgreSQL 16
- Migration tool: `golang-migrate` (dijalankan melalui `cmd/migrate`)

## File Migrasi

File migrasi berada di `backend/migrations/`:

- `000001_create_users.up.sql` — membuat tabel `users`
- `000001_create_users.down.sql` — menghapus tabel `users`

## Setup Database (Manual)

1. Pastikan PostgreSQL berjalan.
2. Buat database:

   ```bash
   createdb admin_daebak_music
   ```

3. Sesuaikan `backend/.env`:

   ```env
   DB_HOST=localhost
   DB_PORT=5432
   DB_USER=postgres
   DB_PASSWORD=postgres
   DB_NAME=admin_daebak_music
   DB_SSLMODE=disable
   ```

## Menjalankan Migrasi

Dari folder `backend`:

```bash
# Terapkan semua migrasi (up)
go run ./cmd/migrate up

# Rollback satu langkah (down)
go run ./cmd/migrate down
```

## Catatan

- Backend juga menjalankan `AutoMigrate` GORM saat startup, namun migrasi SQL adalah sumber utama skema.
- Untuk menambah migrasi baru, buat file dengan penomoran berikutnya, contoh:
  - `000002_add_column.up.sql`
  - `000002_add_column.down.sql`
- Pastikan setiap migrasi `up` memiliki pasangan `down` agar dapat di-rollback.