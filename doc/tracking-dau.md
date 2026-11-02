# Tracking Daily Active User (DAU) — Integrasi Mobile & Dashboard

Dokumen integrasi untuk endpoint tracking anonim daily-use dan dashboard admin.
Backend: `backend/internal/handlers/tracking.go`, tabel `daily_app_usages`
(migrasi `000005_daily_app_usages`, `000006_daily_app_usages_country_code`).

## Konsep

- **Anonim**: `POST /api/tracking/daily-use` publik, tanpa auth/JWT.
- **Setiap hit = 1 baris**. Post 2x sehari dengan `device_id` + `apps_name`
  yang sama = 2 baris. Tidak ada unique constraint.
- **DAU = `COUNT(DISTINCT device_id)`** per `apps_name` + `usage_date`.
  Dihitung saat baca (`GET /api/tracking/stats/daily`), bukan saat tulis.
- **Zona waktu dikunci `Asia/Jakarta`**. `usage_date` diisi server dari
  `now.In(Asia/Jakarta).Format("2006-01-02")`, batas ganti hari ikut 00:00 WIB.
- **Email opsional**, hanya pelengkap anonim. `device_id` + `apps_name` wajib.

## 1. Record Hit (Mobile → Backend)

`POST /api/tracking/daily-use` (publik).

Request:

```json
{
  "device_id": "abc-53342",
  "apps_name": "BTS ARMY",
  "android_version": "14",
  "device_brand": "Samsung",
  "device_model": "Galaxy S24",
  "country": "Ukraine",
  "country_code": "UA",
  "email": "user@mail.com"
}
```

Aturan field:

| Field | Wajib | Normalisasi server | Keterangan |
|-------|-------|--------------------|------------|
| `device_id` | ya | `TrimSpace` | ID unik perangkat (SSAID / `Settings.Secure.ANDROID_ID` / UUID vendor). |
| `apps_name` | ya | `lowercase + trim` | Contoh `bts army` (disimpan lowercase). Beda app = baris/DAU terpisah. |
| `android_version` | tidak | `trim` | Contoh `14`, `Android 15`. |
| `device_brand` | tidak | `trim`, maks 100 char | Brand HP, contoh `Samsung`, `Xiaomi`, `Google`. |
| `device_model` | tidak | `trim`, maks 100 char | Model HP, contoh `Galaxy S24`, `Pixel 8`. |
| `country` | tidak | `trim`, maks 100 char | Nama negara dari ipwho.is (`Ukraine`). Untuk label dashboard. |
| `country_code` | tidak | `UPPER + trim`, maks 10 char | **Kode ISO dari ipwho.is `country_code`** (`UA`). Untuk marker peta + grouping. |
| `email` | tidak | `trim`, validasi mengandung `@` | Opsional, boleh kosong. |

Response `201`:

```json
{ "status": "recorded", "event": { "id": 12, "usage_date": "2026-11-03", "...": "..." } }
```

Error umum:

| Status | Penyebab |
|--------|----------|
| 400 | `device_id` / `apps_name` kosong, `country` > 100 char, `country_code` > 10 char, `device_brand` / `device_model` > 100 char, `email` tanpa `@` |
| 500 | `failed to record usage` (DB down) |

### Kapan menembak dari HP

- Panggil **setiap app dibuka / foreground** (misal `onCreate` + `onResume` pertama
  per sesi). Tidak perlu logika "sekali sehari" di client — server tetap
  mencatat semua hit, DAU-nya yang distinct.
- Cache `country_code` (lihat §2), jangan panggil ipwho.is setiap hit.
- Fire-and-forget dengan timeout pendek (misal 5 detik); kegagalan tracking
  tidak boleh mengganggu UX. Retry maksimal 1x atau abaikan.

Contoh Kotlin (Android):

```kotlin
val body = JSONObject().apply {
  put("device_id", Settings.Secure.getString(cr, Settings.Secure.ANDROID_ID))
  put("apps_name", "BTS ARMY")
  put("android_version", Build.VERSION.RELEASE)
  put("device_brand", Build.MANUFACTURER) // "samsung" → normalisasi di server? trim saja
  put("device_model", Build.MODEL) // "SM-S921B"
  put("country", cachedCountry) // "Ukraine", dari ipwho.is
  put("country_code", cachedCountryCode) // "UA", dari ipwho.is
}
// POST https://api-mu/api/tracking/daily-use (Content-Type: application/json)
```

Contoh curl:

```bash
curl -X POST http://localhost:8080/api/tracking/daily-use \
 -H 'Content-Type: application/json' \
 -d '{"device_id":"abc-53342","apps_name":"BTS ARMY","android_version":"14","device_brand":"Samsung","device_model":"Galaxy S24","country":"Ukraine","country_code":"UA"}'
```

## 2. Mengisi `country` + `country_code` via ipwho.is

Kirim **keduanya**: `country` = nama, `country_code` = ISO 3166-1 alpha-2.

```
GET https://ipwho.is/
→ { "country": "Ukraine", "country_code": "UA", "city": "Kyiv",
    "latitude": 50.45, "longitude": 30.52,
    "timezone": { "id": "Europe/Kyiv" }, ... }
```

| Field ipwho.is | Dikirim sebagai | Dipakai untuk |
|----------------|-----------------|---------------|
| `country_code` | `country_code` (`UA`) | Marker peta + grouping `by_country` (stabil, selalu 2 huruf). |
| `country` | `country` (`Ukraine`) | Label dashboard. Fallback grouping untuk data lama yang belum ada `country_code`. |
| `city` / `region` | — (belum disimpan) | Opsional, butuh kolom baru jika ingin breakdown kota. |
| `latitude` / `longitude` | — | Opsional, saat ini peta memakai titik tengah negara. |
| `timezone.id` | — | Info saja, `usage_date` tetap ditentukan server (WIB). |

Pola disarankan: saat app start, `GET https://ipwho.is/ → ambil country +
country_code → simpan di SharedPreferences/DataStore → pakai untuk semua hit
hari itu`. Fallback kosongkan keduanya jika request gagal — hit tetap dikirim,
masuk `UNKNOWN`.

> Data lama yang hanya punya `country` (= kode seperti `ID`) tetap didukung:
> `by_country` fallback ke `UPPER(country)` saat `country_code` kosong, dan
> dashboard memetakan alias (`AMERIKA`/`USA`/`US` → USA).

## 3. Statistik Harian (Dashboard Admin)

`GET /api/tracking/stats/daily?date=YYYY-MM-DD&apps_name=` — butuh JWT admin
(cookie `token` atau `Authorization: Bearer`), `RequireRole("admin")`.

- `date` default hari ini WIB jika kosong. Format wajib `YYYY-MM-DD`.
- `apps_name` opsional (lowercase). Jika diisi, `hourly`/`by_country`/total
  difilter 1 app; `by_app` selalu semua app di tanggal itu (untuk perbandingan).

Response `200`:

```json
{
  "date": "2026-11-03",
  "timezone": "Asia/Jakarta",
  "apps_name": "bts army",
  "total_hits": 8,
  "dau": 3,
  "hourly": [{ "hour": 0, "hits": 0, "unique_devices": 0 }, "...24 item"],
  "by_country": [{ "country": "Ukraine", "country_code": "UA", "hits": 5, "unique_devices": 2 }],
  "by_app": [{ "apps_name": "bts army", "hits": 8, "unique_devices": 3 }]
}
```

Query (semua `Raw SQL`, `WHERE usage_date = ?`):

- hourly: `EXTRACT(HOUR FROM occurred_at AT TIME ZONE 'Asia/Jakarta')`,
  `GROUP BY 1` → 24 bucket (jam kosong = 0).
- total/`dau`: `COUNT(*)` + `COUNT(DISTINCT device_id)`.
- `by_country`: `COALESCE(NULLIF(UPPER(country),''),'UNKNOWN')`, top 10.
- `by_app`: `GROUP BY apps_name`, top 20.

## 4. Dashboard (`/dashboard`)

Komponen: `frontend/src/components/main/dashboard/tracking/`.

- `tracking-section.tsx` — filter tanggal (default hari ini WIB via
  `Intl.DateTimeFormat` + `timeZone: Asia/Jakarta`, tombol ←/→, input date)
  + filter `apps_name` + 3 kartu (DAU unik `device_id`, total hits, rata-rata hit/device).
- `hourly-chart.tsx` — bar 24 jam (`00–23` WIB), series user unik + hits.
- `countries-chart.tsx` — gaya Customers Demographic: peta dunia (jsvectormap
  reuse `demographic-card/country-map`) + list 5 teratas (nama, user unik · hits, %).
  Marker dari lookup `COUNTRY_COORDS`; kode tak dikenal tetap di list tanpa marker.
- `apps-chart.tsx` — horizontal bar distributed (1 warna per `apps_name`),
  hanya series user unik.

Satu filter tanggal me-refresh semua grafik (state diangkat ke `TrackingSection`).

## 5. Migrasi & Verifikasi

```bash
cd backend && go run ./cmd/migrate up   # membuat daily_app_usages
go run ./cmd/server                      # AutoMigrate juga mencakup AppUsageEvent
```

Verifikasi:

```bash
# 2 hit device sama
curl -X POST :8080/api/tracking/daily-use -H 'Content-Type: application/json' \
 -d '{"device_id":"d1","apps_name":"daebak-music","country":"ID"}'
curl -X POST :8080/api/tracking/daily-use -H 'Content-Type: application/json' \
 -d '{"device_id":"d1","apps_name":"daebak-music","country":"ID"}'
# GET stats (ganti token admin) → total_hits=2, dau=1
```

```sql
SELECT usage_date, apps_name, COUNT(*), COUNT(DISTINCT device_id)
FROM daily_app_usages GROUP BY 1,2 ORDER BY 1 DESC;
```
