# Login Customer via Google Auth (Mobile)

Dokumen praktis untuk login akun **customer** dari aplikasi mobile (Android/iOS) memakai Google ID token.
Ringkasan konsep ada di `flow-aplikasi.md` → "Flow Login Google (customer mobile)".

## Aturan Peran

| Akun | Cara login | Klien |
|------|-----------|-------|
| `customer` | Google ID token → `POST /api/auth/google` | Mobile |
| `admin` | Email + password → `POST /api/auth/login` | Web |

Admin yang mencoba `POST /api/auth/google` ditolak `403`. Customer tidak punya password (kolom `password` kosong) sehingga tidak bisa lewat `/login`.

## Prasyarat

1. **Google Cloud Console**: buat OAuth Client ID tipe **Android** (per `package name` + SHA-1) dan **iOS** (per `bundle ID`). Web client ID tidak dipakai di sini.
2. **Backend `.env`** (lihat `backend/.env.example`):

   ```env
   GOOGLE_ANDROID_CLIENT_ID=xxx.apps.googleusercontent.com
   GOOGLE_IOS_CLIENT_ID=yyy.apps.googleusercontent.com
   JWT_EXPIRY=1h
   REFRESH_EXPIRY_CUSTOMER=2160h   # 90 hari
   ```

   Jika salah satu client ID kosong, platform tersebut gagal verifikasi (`401`).
3. **Migrasi**: `000002_auth_roles_google_refresh`, `000003_refresh_token_index`, `000004_customer_profile_fields` sudah di-apply (`go run ./cmd/migrate up`).

## Alur dari Sisi HP

```
1. Google SDK (Credential Manager Android / AuthenticationServices iOS)
   → dapatkan id_token (JWT bertanda tangan Google, aud = client ID aplikasimu)
2. POST /api/auth/google { id_token, ...profil & device }
   → response { access_token, refresh_token, user }
3. Simpan di Keychain (iOS) / EncryptedSharedPreferences / Keystore (Android).
   JANGAN pakai cookie — cookie httpOnly backend hanya untuk browser web.
4. Setiap request privat: header Authorization: Bearer <access_token>
5. access_token hanya 1 jam. Saat API jawab 401 invalid or expired token:
   POST /api/auth/refresh { refresh_token } → dapat pasangan baru
   (token lama otomatis di-revoke = rotasi).
6. Logout: POST /api/auth/logout → refresh di-revoke di server + hapus token lokal.
```

## 1. Google Login

`POST /api/auth/google` (publik, tanpa auth).

Request:

```json
{
  "id_token": "eyJhbGciOiJSUzI1NiIs...",
  "username": "budi Santoso",
  "photo_url": "https://lh3.googleusercontent.com/a/...",
  "gender": "male",
  "birthdate": "1998-05-20",
  "country": "ID",
  "device_brand": "Samsung",
  "device_model": "Galaxy S24",
  "os_version": "Android 15",
  "platform": "android"
}
```

Hanya `id_token` yang wajib. Field profil opsional: identity (`username`, `photo_url`, `gender`, `birthdate` format `YYYY-MM-DD`) dipakai saat akun baru dibuat; `country` + info device di-refresh setiap login.

Response `200 OK`:

```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "refresh_token": "xvH7...64 karakter acak...",
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "expires_in": 3600,
  "refresh_expires_in": 7776000,
  "user": {
    "id": 7,
    "email": "budi@gmail.com",
    "username": "budi",
    "full_name": "budi Santoso",
    "avatar_url": "https://lh3.googleusercontent.com/a/...",
    "role": "customer",
    "auth_provider": "google",
    "gender": "male",
    "birthdate": "1998-05-20T00:00:00Z",
    "country": "ID",
    "device_brand": "Samsung",
    "device_model": "Galaxy S24",
    "os_version": "Android 15",
    "platform": "android"
  }
}
```

Field `token` = duplikat `access_token` untuk kompatibilitas klien web lama; HP cukup pakai `access_token` + `refresh_token`.

### Logika pencocokan akun (berurutan)

1. Cari `google_sub` (ID unik Google) → ketemu = login biasa, refresh data device/country.
2. Tidak ketemu → cari `email` → ketemu = **link** akun lama ke Google `sub` ini (kecuali role `admin` → `403`).
3. Tidak ketemu juga → buat `User{role: customer, auth_provider: google}`, username dari prefix email (`budi@gmail.com` → `budi`), tambah suffix `-1`, `-2`, ... jika sudah dipakai.

## 2. Refresh Session (rotasi)

`POST /api/auth/refresh` (publik). Body `{ "refresh_token": "..." }` — alternatifnya cookie `refresh_token` (untuk web).

Response `200 OK`: pasangan token **baru** dengan format sama seperti login. Token refresh lama langsung di-revoke — simpan yang baru, buang yang lama. Refresh customer berlaku 90 hari; yang kedaluwarsa/di-revoke menjawab `401 refresh token expired or revoked`.

## 3. Request Terautentikasi

```
GET /api/auth/me
Authorization: Bearer <access_token>
```

Middleware menerima header `Bearer` dulu, lalu fallback cookie `token` (untuk web). Klaim JWT berisi `user_id`, `username`, `role`.

Endpoint yang relevan untuk customer:

| Method | Path | Auth | Keterangan |
|--------|------|------|------------|
| GET | `/api/auth/me` | Bearer | Profil sendiri |
| PATCH | `/api/users/me` | Bearer | Update `full_name`, `username`, `avatar_url` |
| POST | `/api/auth/logout` | Bearer | Revoke refresh + hapus cookie |

`GET /api/users` dan `POST /api/users` khusus `admin` (`403 forbidden: insufficient role` untuk customer).

## 4. Logout

`POST /api/auth/logout` dengan header `Bearer`. Backend me-revoke refresh token yang sedang dipakai (diambil dari body atau cookie) lalu menghapus cookie. HP wajib menghapus `access_token` + `refresh_token` lokal — JWT access yang lama tetap valid sampai 1 jam, tapi tanpa refresh token sesi tidak bisa diperpanjang.

## Tabel Error

| Status | `error` | Penyebab umum |
|--------|---------|---------------|
| 400 | validasi binding | `id_token` kosong / JSON tidak valid |
| 401 | `invalid google id_token: ...` | token palsu/kedaluwarsa, `aud` bukan client ID terdaftar, email Google belum terverifikasi, `GOOGLE_*_CLIENT_ID` kosong |
| 401 | `invalid or expired token` | access token lewat 1 jam → panggil `/refresh` |
| 401 | `invalid refresh token` / `refresh token expired or revoked` | refresh salah, sudah dipakai (rotasi), logout, atau lewat 90 hari → login ulang via Google |
| 403 | `admin must login with email and password` | akun admin mencoba jalur Google |
| 409 | `email or username already exists` | bentrok saat pembuatan akun (jarang, ada retry suffix username) |

## Testing

`id_token` tidak bisa dipalsukan via curl — harus JWT asli bertanda tangan Google dengan `aud` = client ID-mu. Cara termudah:

1. Ambil token asli dari aplikasi (log temporer `id_token` saat login Google di HP dev).
2. Import Postman collection `postman/customer-google-auth.postman_collection.json`.
3. Isi variable `id_token`, lalu jalankan berurutan: **Google Login → Get Me → Refresh Session → Logout**. Script `Tests` otomatis menyimpan pasangan token baru ke variable (penting karena rotasi).

Jangan commit `id_token`/token asli ke repo.
