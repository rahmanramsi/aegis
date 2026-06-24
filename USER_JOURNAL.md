# Aegis — User Journal (North Star)

## Current State (24 Jun 2026)

**Sudah bisa:** register → login → workspace → daemon → agent → Telegram bot langsung hidup. Semua dari web UI. Tidak perlu edit config atau restart.

---

## Flow yang Kuinginkan

### 1. Buka Dashboard → Langsung Bisa Pakai
```
Aku buka http://localhost:8080
→ Login page (email + password)
→ Belum punya akun? Klik "Create account" → register → dapat API key
→ Auto-login → dashboard kosong: "Create your first workspace"
```
**Status:** ✅ Done

### 2. Buat Workspace → Connect Daemon
```
Aku klik "New Workspace" → kasih nama "Project Alpha"
→ Workspace muncul di dashboard
→ Aku klik "Add Daemon" → kasih nama "macbook-pro"
→ Muncul token enrollment → aku copy
→ Di terminal: AEGIS_DAEMON_TOKEN=xxx ./aegis-agent
→ Daemon connect → harnesses muncul: [echo, claude, codex, opencode]
→ Status: online (hijau)
```
**Status:** ✅ Done (tapi daemon list belum show harnesses — known bug)

### 3. Buat Agent → Pilih Personality → Bot Langsung Jalan
```
Aku klik "New Agent" → isi form:
  - Name: "Support Engineer"
  - Daemon: pilih "macbook-pro" (dropdown)
  - Harness: pilih "claude" (dropdown dari daemon)
  - Model: "claude-sonnet-4-20250514" (opsional)
  - Telegram token: paste dari @BotFather
  - Personality: "Kamu technical support..." (textarea)

→ Agent created
→ Bot Telegram langsung jalan — TIDAK PERLU RESTART
→ Agent card: "Support Engineer · claude · Telegram connected"
```
**Status:** ✅ Done (kecuali personality prompt field — belum ada)

### 4. Connect Chat ke Agent
```
Aku invite bot Telegram-ku ke grup
→ User kirim message di grup
→ Connection auto-terbentuk (platform: telegram, chat_id: -100xxx)
→ Aku lihat di dashboard: agent → connections → list
→ Sekarang: semua message grup itu → agent "Support Engineer" → response balik
```
**Status:** ✅ Done (auto-connection di route Handler)

### 5. Chat Experience
```
User di Telegram: "@bot kenapa deploy gagal?"
→ Agent "Support Engineer" menerima
→ Dispatch ke daemon → spawn CLI harness
→ Response balik ke Telegram
→ Conversation tersimpan di sessions/messages
```
**Status:** ✅ Done (via echo harness, real harnesses needing testing)

### 6. Multi-Workspace Isolation
```
Aku buat workspace kedua: "Client XYZ"
→ Daftarin daemon (atau pakai yg sama)
→ Bikin agent baru: "Client Support" · harness: opencode
→ Connect ke Telegram grup client
→ Dua workspace terisolasi penuh
→ Ganti workspace dari sidebar
```
**Status:** ✅ Done

### 7. Monitoring & Control
```
Dashboard: lihat semua workspace
Workspace page: lihat agent, daemon status, recent sessions
Agent page: connections, message history, toggle enabled/disabled
Connection page: sessions, chat history
Daemon page: harnesses, status, uptime
```
**Status:** 🔧 Partial (message history view belum di UI)

---

## Fitur Prioritas — Segera

| # | Fitur | Prioritas |
|---|---|---|
| 1 | **Personality prompt per agent** — system prompt yg bikin tiap agent punya "jiwa" beda | 🔴 P0 |
| 2 | **Fix daemon list harnesses** — dropdown agent create kosong karena List tidak return harnesses | 🔴 P0 |
| 3 | **Send-on-closed-channel panic fix** — semua harness punya race condition | 🔴 P0 |
| 4 | **Dashboard compile fix** — missing Input/Label imports | 🔴 P0 |
| 5 | **Task cancel on disconnect** — subprocess orphan 30 menit | 🟠 P1 |
| 6 | **Daemon harnesses in List** — return harnesses array di GET list | 🟠 P1 |
| 7 | **Output daemon persist to messages** — history chat lengkap (user + agent) | 🟠 P1 |
| 8 | **Message history viewable di UI** — sessions + messages page | 🟡 P2 |
| 9 | **Personality prompt field di agent form** — UI untuk set system prompt | 🟡 P2 |

---

## Fitur V2

| # | Fitur |
|---|---|
| 10 | Slack adapter |
| 11 | Discord adapter |
| 12 | Streaming response (chunk per baris) |
| 13 | Token usage tracking per agent |
| 14 | RBAC (admin/member) |
| 15 | Attachment support (gambar, file) |

---

## Yang Tidak Kuinginkan

- ❌ Edit config file / ENV untuk setup
- ❌ Restart service tiap tambah agent/bot
- ❌ Vendor lock-in ke satu AI provider
- ❌ Satu bot untuk semua chat
- ❌ Token bot plaintext di DB

---

## Verdict

Aegis v0.8: core loop (register → agent → Telegram bot) works end-to-end. Butuh P0 fixes + personality prompt untuk jadi usable. Target: user bisa bikin multiple agents dengan personality berbeda, semua dari web UI.
