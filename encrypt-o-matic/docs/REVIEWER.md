# Encrypt-O-Matic — Reviewer Documentation

Professional overview of the project for technical review. For presentation preparation and demo scripts, see [STUDENT_NOTES.md](STUDENT_NOTES.md).

---

## Review Quick Start

From the project root:

```bash
make test
make build
make server
```

Open **http://localhost:8080** and review:

| Page | Purpose |
|------|---------|
| **Dashboard** | System status, encrypted file counts, recent activity |
| **Encrypt** | Encryption workflow with local path picker |
| **Encrypted Files** | Per-file metadata, decrypt actions |
| **Security** | Layered security model (password, bcrypt, PBKDF2, algorithms) |
| **Reviewer Mode** | Requirements checklist and automated test runner |
| **Debug** | Configuration paths, password verification, metadata listing |

CLI verification (Windows target):

```bash
./dist/encrypt-o-matic.exe --help
./dist/encrypt-o-matic.exe debug-info
./dist/encrypt-o-matic.exe verify-password
./dist/encrypt-o-matic.exe tests/testdata/demo.exe AES 0 0-1000 1
./dist/encrypt-o-matic.exe decrypt tests/testdata/demo.exe
```

Related documents: [README.md](../README.md), [PROJECT_STRUCTURE.md](PROJECT_STRUCTURE.md).

---

## Project Overview

Encrypt-O-Matic is an **educational** Windows CLI file encryption tool written in Go. It encrypts **only** files or directories explicitly provided by the user. It is not malware: there is no stealth, persistence, spreading, privilege escalation, or automatic file discovery.

**Primary capabilities:**

- Password-protected encryption with bcrypt master password verification
- Three authenticated encryption algorithms: AES-256-GCM, ChaCha20-Poly1305, Twofish-GCM
- gzip compression before encryption
- Configurable random padding (file size manipulation)
- SHA-256 integrity verification after decryption
- Timer-based locking with metadata-driven auto-restore
- Directory encryption (`.exe` files only)
- Local web dashboard for review and demonstration

**Layout:** Standard Go project — `cmd/` (entry point), `internal/` (packages), `tests/` (integration), `docs/`.

---

## Features

| Feature | Description |
|---------|-------------|
| **CLI encryption** | Five positional arguments: target, algorithm, padding (MB), custom range, duration (minutes) |
| **Decrypt subcommand** | Immediate restoration with master password |
| **Master password** | bcrypt hash in `.encryptomatic/master.hash`; never stored in plaintext |
| **Metadata** | Per-file JSON: algorithm, salt, nonce, padding, hash, unlock time |
| **Backups** | Timestamped copy before every file modification |
| **Custom operation** | Configurable SHA-256 workload loop before encryption |
| **Debug commands** | `verify-password`, `debug-info` |
| **Web dashboard** | Local HTTP UI; all crypto via Go backend |
| **Automated tests** | Unit and integration tests for all three algorithms |

**CLI usage:**

```bash
./dist/encrypt-o-matic.exe <target> <algorithm> <padding_mb> <custom_range> <duration_min>
./dist/encrypt-o-matic.exe decrypt <target>
```

---

## Security Design

### Threat model

The tool is designed for **authorized lab use**. It modifies only user-specified paths. Security controls protect encrypted content and verify integrity; the tool does **not** attempt to evade antivirus or security monitoring.

### Detection surfaces

| Surface | Observable behavior |
|---------|---------------------|
| **Static** | Binary hash; encrypted files use `EOMENC01` magic header instead of `MZ` |
| **Heuristic** | High entropy in previously valid PE files; destroyed PE structure |
| **Behavioral** | Process overwrites executable, creates `.encryptomatic/` metadata and backups |
| **Artifacts** | JSON metadata, backup files, optional size increase from padding |

### Malware detection context

Understanding how security products detect threats provides context for what this tool demonstrates:

| Method | Principle | Relevance to this project |
|--------|-----------|---------------------------|
| **Signature-based** | Matches known byte patterns or hashes | Encrypted files no longer match original executable signatures |
| **Behavioral** | Correlates runtime actions (mass file modification, etc.) | Tool modifies only explicit targets; no registry or network activity |
| **Heuristic** | Rules on entropy, structure, suspicious combinations | Encryption increases entropy and removes normal PE headers |

The project implements **encryption mechanics** for education, not evasion techniques (packing, injection, anti-analysis, timestomping, etc.).

### Encryption pipeline

```
Target file → SHA-256 workload (custom range) → gzip compress → PBKDF2 key derivation →
Authenticated encryption → Random padding → Write payload + metadata
```

Implementation: `internal/fileops/fileops.go`, `internal/crypto/crypto.go`, `internal/custom/custom.go`.

---

## Encryption Algorithms

All three algorithms use **authenticated encryption** (confidentiality + integrity).

| Algorithm | Mode | Key size | Typical use |
|-----------|------|----------|-------------|
| **AES** | AES-256-GCM | 256-bit | Industry standard; TLS, BitLocker, VPNs |
| **ChaCha20** | ChaCha20-Poly1305 | 256-bit | Strong software performance without AES-NI |
| **Twofish** | Twofish-GCM | 256-bit | Educational comparison; 128-bit block cipher |

Algorithm is selected at encryption time and stored in metadata for correct decryption. Input is normalized case-insensitively (`AES`, `ChaCha20`, `Twofish`).

**Implementation:** `internal/crypto/crypto.go` — `NormalizeAlgorithm()`, `EncryptBytes()`, `DecryptBytes()`

**Verification:**

```bash
go test ./internal/crypto/...
go test ./tests/integration/...
```

---

## Password Security

| Layer | Technology | Purpose |
|-------|------------|---------|
| **Master password verification** | bcrypt (`master.hash`) | Slow password hash; resists offline brute force |
| **Encryption key derivation** | PBKDF2-HMAC-SHA256, 100,000 iterations | Derives 256-bit key from master password |
| **Per-file salt** | 16 random bytes (metadata) | Unique derived key per file |
| **Password input** | Hidden terminal prompt (`golang.org/x/term`) | No echo to screen |

Passwords exist only in memory during operations. No plaintext passwords are stored in source code or on disk.

**Implementation:** `internal/auth/auth.go`, `internal/crypto/crypto.go`

**Verification:**

```bash
./dist/encrypt-o-matic.exe verify-password
```

---

## File Integrity

Integrity is enforced at multiple layers:

1. **Authenticated encryption** — GCM/Poly1305 detects ciphertext tampering
2. **Original SHA-256** — Hash stored in metadata at encryption time
3. **Post-decryption verification** — Hash recomputed; mismatch aborts with error
4. **Backups** — Timestamped copy created before modification
5. **Permission preservation** — Original file mode stored and restored

**On-disk encrypted format:**

```
[EOMENC01 magic header][ciphertext][random padding]
```

The original PE `MZ` header is replaced; the file is not executable while encrypted. Padding is appended after ciphertext; metadata records exact `padding_size` for correct stripping on decrypt.

**Implementation:** `internal/fileops/fileops.go`, `internal/metadata/metadata.go`

---

## Timer-Based Locking

The duration argument sets a lock timer in **minutes**. At encryption, `unlock_time = now + duration` is written to metadata JSON on disk.

| Condition | Behavior |
|-----------|----------|
| **Before expiry** | File remains encrypted; re-encrypt shows lock status |
| **Immediate decrypt** | `decrypt` + correct master password restores anytime |
| **After expiry** | Next encrypt-mode run triggers auto-restore via `CheckExpiredAutoDecrypt` |

Timer state persists across process restarts — no background thread required.

**Implementation:** `internal/timer/timer.go`, metadata field `unlock_time`

---

## Directory Encryption

Directory targets are processed recursively via `filepath.WalkDir`. **Only `.exe` files** are encrypted; other file types are skipped.

Each file receives independent metadata (salt, nonce, algorithm). Decrypt on a directory restores all encrypted files found.

```bash
./dist/encrypt-o-matic.exe ./myapps AES 0 0-1000 5
./dist/encrypt-o-matic.exe decrypt ./myapps
```

**Implementation:** `internal/fileops/fileops.go` — `collectTargetFiles()`, `EncryptTarget()`, `DecryptTarget()`

---

## Compression Pipeline

**Encrypt:** `original bytes → gzip compress → encrypt → pad → write`

**Decrypt:** `read → strip padding → decrypt → gzip decompress → verify SHA-256 → restore`

Compression reduces ciphertext size for repetitive data. Metadata records `"compressed": true`.

**Implementation:** `internal/fileops/fileops.go` — `CompressData()`, `DecompressData()`

---

## Automated Testing

| Package | Coverage |
|---------|----------|
| `internal/crypto` | AES, ChaCha20, Twofish round-trip; wrong-password failure |
| `internal/auth` | bcrypt verification; master password setup |
| `internal/debug` | Metadata and backup listing |
| `internal/web` | Embedded static assets; path browse API |
| `tests/integration` | End-to-end workflow for all three algorithms |

```bash
make test
# or
go test ./...
```

The web dashboard **Reviewer Mode** page can run tests via `POST /api/reviewer/run-tests` and reports pass/fail counts.

Test fixture: `tests/testdata/demo.exe` (minimal PE stub).

---

## Web Dashboard

Optional local HTTP server for review and demonstration:

```bash
make server
# or
./dist/encrypt-o-matic.exe server --port 8080
```

| Route | Function |
|-------|----------|
| `/` | Dashboard |
| `/encrypt` | Encryption form with backend path picker |
| `/files` | Encrypted file listing |
| `/security` | Security layer documentation |
| `/reviewer` | Checklist and test runner |
| `/debug` | Runtime configuration |

All cryptographic operations execute in the Go backend. JavaScript handles UI only.

**Implementation:** `internal/web/`

---

## Bonus Features

Features beyond core requirements:

| Feature | Description |
|---------|-------------|
| `verify-password` | Test master password without encrypt/decrypt |
| `debug-info` | Config dir, hash status, metadata, backups |
| **Twofish-GCM** | Third encryption algorithm |
| **Auto-decrypt on timer expiry** | Metadata-driven restore when lock expires |
| **Web dashboard** | Reviewer-friendly educational UI |
| **Makefile** | `make test`, `make build`, `make server`, `make clean` |
| **Structured layout** | `cmd/`, `internal/`, `tests/`, `docs/` |
| **PROJECT_STRUCTURE.md** | Architecture reference |
| **Magic header** | `EOMENC01` identifies encrypted payload format |

---

## Ethics Statement

Encrypt-O-Matic was built for **authorized educational testing only**. It encrypts **only files explicitly provided by the user**. It does not spread, persist, escalate privileges, bypass antivirus, or target files automatically.

Understanding how encryption tools modify files helps defenders recognize ransomware-like techniques. This project is intended for that learning purpose, not for misuse.

---

## Edge Cases and Error Handling

| Scenario | Behavior |
|----------|----------|
| Path not found | `invalid path: ... does not exist` |
| Permission denied | `file permission error: ...` |
| Already encrypted | Lock status shown; auto-restore if timer expired |
| Corrupted metadata | `corrupted metadata: ...` |
| Hash mismatch | `hash mismatch after decryption` |
| Invalid algorithm | Lists supported algorithms |
| Directory with no `.exe` | `no supported files found in directory` |
| Empty file | Encrypted and restored; empty-content hash verified |

Backups in `.encryptomatic/backups/` provide rollback before any modification.

---

## Key Per-File Derivation

One master password is used for all operations. Each encrypted file receives a **unique 16-byte random salt** in metadata. PBKDF2(password, salt) produces a distinct 256-bit encryption key per file — same password, different salt, different ciphertext.

This is not full per-file random key wrapping (keys derived from password + salt, not independent random keys encrypted by a master key).

---

## Documentation Index

| Document | Audience | Contents |
|----------|----------|----------|
| [README.md](../README.md) | All | Setup, build, usage, troubleshooting |
| [PROJECT_STRUCTURE.md](PROJECT_STRUCTURE.md) | Developers / reviewers | Package layout, execution flows |
| [REVIEWER.md](REVIEWER.md) | Reviewers | This document |
| [STUDENT_NOTES.md](STUDENT_NOTES.md) | Presenters | Demo scripts, talking points, FAQ |
