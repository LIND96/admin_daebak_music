# Alur Aplikasi (Flow Backend & Frontend)

Dokumen ini menjelaskan alur kerja aplikasi `admin_daebak_music`, mulai dari interaksi pengguna di frontend hingga pemrosesan di backend dan database.

## Arsitektur Singkat

```
Browser (Next.js)  --HTTP/JSON-->  Gin API (Go)  --SQL-->  PostgreSQL
       |                                |
       |  cookie httpOnly: token (JWT)   |
```

- Frontend berkomunikasi dengan backend melalui `fetch` ke `NEXT_PUBLIC_API_URL` dengan `credentials: "include"`.
- Backend memvalidasi JWT dari cookie httpOnly `token` (header `Authorization: Bearer` tetap didukung untuk kompatibilitas).
- Semua respons berbentuk JSON.

## Flow Register

Register publik **dimatikan**. Alurnya diganti dua endpoint khusus:

### Bootstrap admin pertama (sekali pakai)

1. Client `POST /api/auth/bootstrap-admin` dengan `{ email, username, password, full_name?, setup_key }`.
2. Backend tolak (`403`) jika `setup_key != BOOTSTRAP_SECRET` atau admin sudah ada.
3. Buat `User{role: admin, auth_provider: email}`, terbitkan access (1h) + refresh (24h), set cookie, response `{ access_token, refresh_token, user }`.

### Admin menambah user

1. Admin login `POST /api/users` dengan `{ email, username, password, full_name?, role: admin|customer }` + Bearer admin.
2. Handler `UsersHandler.Create` hash password dan buat user (`auth_provider: email`).

## Flow Login Admin (web)

### Frontend

1. Pengguna membuka `/signin`, mengisi email dan password.
2. Form sign-in TailAdmin memanggil `useAuth().login(email, password)`.
3. `AuthProvider.login` memanggil `api.login(...)` (POST `/api/auth/login`).
4. `AuthProvider` menolak state jika `user.role !== "admin"` (web admin-only).
5. Jika sukses, cookie httpOnly diset dan redirect ke `/dashboard`.
6. Jika gagal (termasuk akun customer), tampilkan error inline pada form.

### Backend

1. Middleware `LoginRateLimit` (`internal/middleware/ratelimit.go`) berjalan dulu:
   token bucket per IP — burst 10x, lalu ~10x/menit. Lebih dari itu dijawab
   `429 too many login attempts` + header `Retry-After: 60` (anti brute-force/DDoS).
2. Handler `AuthHandler.Login` mem-bind JSON ke `loginRequest`.
2. Cari user berdasarkan email (`Where("email = ?")`).
3. Jika `role != "admin"`, kembalikan `403 Forbidden` (customer wajib Google login).
4. Bandingkan password dengan bcrypt (`utils.CheckPassword`).
5. Generate access JWT (1h) + refresh token opaque (24h untuk admin), simpan hash di `refresh_tokens`.
6. Set cookie httpOnly `token` + `refresh_token`, response `{ access_token, refresh_token, user }`.

## Flow Login Google (customer mobile)

1. HP ambil `id_token` via Google SDK (Android/iOS), kirim `POST /api/auth/google` beserta data koleksi:
   `{ id_token, device_info?, username?, photo_url?, gender?, birthdate? (YYYY-MM-DD), country?, device_brand?, device_model?, os_version?, platform? }`.
   Contoh: merk HP (`Samsung`), versi Android (`Android 15`), negara (`ID`), gender, usia (via `birthdate`), username, foto.
2. Backend verifikasi via `utils.VerifyGoogleIDToken` (cek `aud` ke `GOOGLE_ANDROID_CLIENT_ID`/`GOOGLE_IOS_CLIENT_ID`, email terverifikasi).
3. Cari by `google_sub` → jika tidak ada, cocokkan/link by email → jika tidak ada, buat `User{role: customer, auth_provider: google}` (username dari prefix email + suffix unik).
4. Akun `role=admin` ditolak (`403`) — admin wajib email+password.
5. Terbitkan access (1h) + refresh (90 hari / 2160h untuk customer), simpan hash, response JSON untuk disimpan di Keychain/Keystore HP.

## Flow Refresh Token (rotasi)

1. Client `POST /api/auth/refresh { refresh_token }` (atau cookie `refresh_token`).
2. Backend cocokkan hash, cek belum revoked/kedaluwarsa, revoke token lama, terbitkan pasangan baru.
3. Umur refresh ditentukan role: admin 24h (`REFRESH_EXPIRY_ADMIN`), customer 90 hari (`REFRESH_EXPIRY_CUSTOMER`).

## Flow Logout

### Frontend

1. Pengguna klik "Sign out" pada user dropdown di header.
2. `AuthProvider.logout` memanggil `api.logout()` (POST `/api/auth/logout`), backend menghapus cookie, state `user = null`.
3. Redirect ke `/signin` (navbar/header berubah ke mode guest).

### Backend

- Endpoint `/api/auth/logout` dilindungi middleware Auth.
- Backend menghapus cookie (`Set-Cookie token="" MaxAge=-1`) dan mengembalikan `{ message: "logged out successfully" }`.

## Flow Menampilkan Daftar Users

### Frontend

1. Halaman `/dashboard` dan `/users` dilindungi `src/proxy.ts`; jika belum login, redirect ke `/signin` sebelum halaman render. `AuthProvider` juga menolak sesi non-admin.
2. Setelah login admin terkonfirmasi, halaman `/users` memanggil `api.listUsers()` (GET `/api/users`, admin only).
3. Render tabel users TailAdmin (avatar, nama, email, role, tanggal join).

### Backend

1. Middleware `Auth` memvalidasi JWT dari cookie `token` (fallback: header `Authorization`), lalu `RequireRole("admin")` menolak non-admin (`403`).
2. Handler `UsersHandler.List` mengambil semua user dari DB.
3. Response `{ users: [...], total }` (password tidak disertakan).

## Rate Limit Login (anti brute-force/DDoS)

File: `internal/middleware/ratelimit.go`, dipasang di `POST /api/auth/login` dan
`POST /api/auth/bootstrap-admin` (`cmd/server/main.go`).

- Token bucket per IP klien: burst 10x percobaan rapat, lalu ~10x/menit sustained.
- Melebihi batas → `429 {"error": "too many login attempts, try again later"}` +
  header `Retry-After: 60`. Form sign-in menampilkan pesan itu otomatis (inline error).
- IP dibaca via `c.ClientIP()`; `r.SetTrustedProxies(10/8, 172.16/12, 192.168/16)`
  wajib agar di belakang Caddy yang terbaca IP asli user (dari `X-Forwarded-For`),
  bukan IP container Caddy — kalau tidak, semua user berbagi 1 bucket dan saling memblokir.
- Counter in-memory (hilang saat restart, cukup untuk 1 instance). Pengguna sekantor
  di belakang 1 NAT berbagi kuota — 10 burst + 10/menit masih longgar untuk manusia.

## Flow Menampilkan & Update Profil

### Frontend

1. Halaman `/profile` memanggil `api.me()` (GET `/api/auth/me`) untuk data terbaru.
2. Form diisi dengan data profil.
3. Saat submit, panggil `api.updateProfile(...)` (PATCH `/api/users/me`).
4. Update state profil dan tampilkan pesan sukses inline pada form.

### Backend

1. Middleware `Auth` mengekstrak `userID` dari JWT.
2. Handler `UsersHandler.UpdateProfile` mencari user berdasarkan `userID`.
3. Update field yang dikirim (full_name, username, avatar_url).
4. Simpan ke DB dan response `{ user }`.

## Flow Ganti Password Sendiri

Endpoint: `PATCH /api/users/me/password` (butuh login, semua role yang punya password).

Request:

```json
{ "current_password": "secret123", "new_password": "secret-baru-123" }
```

Aturan:

1. `current_password` wajib benar (dicek via bcrypt) → salah = `401 current password is incorrect`.
2. `new_password` min 6 karakter dan harus beda → selain itu `400`.
3. Sukses (`200 {"message": "password updated, please login again"}`) sekaligus me-revoke SEMUA refresh token user → logout di semua perangkat/sesi, login ulang pakai password baru. Access JWT yang sedang berjalan tetap valid sampai kedaluwarsa (maks 1 jam).
4. UI dashboard: halaman `/profile` → card "Change Password" (password lama + baru + konfirmasi, validasi min 6 & cocok). Sukses otomatis logout → redirect `/signin`.

Contoh curl (ganti token admin):

```bash
curl -X PATCH https://daebak.artetis.my.id/api/users/me/password \
 -H 'Content-Type: application/json' \
 -H 'Authorization: Bearer <access_token>' \
 -d '{"current_password":"secret123","new_password":"secret-baru-123"}'
```

## Middleware Autentikasi (Backend)

File: `internal/middleware/auth.go`

1. Ambil JWT dari cookie `token`; jika kosong, coba header `Authorization` (kompatibilitas).
2. Jika memakai header, pastikan format `Bearer <token>`; jika keduanya kosong, kembalikan `401`.
3. Parse & verifikasi JWT dengan `utils.ParseToken`.
4. Simpan `userID`, `username`, `role` ke Gin context (`c.Set`).
5. Lanjut ke handler (`c.Next()`).
6. Jika gagal, kembalikan `401 Unauthorized` dan hentikan request.

## Penyimpanan Token

- Access JWT pendek (1h) + refresh token opaque terpisah disimpan di cookie httpOnly `token` dan `refresh_token` (diset backend via `Set-Cookie`, flag `HttpOnly`, `SameSite=Lax`, path refresh dibatasi `/api/auth`).
- Umur refresh by role: admin 24h, customer 90 hari. Rotasi di `POST /api/auth/refresh` (token lama di-revoke).
- `fetch` selalu memakai `credentials: "include"` agar cookie terkirim; mobile menyimpan pasangan token di Keychain/Keystore dan kirim `Authorization: Bearer`.
- Saat startup, `AuthProvider` memanggil `/api/auth/me` dan menolak sesi non-admin.
- Penjagaan route di `src/proxy.ts` (cek keberadaan cookie + redirect `/signup` → `/signin` karena register publik mati): tanpa cookie, semua halaman kecuali `/`, `/signin`, `/error-404` dialihkan ke `/signin`; user ber-cookie yang buka `/signin` dialihkan ke `/dashboard`.

## Catatan Keamanan

- Password tidak pernah dikembalikan ke client (field `Password` bertag `json:"-"`).
- Gunakan `JWT_SECRET` yang kuat di produksi, aktifkan `COOKIE_SECURE=true` (wajib HTTPS), dan pertimbangkan proteksi CSRF tambahan untuk pengerasan lanjutan.