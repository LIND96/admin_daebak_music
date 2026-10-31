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

### Frontend

1. Pengguna membuka halaman `/signup` dan mengisi form (first/last name, email, password).
2. Saat submit, form `SignUpFormFields` memanggil `useAuth().register(...)` (username diisi dari bagian lokal email, full name dari first+last name).
3. `AuthProvider.register` memanggil `api.register(...)` (POST `/api/auth/register`).
4. Jika sukses, cookie httpOnly `token` diset oleh backend dan state `user` di-set, lalu redirect ke `/dashboard`.

### Backend

1. Handler `AuthHandler.Register` mem-bind JSON ke `registerRequest` (validasi binding).
2. Password di-hash dengan bcrypt (`utils.HashPassword`).
3. Buat record `User` di database (GORM `Create`).
4. Jika email/username duplikat, kembalikan `409 Conflict`.
5. Generate JWT (`utils.GenerateToken`) berisi `user_id`, `username`, `role`.
6. Set cookie httpOnly `token` (`setAuthCookie`, `SameSite=Lax`) dan response `201 Created` dengan `{ token, user }`.

### Diagram

```
User -> Form Register -> api.register -> POST /api/auth/register
  -> validate -> hash password -> INSERT users -> JWT -> Set-Cookie + {token,user}
  -> redirect /dashboard
```

## Flow Login

### Frontend

1. Pengguna membuka `/signin`, mengisi email dan password.
2. Form sign-in TailAdmin memanggil `useAuth().login(email, password)`.
3. `AuthProvider.login` memanggil `api.login(...)` (POST `/api/auth/login`).
4. Jika sukses, cookie httpOnly diset dan redirect ke `/dashboard`.
5. Jika gagal, tampilkan error inline pada form.

### Backend

1. Handler `AuthHandler.Login` mem-bind JSON ke `loginRequest`.
2. Cari user berdasarkan email (`Where("email = ?")`).
3. Jika tidak ditemukan, kembalikan `401 Unauthorized`.
4. Bandingkan password dengan bcrypt (`utils.CheckPassword`).
5. Jika salah, kembalikan `401 Unauthorized`.
6. Generate JWT, set cookie httpOnly, dan response `{ token, user }`.

### Diagram

```
User -> Form Login -> api.login -> POST /api/auth/login
  -> cari user by email -> verifikasi password -> JWT -> Set-Cookie + {token,user}
  -> redirect /dashboard
```

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

1. Halaman `/dashboard` dan `/users` dilindungi `src/proxy.ts`; jika belum login, redirect ke `/signin` sebelum halaman render.
2. Setelah login terkonfirmasi, halaman `/users` memanggil `api.listUsers()` (GET `/api/users`).
3. Render tabel users TailAdmin (avatar, nama, email, role, tanggal join).

### Backend

1. Middleware `Auth` memvalidasi JWT dari cookie `token` (fallback: header `Authorization`).
2. Handler `UsersHandler.List` mengambil semua user dari DB.
3. Response `{ users: [...], total }` (password tidak disertakan).

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

## Middleware Autentikasi (Backend)

File: `internal/middleware/auth.go`

1. Ambil JWT dari cookie `token`; jika kosong, coba header `Authorization` (kompatibilitas).
2. Jika memakai header, pastikan format `Bearer <token>`; jika keduanya kosong, kembalikan `401`.
3. Parse & verifikasi JWT dengan `utils.ParseToken`.
4. Simpan `userID`, `username`, `role` ke Gin context (`c.Set`).
5. Lanjut ke handler (`c.Next()`).
6. Jika gagal, kembalikan `401 Unauthorized` dan hentikan request.

## Penyimpanan Token (Frontend)

- Token disimpan di cookie httpOnly bernama `token` (diset oleh backend via `Set-Cookie`, flag `HttpOnly`, `SameSite=Lax`).
- `fetch` selalu memakai `credentials: "include"` agar cookie terkirim.
- Saat startup, `AuthProvider` memanggil `/api/auth/me` untuk memulihkan sesi.
- Penjagaan route di `src/proxy.ts` (cek keberadaan cookie): tanpa cookie, semua halaman kecuali `/`, `/signin`, `/signup`, `/error-404` dialihkan ke `/signin`; user ber-cookie yang buka `/signin`/`/signup` dialihkan ke `/dashboard`.

## Catatan Keamanan

- Password tidak pernah dikembalikan ke client (field `Password` bertag `json:"-"`).
- Gunakan `JWT_SECRET` yang kuat di produksi, aktifkan `COOKIE_SECURE=true` (wajib HTTPS), dan pertimbangkan proteksi CSRF tambahan untuk pengerasan lanjutan.