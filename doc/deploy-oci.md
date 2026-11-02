# Deploy ke OCI Compute (Indonesia North / Batam)

Target: 1 VM Ubuntu + Docker Compose (`db`, `backend`, `frontend`, `caddy`).
Caddy otomatis HTTPS jika `DOMAIN` menunjuk ke IP publik VM.

## 1. Siapkan Jaringan (sekali saja)

1. OCI Console → Networking → Virtual Cloud Networks → Create VCN (default, dengan Internet Gateway + Route Table + Security List).
2. Security List → tambahkan Ingress:
   - `0.0.0.0/0` TCP `22` (SSH, batasi ke IP kamu jika bisa)
   - `0.0.0.0/0` TCP `80` (HTTP, untuk validasi Let's Encrypt)
   - `0.0.0.0/0` TCP `443` (HTTPS)
3. Catat DNS: arahkan `A record` domainmu → Public IP VM (lihat langkah 3).
   Tunggu propagasi sebelum `docker compose up` agar Caddy bisa terbit sertifikat.

## 2. Buat VM (Compute → Instances → Create instance)

- Region: Indonesia North (Batam), Compartment sesuai screenshot kamu.
- Image: Ubuntu 24.04 (Minimal juga OK).
- Shape: Ampere A1 (gratis / murah, arm64) atau E2 (amd64). 2 OCPU / 4 GB cukup untuk stack ini.
  Catatan: image Docker `postgres:16-alpine`, `caddy:2-alpine`, `node:20-alpine`,
  `golang:1.26-alpine` semua multi-arch → aman di Ampere.
- VCN/Subnet: yang dibuat di langkah 1 (public subnet).
- SSH: upload public key kamu (`~/.ssh/id_ed25519.pub`).
- Boot volume: default 50 GB cukup.

## 3. Siapkan VM via SSH

```bash
ssh -i ~/.ssh/id_ed25519 ubuntu@<PUBLIC_IP>

# Docker (Ubuntu)
sudo apt-get update && sudo apt-get install -y ca-certificates curl git
sudo install -m 0755 -d /etc/apt/keyrings
sudo curl -fsSL https://download.docker.com/linux/ubuntu/gpg -o /etc/apt/keyrings/docker.asc
echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.asc] https://download.docker.com/linux/ubuntu $(. /etc/os-release && echo $VERSION_CODENAME) stable" | sudo tee /etc/apt/sources.list.d/docker.list
sudo apt-get update && sudo apt-get install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
sudo usermod -aG docker $USER && newgrp docker

# Firewall Ubuntu (jika ufw aktif)
sudo ufw allow 22,80,443/tcp
```

## 4. Deploy Aplikasi

```bash
git clone <repo-mu> admin_daebak_music && cd admin_daebak_music

# Env produksi — JANGAN pakai default!
export DOMAIN=app.example.com
export JWT_SECRET='isi-acak-panjang-min-32-karakter'
export COOKIE_SECURE=true   # wajib HTTPS; didukung via compose? lihat catatan

docker compose up --build -d
docker compose exec backend /migrate up
docker compose ps
docker compose logs -f caddy backend
```

Catatan env saat ini:
- `JWT_SECRET`, `FRONTEND_URL`, `COOKIE_SECURE` di-hardcode di `docker-compose.yml`.
  Untuk produksi, override via shell sebelum `up`:
  ```bash
  export JWT_SECRET=... 
  ```
  lebih baik pindahkan ke interpolasi `${JWT_SECRET:?...}` — minta saya jika mau saya rapikan.
- `NEXT_PUBLIC_API_URL=http://localhost` di compose adalah build-time.
  Untuk domain produksi, ubah sementara menjadi `https://app.example.com`
  lalu `docker compose up --build` ulang agar JS memanggil same-origin HTTPS.
- Buka `https://app.example.com` → harusnya gembok hijau (Caddy menerbitkan sertifikat otomatis).

## 5. Operasional

```bash
docker compose logs -f            # semua service
docker compose logs -f caddy      # masalah sertifikat / routing
docker compose exec backend /migrate up   # setelah update migrasi
docker compose up --build -d      # setelah git pull
docker compose down               # stop (data aman di volume pgdata)
docker compose down -v            # HATI-HATI: hapus database
```

Backup DB:

```bash
docker compose exec db pg_dump -U postgres admin_daebak_music > backup.sql
```

## Troubleshooting

| Gejala | Penyebab / Solusi |
|--------|-------------------|
| HTTPS tidak terbit | DNS belum propagasi / port 80 tertutup / `DOMAIN` salah. Cek `logs caddy`. |
| 502 dari Caddy | backend/frontend belum sehat. Cek `docker compose ps` + `logs backend frontend`. |
| Cookie login hilang | Masih HTTP tapi `COOKIE_SECURE=true`, atau `FRONTEND_URL` tidak sama dengan origin. Samakan dengan domain. |
| OOM di Ampere kecil | Kurangi `NODE_OPTIONS` build atau build frontend di lokal lalu push image. |
| Port 80 bentrok | Pastikan hanya Caddy yang bind 80/443 di VM. |
