# Aegis — User Journal (North Star)

## Siapa Aku
Aku developer yang manage beberapa project. Aku ingin AI agent yang bisa kupasang di Telegram/Slack, bebas pilih model, dan aku kontrol penuh dari web dashboard. Tidak ada vendor lock-in. Semua self-hosted.

---

## Flow yang Kuinginkan

### 1. Buka Dashboard → Langsung Bisa Pakai

```
Aku buka browser → http://localhost:8080
→ Muncul halaman login
→ Aku register dengan email + password
→ Dapat API key → auto-login ke dashboard
→ Dashboard kosong: "Create your first workspace"
```

**Tidak boleh:** harus edit file config, set ENV, restart service, atau baca README dulu.

---

### 2. Buat Workspace → Connect Daemon

```
Aku klik "New Workspace" → kasih nama "Project Alpha"
→ Workspace muncul di dashboard
→ Aku klik "Add Daemon" → kasih nama "macbook-pro"
→ Muncul token enrollment → aku copy
→ Di terminal: AEGIS_DAEMON_TOKEN=xxx ./aegis-agent
→ Daemon connect → muncul harnesses yg tersedia: [echo, claude, codex, opencode]
→ Status daemon: online (hijau)
```

**Yang kulihat di UI:** daemon card dengan status online, daftar harness, last seen.

---

### 3. Buat Agent → Pilih Personality

```
Aku klik "New Agent" → isi form:
  - Name: "Support Engineer"
  - Daemon: pilih "macbook-pro" (dropdown)
  - Harness: pilih "claude" (dropdown dari daemon, auto-populate)
  - Model: "claude-sonnet-4-20250514" (opsional)
  - Personality prompt: "Kamu technical support yg... " (textarea)
  - Telegram bot token: aku paste dari @BotFather

→ Agent created → token disimpan hashed
→ Bot Telegram langsung jalan — TIDAK PERLU restart aegisd
→ Aku lihat agent card: "Support Engineer · claude · online · Telegram connected"
```

**Personality prompt** adalah system prompt yg dikirim ke harness bersama user message. Ini yg bikin tiap agent punya "jiwa" berbeda.

---

### 4. Connect Chat ke Agent

```
Aku invite bot Telegram-ku ke grup "Project Alpha Dev"
→ Di dashboard, aku klik agent "Support Engineer"
→ Muncul daftar Connections
→ Aku lihat: Telegram · chat -100xxx · auto-detected
→ Atau aku bisa manual add: WhatsApp group, Discord channel, dsb

Sekarang: semua message di grup itu → ke agent "Support Engineer" → response balik
```

**Auto-detect:** saat ada message pertama dari chat baru, connection auto-terbentuk. Admin approve dari dashboard.

---

### 5. Chat Experience

```
User di Telegram: "@bot kenapa deploy gagal?"
→ Agent "Support Engineer" nerima
→ Lihat personality: "Kamu technical support..."
→ Lihat model: claude-sonnet-4
→ Dispatch ke daemon → daemon spawn `claude -p "..." --model claude-sonnet-4`
→ Response streaming balik ke Telegram (per-chunk atau per-baris)
→ Semua conversation tersimpan di sessions/messages
```

**Yang kulihat di dashboard:** list sessions per connection, message history lengkap (user + agent).

---

### 6. Multi-Workspace Isolation

```
Aku buat workspace kedua: "Client XYZ"
→ Daftarin daemon baru (atau pakai yg sama)
→ Bikin agent baru: "Client Support" · harness: opencode · personality: formal & helpful
→ Connect ke Telegram grup client
→ Dua workspace terisolasi penuh — agent, chat history, daemon
→ Aku ganti workspace dari dropdown di sidebar
```

---

### 7. Monitoring & Control

```
Dashboard utama: aku lihat semua workspace
Workspace page: aku lihat semua agent, status daemon, recent sessions
Agent page: aku lihat connections, message history, bisa toggle enabled/disabled
Connection page: aku lihat sessions, bisa lihat isi chat
Daemon page: aku lihat harnesses, status, uptime

Aku bisa:
- Disable agent sementara (pause bot)
- Ganti model tanpa restart
- Edit personality prompt
- Lihat token usage per agent (future)
```

---

## Fitur Prioritas (MVP — Kerjakan Sekarang)

| # | Fitur | Status |
|---|---|---|
| 1 | Register/Login + per-user API key | ✅ Done |
| 2 | Workspace CRUD + isolation | ✅ Done |
| 3 | Daemon connect + harness discovery | ✅ Done |
| 4 | Agent CRUD + harness validation | ✅ Done |
| 5 | Per-agent Telegram token (dynamic start/stop) | 🔧 In progress |
| 6 | Personality prompt per agent | ❌ |
| 7 | Connection auto-detect from chat | ❌ |
| 8 | Message history viewable in dashboard | ❌ |
| 9 | Daemon harnesses in list endpoint | ❌ |
| 10 | Agent enable/disable toggle | ✅ Done |

---

## Fitur V2 (Nanti)

| # | Fitur |
|---|---|
| 11 | Slack adapter |
| 12 | Discord adapter |
| 13 | Streaming response (chunk per baris ke chat) |
| 14 | Token usage tracking |
| 15 | RBAC (admin/member roles) |
| 16 | Attachment support (gambar, file) |
| 17 | Webhook mode (selain long polling) |

---

## Yang Tidak Kuinginkan

- ❌ Harus SSH ke server untuk setup
- ❌ Config file YAML/TOML/JSON
- ❌ Restart service tiap tambah agent/bot
- ❌ Vendor lock-in ke satu AI provider
- ❌ Satu bot untuk semua chat (tidak bisa beda personality)
- ❌ API key bot disimpan plaintext di DB atau ENV yang gampang bocor

---

## Cara Baca Dokumen Ini

File ini adalah **spesifikasi produk**, bukan spesifikasi teknis.
Setiap kali bingung "harusnya gimana?", kembali ke file ini.
Urutan section = urutan prioritas: flow 1→7, fitur 1→10.
