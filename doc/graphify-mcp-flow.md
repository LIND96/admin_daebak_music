# Flow Menggunakan Graphify + MCP di `admin_daebak_music`

Dokumen ini menjelaskan cara memakai Graphify (knowledge graph) + MCP server agar AI paham keseluruhan project tanpa baca semua file.

Prinsip: **query graf dulu, baru baca 2-3 file yang disebut graf.**

## 1. Setup Sekali

```bash
uv tool install "graphifyy[mcp]"  # CLI + MCP server (package PyPI: graphifyy, double-y)
graphify install                   # daftarkan skill ke assistant (Claude/Cursor/Codex/dll)
```

Lalu di dalam AI assistant, dari root repo:

```
/graphify .
```

Hasil di `graphify-out/`:

- `graph.json` — graf penuh, dibaca MCP server
- `GRAPH_REPORT.md` — ringkasan arsitektur
- `graph.html` — visual klik di browser

Aktifkan query-first (agar AI tidak langsung grep):

```bash
graphify claude install   # sesuaikan: codex / cursor / gemini
graphify hook install     # auto-rebuild graf tiap commit/checkout
```

## 2. Pilih 1 Jalur Akses (Jangan Dobel Wajib)

| Jalur | Kapan | Config |
|-------|-------|--------|
| Skill CLI lokal | Solo, sudah ada CLI | jalankan `graphify query "..."`, `graphify path A B`, `graphify explain "..."` |
| MCP STDIO | Solo + client support MCP | `graphify-mcp /absolut/path/admin_daebak_music/graphify-out/graph.json` |
| MCP HTTP | Tim / 1 graf dipakai bersama | serve via Docker, client arahkan ke `http://host:8099/mcp` |

Contoh config STDIO (Claude Code / Cursor / JSON client):

```bash
claude mcp add --transport stdio graphify -- graphify-mcp /absolut/path/admin_daebak_music/graphify-out/graph.json
```

```json
{
  "mcpServers": {
    "graphify": {
      "command": "graphify-mcp",
      "args": ["/absolut/path/admin_daebak_music/graphify-out/graph.json"]
    }
  }
}
```

Contoh serve tim via Docker (image dari `Dockerfile` root graphify):

```bash
export GRAPHIFY_API_KEY="$(openssl rand -hex 32)"
docker run -d --name graphify-mcp -p 127.0.0.1:8099:8080 \
  -v "$PWD/graphify-out:/data:ro" graphify \
  /data/graph.json --transport http --host 0.0.0.0 --api-key "$GRAPHIFY_API_KEY"
```

Client kirim key via `Authorization: Bearer <key>` atau `X-API-Key`. Wajib pakai `--api-key` jika bind `0.0.0.0`.

## 3. Flow Harian

```
1. Update graf jika habis pull / ubah banyak file:
   graphify update .   (atau: /graphify . --update)

2. Tanya ke graf dulu:
   query_graph "bagaimana login Frontend Next.js ke backend Go?"
   shortest_path "AuthHandler" "User"
   get_node "proxy.ts"

3. Baru baca file yang disebut graf (cukup 2-3 file).

4. Cek confidence edge:
   EXTRACTED = eksplisit di source, percaya.
   INFERRED  = hasil resolusi, cek source manual.
```

Contoh di repo ini:

- Tanya: `query_graph "alur login sampai cookie?"`
- Harusnya kembali: `lib/auth-context.tsx -> lib/api.ts -> handlers/auth.go -> middleware/auth.go -> proxy.ts`
- Baru buka file tersebut, tidak perlu grep seluruh `frontend/src` + `backend/internal`.

Perintah CLI setara (tanpa MCP):

```bash
/graphify query "what connects auth to the database?"
/graphify path "AuthHandler" "User"
/graphify explain "proxy.ts"
```

## 4. Tools MCP yang Tersedia

| Tool | Fungsi |
|------|--------|
| `query_graph` | subgraph hemat token (`mode: bfs/dfs`, `depth: 1-6`, `token_budget` default 2000) |
| `get_node` | detail 1 node by label/ID |
| `get_neighbors` | semua tetangga langsung + relasi |
| `get_community` | semua node dalam 1 community |
| `god_nodes` | node paling terkoneksi |
| `graph_stats` | jumlah node/edge/community + confidence |
| `shortest_path` | jalur antar 2 node (`max_hops` default 8) |
| `list_prs`, `get_pr_impact`, `triage_prs` | dampak PR terhadap graf |

Resources: `graphify://report`, `graphify://stats`, `graphify://god-nodes`, `graphify://surprises`, `graphify://audit`, `graphify://questions`.

## 5. Batasan (Tidak Otomatis 100%)

- AI harus memanggil tools, graf tidak auto-masuk konteks.
- Graf basi jika lupa `update` setelah ubah kode.
- Jawaban dipotong `token_budget/depth`, repo besar perlu query lanjutan.
- Nama ambigu ditolak dengan daftar kandidat, pilih manual.
- Docs/PDF butuh semantic pass + API key (`GEMINI/ANTHROPIC/OPENAI`), code-only bisa offline via `--code-only`.

## 6. Troubleshooting

- `command not found: graphify-mcp` → install ulang `uv tool install "graphifyy[mcp]"`.
- Tools MCP kosong → cek koneksi MCP vs instalasi skill, keduanya beda jalur.
- Jawaban lawas → `graphify update .`, pastikan server reload `graph.json`.
- Butuh full konteks → naikkan `depth: 5`, `token_budget: 4000`, atau mode `dfs`.
