# Rangkuman Deployment — OCI + Cloudflare + Docker + Bootstrap

End-to-end dari akun OCI kosong sampai `https://daebak.artetis.my.id` hidup.
Repo: `LIND96/admin_daebak_music`, branch **`daily-record`** (branch `main` masih kode lama!).
Stack: Caddy + Next.js + Go + Postgres + Dozle via `docker-compose.yml`.

## 0. Prinsip env (baca dulu!)

- Semua setting beda-per-environment dibaca dari **environment shell** via
  interpolasi `${VAR:-default}` di `docker-compose.yml` (backend, frontend
  `build.args` + `environment`, Caddy `DOMAIN`, Dozle kredensial).
- Artinya: cukup `export` sekali (atau via file `~/deploy.env` di §7), **tanpa
  edit file compose**. Nilai default = mode lokal (`http://localhost`).
- Pengecualian: `NEXT_PUBLIC_API_URL` di-bake ke JS **saat `docker build`** —
  export WAJIB dilakukan **sebelum** `docker compose up --build` (lihat §9).

## 1. Jaringan (VCN) — Networking → Virtual Cloud Networks

- Jangan isi form manual. Klik **Actions → Start VCN Wizard** → *VCN with Internet Connectivity*.
- Isi: nama `vcn-daebak`, CIDR `10.0.0.0/16`, public subnet `10.0.0.0/24`, private `10.0.1.0/24`, DNS ON. Next → Create.
- Wizard otomatis membuat: VCN + public/private subnet + Internet Gateway + Route Table + Security List.
- Pelajaran: membuat VCN + subnet *bersamaan* dengan instance (opsi *Create new*)
  mengunci toggle Public IPv4 (OFF, tidak bisa diklik). Selalu buat jaringan dulu via wizard.

## 2. Instance — Compute → Instances → Create instance

| Bagian | Isi |
|--------|-----|
| Placement | AD 1 |
| Image | **Change image → Canonical Ubuntu 22.04 Minimal** (bawaan Oracle Linux 9 diganti agar sesuai panduan Docker Ubuntu) |
| Shape | Ampere A1 Flex atau AMD E5 Flex, 1 OCPU / 6–12 GB (cukup untuk stack ini) |
| Security | Shielded instance OFF, Security attributes kosong |
| Networking | *Select existing* → VCN `vcn-daebak` + public subnet-nya. **Public IPv4 = ON** (baru bisa diklik kalau subnet sudah ada). IPv6 OFF |
| SSH keys | *Paste/Upload public key* (`id_ed25519.pub`). Kalau *Generate*, wajib **Download private + public key** sebelum Create |
| Storage | Default (boot volume ±46 GB) |

- Review → Create → tunggu Running → catat **Public IP**.
- **Penting — IP bisa berubah!** IP publik ephemeral OCI berganti tiap instance
  stop/start/dibuat ulang (pernah kejadian: `.188` → `.191`, web down padahal
  container sehat). Solusi permanen: Networking → IP Management →
  **Reserved Public IPs** → Reserve + attach ke VNIC instance. Arahkan DNS ke IP
  reserved sekali, selesai.
- Key SSH di Mac biasanya di `~/Downloads`, bukan home — cek `ls ~/Downloads | grep -i key`,
  lalu `chmod 600 <key>` dan `ssh -i <key> ubuntu@<PUBLIC_IP>`.

## 3. Security List — buka port 80 / 443 / 22

- Networking → VCN `vcn-daebak` → Security → **Default Security List** → Add Ingress Rules (3x, semua Stateful OFF, Source `0.0.0.0/0`, TCP):
  - Destination `80` (HTTP — wajib untuk validasi Let's Encrypt)
  - Destination `443` (HTTPS — aplikasi)
  - Destination `22` (SSH — admin; idealnya batasi ke IP sendiri)
- Port `5432` (Postgres) dan `9999` (Dozle) **jangan dibuka** — akses keduanya via SSH tunnel (lihat §10).
- Fungsi: tanpa ini request mati di gerbang OCI sebelum sampai Caddy. `0.0.0.0/0`
  artinya dari IP mana pun — perlu karena user, HP, dan server Let's Encrypt datang dari IP acak.

## 4. DNS — Cloudflare → domain `artetis.my.id` → DNS → Records

- Jangan ubah record lama (root dipakai R2, `api-mobile` ke IP lain).
- **Add record**: Type `A`, Name `daebak`, IPv4 = Public IP VM (**yang reserved**,
  lihat §2), **Proxy OFF (DNS only, abu-abu)**, TTL Auto.
- Verifikasi: `nslookup daebak.artetis.my.id` harus jawab IP VM.
- Setelah HTTPS hijau boleh nyalakan proxy ON + SSL mode **Full (strict)**.
  Jangan pakai Flexible (rusak cookie httpOnly/secure).

## 5. Docker di VM (via SSH)

```bash
sudo apt-get update && sudo apt-get install -y ca-certificates curl git
sudo install -m 0755 -d /etc/apt/keyrings
sudo curl -fsSL https://download.docker.com/linux/ubuntu/gpg -o /etc/apt/keyrings/docker.asc
echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.asc] https://download.docker.com/linux/ubuntu $(. /etc/os-release && echo $VERSION_CODENAME) stable" | sudo tee /etc/apt/sources.list.d/docker.list
sudo apt-get update && sudo apt-get install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
sudo usermod -aG docker $USER && newgrp docker
```

Ubuntu Minimal tidak ada `nano` — edit file pakai `vi` atau `sed`, bukan `nano`.

## 6. Kode — repo privat via token + branch yang benar

- GitHub → Settings → Developer settings → Personal access tokens → **Fine-grained**:
  resource owner akunmu, *Only select repositories* → repo ini,
  Permissions → **Contents: Read-only** (+ Metadata otomatis). Generate → copy sekali.
- Di server:
  ```bash
  git clone https://<TOKEN>@github.com/LIND96/admin_daebak_music.git admin_daebak_music
  cd admin_daebak_music
  git checkout daily-record   # WAJIB — main masih kode lama tanpa tracking/Caddy!
  ```
  (jangan tempel token mentah ke terminal — itu bukan perintah; token hanya di dalam URL clone).
- Update berikutnya: `git stash` (kalau ada edit lokal) → `git pull` (pastikan tetap di `daily-record`).

## 7. File env persisten (sekali saja, anti lupa)

Env shell hilang tiap SSH baru — itu pernah bikin Caddy balik ke `localhost`.
Simpan sekali, source tiap deploy:

```bash
cat > ~/deploy.env <<'EOF'
export DOMAIN=daebak.artetis.my.id
export JWT_SECRET='ganti-acak-min-32-karakter'
export COOKIE_SECURE=true
export FRONTEND_URL=https://daebak.artetis.my.id
export NEXT_PUBLIC_API_URL=https://daebak.artetis.my.id
export BOOTSTRAP_SECRET=''
export GOOGLE_ANDROID_CLIENT_ID=''
export GOOGLE_IOS_CLIENT_ID=''
export DOZZLE_USERNAME=admin
export DOZZLE_PASSWORD='ganti-password-dozle'
EOF
```

Pola deploy standar (selalu dari `~/admin_daebak_music`):

```bash
set -a; source ~/deploy.env; set +a
docker compose up --build -d
```

Tabel env lengkap (semua sudah `${VAR:-default}` di compose):

| Variabel | Default | Wajib ganti produksi |
|----------|---------|----------------------|
| `DOMAIN` | `localhost` | Ya — domain Caddy + HTTPS otomatis |
| `JWT_SECRET` | `change-me-in-production` | Ya — acak min 32 char |
| `COOKIE_SECURE` | `false` | Ya — `true` (HTTPS) |
| `FRONTEND_URL` | `http://localhost` | Ya — `https://domain` (CORS backend) |
| `NEXT_PUBLIC_API_URL` | `http://localhost` | Ya — `https://domain`, **sebelum build** (§9) |
| `BOOTSTRAP_SECRET` | kosong (= endpoint mati) | Isi sementara saat buat admin pertama (§8) |
| `GOOGLE_*_CLIENT_ID` | kosong | Ya jika login Google HP dipakai |
| `DOZZLE_USERNAME/PASSWORD` | `admin` / `ganti-password-dozle` | Ya — passwordnya |
| `JWT_EXPIRY`, `REFRESH_EXPIRY_*` | `24h`, `24h`, `2160h` | Opsional |

## 8. Migrasi DB

```bash
docker compose exec backend /migrate up
docker compose exec backend /migrate version   # pastikan 000007 (device_brand)
```

Compose meng-override command backend menjadi `/server` saja, jadi migrasi
**tidak otomatis** — selalu jalankan manual setelah `up` (mencakup tabel
`daily_app_usages` + kolom `country_code`, `device_brand/model`).

## 9. BOOTSTRAP_SECRET — admin pertama

- `setup_key` bukan diambil dari mana-mana — **kamu yang mengarang**, taruh di
  `~/deploy.env` (`BOOTSTRAP_SECRET`), source, lalu `docker compose up -d backend`.
- Verifikasi: `docker compose exec backend printenv BOOTSTRAP_SECRET`.
- Request (Postman → Auth → Bootstrap, atau curl ke
  `POST https://daebak.artetis.my.id/api/auth/bootstrap-admin`):
  `{email, username, password min 6, full_name?, setup_key}` → `201` sukses.
  `403 bootstrap disabled` = secret belum masuk container / tidak sama.
- Setelah admin jadi: kosongkan `BOOTSTRAP_SECRET` di `deploy.env` + `up -d backend`
  agar endpoint sekali-pakai itu mati permanen.

## 10. NEXT_PUBLIC_API_URL — jebakan build-time (penting!)

- Gejala: produksi menembak `http://localhost:8080` walau env sudah benar + rebuild.
- Penyebab: variabel `NEXT_PUBLIC_*` di-bake ke JS **saat `docker build`**.
- Perbaikan repo (sudah diterapkan): `frontend/Dockerfile` memakai `ARG`, compose
  meneruskannya via `build.args`. Syaratnya: `export` (atau source `deploy.env`)
  dilakukan **sebelum** `docker compose up --build`.
- Verifikasi di browser: DevTools → Network → request `login` harus ke
  `https://<domain>/api/...`, bukan `localhost:8080`. Kalau masih lama:
  hard-refresh (`Ctrl+Shift+R`), bundle lama ke-cache.

## 11. Rate limit login (bawaan, tanpa config)

- `POST /api/auth/login` dan `/bootstrap-admin` dibatasi per IP: burst 10x,
  lalu ~10x/menit → lebihnya `429 + Retry-After: 60` (pesan tampil inline di form).
- Syarat: `SetTrustedProxies` private-range sudah di kode agar IP asli user
  terbaca dari `X-Forwarded-For` Caddy. Tidak perlu setting tambahan.

## 12. Dozle — log via browser

- Service `dozle` ikut menyala (`docker compose ps`). Bind `127.0.0.1:9999` (tidak publik).
- Akses: di Mac/laptop, buka tunnel `ssh -i <key> -N -L 9999:localhost:9999 ubuntu@<IP>`,
  lalu `http://localhost:9999` → login `DOZZLE_USERNAME/PASSWORD`.
- Alternatif tanpa tunnel: `docker compose logs -f backend|caddy|frontend`.

## 13. Operasional & debug

```bash
docker compose ps / logs -f / logs -f caddy
docker compose up --build -d      # setelah git pull (backend+frontend berubah)
docker compose exec backend /migrate up
docker compose exec db pg_dump -U postgres admin_daebak_music > backup.sql
# DB dari laptop: DBeaver + SSH tunnel (host localhost:5432, user postgres) — tanpa buka port 5432.
```

### 13a. SOP deploy per jenis perubahan

Selalu awali dari `~/admin_daebak_music` dengan load env (wajib — tanpa ini
Caddy balik ke `localhost`, lihat §7):

```bash
cd ~/admin_daebak_music
set -a; source ~/deploy.env; set +a
```

| Perubahan | Perintah |
|-----------|----------|
| Frontend saja (`frontend/`) | `docker compose up --build -d frontend` — wajib `--build` karena `NEXT_PUBLIC_*` di-bake saat build (§10). Cek `docker compose logs frontend --tail=10` (`Ready`). |
| Backend saja (`backend/*.go`, query) | `docker compose up --build -d backend` lalu `docker compose exec backend /migrate up` (wajib kalau ada file baru di `backend/migrations/`). Cek `docker compose logs backend --tail=20`. |
| Env saja (`~/deploy.env`) | Env backend (`JWT_*`, `GOOGLE_*`, `FRONTEND_URL`, `COOKIE_SECURE`, `BOOTSTRAP_SECRET`): `docker compose up -d backend` (restart saja, tanpa `--build`). Env build-time (`NEXT_PUBLIC_API_URL`) atau `DOMAIN`: `docker compose up --build -d frontend caddy`. |
| Semua / tidak yakin | `docker compose up --build -d` lalu `docker compose exec backend /migrate up`, tutup dengan `docker compose ps` (semua harus `Up`). |

Checklist cepat setelah deploy:

```bash
docker compose ps
curl -s -o /dev/null -w "backend: %{http_code}\n" http://localhost:8080/health
curl -sk -o /dev/null -w "homepage: %{http_code}\n" https://daebak.artetis.my.id/
curl -sk -o /dev/null -w "api: %{http_code}\n" "https://daebak.artetis.my.id/api/tracking/filter-options?date=2026-11-05"
# api = 401 itu NORMAL (butuh login admin). Yang tidak normal: 502/503.
```

Catatan: `/health` hanya ada di backend langsung — via Caddy ia jatuh ke
frontend dan menjawab `307` (redirect), itu normal. Jangan pakai
`https://<domain>/health` sebagai health check publik.

Yang jangan dilakukan:

- `docker compose down -v` — menghapus volume `pgdata` = database hilang.
- `up` tanpa `source ~/deploy.env` dulu.
- Lupa `migrate up` setelah pull yang membawa `backend/migrations/*` baru.

Urutan debug "web down": DNS (`nslookup` = IP reserved?) → container (`ps`,
ada yang `Restarting`?) → `logs caddy` (domain yang dilayani harus domainmu,
bukan `localhost`; kalau `localhost` berarti `DOMAIN` tidak ke-export saat `up`) →
`logs backend` → Network tab (`localhost:8080` = build tanpa export, lihat §10) →
role admin (customer = `403`, pakai akun bootstrap) → `429` = kena rate limit, tunggu 1 menit.

Setelah semua hijau: Cloudflare proxy boleh ON + SSL **Full (strict)**.
