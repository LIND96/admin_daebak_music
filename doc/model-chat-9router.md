# Rangkuman Model & Provider untuk Fitur Chat (via 9Router)

Hasil kompilasi pengecekan langsung ke 9Router (`GET /v1/models` + `POST /v1/chat/completions`
per model, body identik dengan yang dikirim backend: `temperature 0.9`, `max_tokens 10`).
Status dapat berubah (kuota, model retired, auth) — backend dirancang tahan terhadap itu
(auto-discovery + failover + cooldown, lihat § Cara kerja).

## Cara kerja singkat (backend)

- Daftar model diambil otomatis dari 9Router (`GET /v1/models`, cache 5 menit).
- Tiap request diundi round-robin, gagal → coba model berikut (maks 3, `maxChatModelAttempts`).
- Model gagal masuk cooldown 24 jam (`failedModelCooldown`) lalu dicoba lagi otomatis.
- Respons: `{"reply","provider":"9router:<served>","model":<served>,"requested_model":<undi>}`.
  `model` = yang benar-benar menjawab (bisa beda karena failover di 9Router).
- Dasbor "Provider terpopuler" grouping generik atas kolom `provider` — jalan otomatis.

## Yang SEHAT (boleh di-enable) — 21 model

| Provider | Model | Ket. |
|---|---|---|
| opencode (`oc/`) | `iceman-free`, `beam-free`, `flint-40b-free`, `nemotron-3.5-super-vl-free`, `nemotron-3.5-lightning-free`, `big-pickle` | 200 valid |
| opencode-go (`ocg/`) | `longcat-2.5-free`, `beam-free`, `iceman-contributor-free` | 200 valid |
| groq | `groq/openai/gpt-oss-120b` | 200 (reasoning) |
| openrouter | `recursive/flint-v1.0:free`, `liquid/lfm-3-1.7b:free`, `nvidia/nemotron-3-nano-omni-30b-a3b-reasoning:free`, `cohere/north-mini-code:free`, `dots-studio/dots-3-note-preview:free`, `nvidia/nemotron-3.5-lightning:free`, `google/lyria-3-pro-preview` | 200 valid |
| openrouter | `openrouter/free` (router → particle-1), `apodex/apodex-1.1-mini:free`, `baidu/ernie-5.5-flash-preview:free`, `particle-ai/particle-1:free` | 200 valid |

## Yang RUSAK (disable di UI 9Router)

| Provider | Model | Sebab |
|---|---|---|
| opencode | `oc/jev-1.13-free` | 400 ModelProtocolUnsupported |
| opencode-go | `ocg/space-bunny-free` | 400 Model is unavailable |
| mmf | `mmf/mimo-auto` | 502 bootstrap 404 |
| groq (sudah dihapus dari daftar) | `llama-3.3-70b-versatile`, `llama-4-maverick-17b-128e-instruct` | 503 tidak ada akses di sisi Groq |
| groq | `qwen/qwen3-32b` | 400 `enable_thinking` unsupported — matikan thinking di setting model bila mau dipakai lagi |
| openrouter | `typesafe/jev-1.13` | 400 |
| openrouter | `thinkingmachines/inkling-2-small:free` | 403 permanen (hanya untuk agentic harness) |
| openrouter | `nvidia/nemotron-3.5-super-vl:free`, `nvidia/nemotron-3-ultra-550b-a55b:free` | respons korup / 503 |
| openrouter | `google/gemma-4-31b-it:free`, `google/gemma-4-26b-a4b-it:free` | 429 rate-limit sementara — tes ulang sebelum enable |
| openrouter | `google/lyria-3-clip-preview` | 400 upstream |
| ollama | `glm-5`, `minimax-m2.5` | 410 retired |
| ollama | `minimax-m3` | 503/402 akun tanpa kredit |
| ollama | `clef`, `clef-flash`, `clef:cloud` | 400 decision model, hanya `/v1/systemone` |
| nvidia | `minimaxai/minimax-m2.7`, `minimaxai/minimax-m3`, `z-ai/glm-5.2`, `deepseek-ai/deepseek-v4-pro`, `deepseek-ai/deepseek-v4-flash` | 410 end of life |
| nvidia | `moonshotai/kimi-k2.6`, `nemotron-3-ultra-550b-a55b`, `parakeet-ctc-1.1b-asr` | 503/404 (parakeet = model ASR, bukan chat) |
| cline (`cl/`, 8 model di UI) | `claude-opus-4.7`, `claude-sonnet-4.6`, `claude-opus-4.6`, `gpt-5.3-codex`, `gpt-5.4`, `gemini-3.1-pro-preview`, `gemini-3.1-flash-lite-preview` | 429/402 saldo Cline $0.00 / 401 auth kedaluwarsa |
| cline | `kat-coder-pro` | tidak muncul di `/v1/models` (cek enable/disable di UI) |

Catatan sweep massal `cl/*` (461 model): 71 terkonfirmasi rusak (56 saldo habis,
12 tidak didukung, 3 respons kosong); 390 kena 429 karena pengetesan paralel —
statusnya "belum tahu", tes ulang sekuensial bila akun Cline sudah sehat.

## Aturan operasional

1. **Verifikasi-sebelum-enable.** Tiap model baru wajib lolos ini dari VPS (pakai key backend):
   ```bash
   KEY='<isi-CHAT_PROVIDER_API_KEY>'
   curl -s http://localhost:20128/v1/chat/completions -H "Authorization: Bearer $KEY" \
     -H 'Content-Type: application/json' \
     -d '{"model":"<id>","messages":[{"role":"user","content":"hi"}],"max_tokens":10,"temperature":0.9}' \
     | head -c 300; echo
   ```
   Hanya `200` + `chat.completion` berisi teks yang boleh enabled.
2. **Sumber kebenaran = daftar di VPS dengan key backend** (`GET /v1/models` dari VPS).
   Tunnel/dev lain bisa menunjuk instance berbeda — jangan jadikan patokan.
3. Model rusak permanen (403 gated, 410 retired, saldo $0) → disable, bukan diandalkan ke cooldown.
4. Setelah ubah daftar: `docker compose restart backend` (segarkan cache 5 menit) → curl chat → harus `200`.
5. API key 9Router hanya di server (`~/deploy.env`), tidak pernah di APK. Bila key pernah
   tersebar, rotasi via UI (Create Key → update env → hapus key lama).
