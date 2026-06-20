# Encrypt-O-Matic — Project Structure

This document explains the layout of the **encrypt-o-matic** project, what each package and file is responsible for, and how the pieces fit together.

> **Safety note:** This is an educational file encryption tool. It only processes paths explicitly provided by the user on the command line.

---

## High-level overview

Encrypt-O-Matic is a Go CLI application split into focused internal packages. There is no web server, no background service, and no automatic file targeting — everything runs from the command line and exits when the command completes.

```
User runs CLI command
        │
        ▼
cmd/encrypt-o-matic/main.go   ← parse args, route commands
        │
        ├── internal/auth/         ← master password prompt & verification
        ├── internal/custom/       ← SHA-256 workload loop
        ├── internal/fileops/      ← read, backup, encrypt, decrypt, restore
        ├── internal/crypto/       ← AES, ChaCha20, Twofish
        ├── internal/metadata/     ← JSON metadata per encrypted file
        ├── internal/timer/        ← unlock time and expiry checks
        ├── internal/debug/        ← verify-password & debug-info
        └── internal/config/       ← shared .encryptomatic path helpers
```

---

## Repository layout

```
encrypt-o-matic/
├── cmd/
│   └── encrypt-o-matic/
│       └── main.go              # CLI entry point and command routing
├── internal/
│   ├── auth/
│   │   ├── auth.go              # bcrypt password create/verify
│   │   └── auth_test.go
│   ├── config/
│   │   └── config.go            # .encryptomatic path constants
│   ├── crypto/
│   │   ├── crypto.go            # AES, ChaCha20, Twofish encryption
│   │   └── crypto_test.go
│   ├── custom/
│   │   └── custom.go            # SHA-256 workload loop
│   ├── debug/
│   │   ├── debug.go             # verify-password & debug-info
│   │   └── debug_test.go
│   ├── fileops/
│   │   └── fileops.go           # file I/O, compression, pipeline
│   ├── metadata/
│   │   └── metadata.go          # metadata struct and JSON persistence
│   └── timer/
│       └── timer.go             # unlock timer logic
├── tests/
│   ├── integration/
│   │   └── workflow_test.go     # end-to-end file encrypt/decrypt tests
│   └── testdata/
│       └── demo.exe             # sample binary for integration tests
├── docs/
│   └── PROJECT_STRUCTURE.md     # this file
├── README.md                    # user guide and troubleshooting
├── go.mod
├── go.sum
├── Makefile
└── .gitignore
```

Built binaries and runtime artifacts are **not** stored in the repository root:

| Output | Location |
|---|---|
| Compiled binary | `dist/encrypt-o-matic.exe` |
| Master password hash | `.encryptomatic/master.hash` (cwd at runtime) |
| Metadata | `.encryptomatic/metadata/` (cwd at runtime) |
| Backups | `.encryptomatic/backups/` (cwd at runtime) |

---

## Package reference

### `cmd/encrypt-o-matic`

**Role:** Application entry point.

- Defines the formatted `--help` output
- Parses CLI arguments and routes to encrypt, decrypt, verify-password, or debug-info
- Validates encrypt-mode arguments before delegating to internal packages
- Orchestrates authentication → file operation flow

**Key functions:** `main()`, `runEncryptCommand()`, `runDecryptCommand()`

---

### `internal/config`

**Role:** Shared configuration paths.

- Centralizes `.encryptomatic` directory name and subpaths
- Provides helpers: `RootDir()`, `MasterHashPath()`, `MetadataDir()`, `BackupsDir()`
- Avoids duplicating path logic across auth, metadata, fileops, and debug

---

### `internal/auth`

**Role:** Master password management.

- Hidden password prompt via `golang.org/x/term`
- First run: create bcrypt hash in `.encryptomatic/master.hash`
- Later runs: verify password against stored hash
- Exports `Authenticate()`, `VerifyStoredPassword()`, `PromptPassword()`
- Distinct error messages for hash not found, unreadable hash, and verification failure

---

### `internal/crypto`

**Role:** Low-level encryption primitives.

- **AES-256-GCM**, **ChaCha20-Poly1305**, **Twofish-GCM**
- PBKDF2-SHA256 key derivation (100,000 iterations)
- Unified `EncryptBytes()` / `DecryptBytes()` interface
- Algorithm normalization from user input

**Exported constants:** `AlgoAES`, `AlgoChaCha`, `AlgoTwofish`

---

### `internal/fileops`

**Role:** File pipeline and directory handling.

- gzip compress before encryption, decompress after
- Backup creation before any file modification
- Encrypted on-disk format: magic header + ciphertext + padding
- In-place encryption so executables cannot run while locked
- SHA-256 verification after decryption
- Directory support: recursively encrypts `.exe` files only
- Auto-decrypt when timer has expired

**Key exports:** `EncryptTarget()`, `DecryptTarget()`, `CheckExpiredAutoDecrypt()`

---

### `internal/metadata`

**Role:** Per-file encryption metadata.

- `FileMetadata` struct: algorithm, nonce, salt, padding, hash, timer, permissions
- JSON files stored at `.encryptomatic/metadata/<sha256-of-path>.json`
- `Save()`, `Load()`, `Remove()`, `IsEncrypted()`

---

### `internal/timer`

**Role:** Time-based lock.

- `ComputeUnlockTime()` — current time + duration minutes
- `IsUnlockExpired()` — check if auto-decrypt is allowed
- `FormatUnlockStatus()` — human-readable lock status

---

### `internal/custom`

**Role:** Educational CPU workload.

- Parses `0-100000` style ranges
- SHA-256 loop with periodic progress output
- `ParseRange()`, `RunOperation()`

---

### `internal/debug`

**Role:** Diagnostic commands.

- `RunVerifyPasswordCommand()` — test password without touching files
- `RunDebugInfoCommand()` — show config dir, hash existence, metadata, backups
- Never prints passwords or hash contents

---

## Tests

| Location | What it covers |
|---|---|
| `internal/auth/auth_test.go` | Password verification outcomes |
| `internal/crypto/crypto_test.go` | AES / ChaCha20 / Twofish round-trips |
| `internal/debug/debug_test.go` | Metadata and backup file listing |
| `tests/integration/workflow_test.go` | Full file write/read workflow using `tests/testdata/demo.exe` |

Run all tests:

```bash
go test -v ./...
# or
make test
```

---

## Build commands

```bash
# Run tests
make test

# Build binary to dist/
make build

# Remove build output and local runtime artifacts
make clean
```

Manual build:

```bash
go build -o dist/encrypt-o-matic.exe ./cmd/encrypt-o-matic
```

---

## Typical execution flows

### Encrypt a file

```
cmd/main.go
  → auth.Authenticate()
  → fileops.CheckExpiredAutoDecrypt()
  → custom.RunOperation()
  → fileops.EncryptTarget()
      → compress → crypto.EncryptBytes()
      → backup → write payload → metadata.Save()
```

### Decrypt a file

```
cmd/main.go
  → auth.Authenticate()
  → fileops.DecryptTarget()
      → metadata.Load()
      → crypto.DecryptBytes()
      → decompress → verify SHA-256 → restore → metadata.Remove()
```

---

## Design principles

1. **`internal/` packages** — application logic is not importable by external modules.
2. **Single CLI entry point** — `cmd/encrypt-o-matic/main.go` wires packages together.
3. **Clean root** — source and docs only; binaries, test artifacts, and runtime data are gitignored.
4. **Explicit targeting** — no file is touched unless the user provides its path.
5. **Safe file operations** — backup before modify, verify hash after restore.
