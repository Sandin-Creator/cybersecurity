# encrypt-o-matic

Educational Windows CLI file encryption tool written in Go. It encrypts **only** files or directories you explicitly provide on the command line.

> **Safety / ethics:** This project is for **authorized educational testing only**. It is **not** malware. It does not include stealth, persistence, spreading, privilege escalation, antivirus bypass, or automatic targeting. Never run it against systems or files you do not own or have explicit permission to modify.

## Project overview

`encrypt-o-matic` demonstrates safe, reversible file encryption workflows:

- Password-protected encryption with bcrypt-stored master password verification
- AES-256-GCM, ChaCha20-Poly1305, and Twofish-GCM encryption
- gzip compression before encryption
- Random padding to increase encrypted file size
- Metadata storage for exact restoration
- Time-based automatic decryption after a configured duration
- Immediate password-based decryption via a `decrypt` subcommand
- Directory support (`.exe` files only by default)
- **Reviewer Dashboard Web UI** — local HTTP dashboard for review and demonstration
- **Interactive encryption workflow visualization**
- **Security layer visualization**
- **Reviewer Mode and requirement checklist**
- **Debug and verification tools** (`verify-password`, `debug-info`)

For a structured review guide, see [docs/REVIEWER.md](docs/REVIEWER.md).

## Quick Start

Build:

```bash
make build
```

Run automated tests:

```bash
make test
```

Start the reviewer dashboard:

```bash
make server
```

Open:

```text
http://localhost:8080
```

This launches the reviewer-focused dashboard while keeping the CLI fully functional.

---

## Setup

### Prerequisites

- Go 1.21 or newer

### Development environment

The primary target platform is **Windows**.

Development and testing were performed using:

- Go 1.21+
- Windows
- WSL/Linux

The application can be built natively on Windows or cross-compiled from Linux/macOS.

### Clone and fetch dependencies

```bash
git clone <your-repo-url>
cd encrypt-o-matic
go mod tidy
make test
```

## Project layout

Source code lives under `cmd/` and `internal/`. Built binaries go to `dist/`. Runtime data (`.encryptomatic/`) is created in your **working directory** when you run the tool — not in the repository root.

See [docs/PROJECT_STRUCTURE.md](docs/PROJECT_STRUCTURE.md) for a full breakdown of each package and file.

## Build

From the project directory:

```bash
make build
```

Or build directly with Go:

```bash
go build -o dist/encrypt-o-matic.exe ./cmd/encrypt-o-matic
```

Cross-compile from Linux/macOS:

```bash
GOOS=windows GOARCH=amd64 go build -o dist/encrypt-o-matic.exe ./cmd/encrypt-o-matic
```

Verify the binary (Windows):

```powershell
.\dist\encrypt-o-matic.exe --help
```

## Usage

### Encrypt a file

```powershell
encrypt-o-matic.exe C:\path\to\app.exe AES 10 0-100000 60
```

Arguments:

| Argument | Description |
|----------|-------------|
| `target_app` | Path to a Windows executable or directory |
| `encryption_algorithm` | `AES`, `ChaCha20`, or `Twofish` |
| `size_manipulation` | Megabytes of random padding to append |
| `custom_variable` | Range like `0-100000` for a SHA-256 workload loop |
| `duration` | Lock duration in minutes before auto-decrypt is allowed |

### Decrypt immediately with password

```powershell
encrypt-o-matic.exe decrypt C:\path\to\app.exe
```

### Help

```powershell
encrypt-o-matic.exe -h
encrypt-o-matic.exe --help
```

## Reviewer Dashboard (Bonus Feature)

The Web UI is a reviewer-focused dashboard built on top of the same Go backend used by the CLI.

**Benefits:**

- Visual encryption workflow
- Security layer explanations
- Reviewer checklist with automated test runner
- Metadata inspection
- Debug tools
- File management dashboard
- Local backend path picker (no manual path typing required for demos)

No encryption logic runs in JavaScript. All operations are performed by the Go backend.

Start the server:

```bash
make server
```

Or directly:

```bash
./dist/encrypt-o-matic.exe server --port 8080
```

Open [http://localhost:8080](http://localhost:8080) in your browser.

| Route | Page |
|-------|------|
| `/` | Dashboard |
| `/encrypt` | Encryption workflow |
| `/files` | Encrypted files |
| `/security` | Security visualization |
| `/reviewer` | Reviewer mode |
| `/debug` | Debug tools |

## Reviewer Workflow

Recommended review process:

1. Run automated tests — `make test`
2. Launch the Web Dashboard — `make server`
3. Review the **Security** page
4. Review the **Reviewer Mode** checklist
5. Encrypt a demo executable (CLI or Web UI)
6. Verify file metadata on the **Encrypted Files** page
7. Decrypt and verify restoration

This provides a complete walkthrough of all assignment requirements.


## Encryption and decryption flow

### Encryption

1. Prompt for master password (hidden input)
2. Verify/create bcrypt hash in `.encryptomatic/master.hash`
3. Run the custom SHA-256 loop (`custom_variable`)
4. For each target file:
   - Create a backup in `.encryptomatic/backups/`
   - Read original bytes and compute SHA-256
   - gzip compress the bytes
   - Encrypt compressed data (AES-256-GCM, ChaCha20-Poly1305, or Twofish-GCM)
   - Append random padding (`size_manipulation` MB)
   - Replace the original file path with encrypted content
   - Save metadata JSON in `.encryptomatic/metadata/`

The original executable becomes unusable while encrypted, but remains restorable byte-for-byte.

### Decryption

1. Authenticate with master password
2. Load metadata for the target file
3. Read encrypted payload and remove padding
4. Decrypt and decompress
5. Verify SHA-256 against stored original hash
6. Restore original bytes to the original path
7. Remove metadata entry

If verification fails, the tool reports a **hash mismatch** and does not silently accept corrupted output.

## Password handling

- First run: you set a master password; a bcrypt hash is stored locally at `.encryptomatic/master.hash`
- Later runs: entered password is verified against the stored hash
- Wrong password: access denied, program exits
- Password input is not echoed to the terminal

Per-file encryption keys are derived with PBKDF2-SHA256 using a per-file salt stored in metadata.

## Timer behavior

After encryption, metadata stores:

`unlock_time = current_time + duration_minutes`

- Before `unlock_time`: file stays encrypted unless you run `decrypt`
- After `unlock_time`: running encrypt mode again on that target triggers **automatic decryption**
- You can always decrypt immediately with `decrypt <target_path>` and the master password

## Directory encryption

If `target_app` is a directory:

- The tool walks the directory recursively
- Only `.exe` files are encrypted by default (safety measure)
- Each processed file is printed to the console
- Decryption restores all encrypted files under the directory

Example:

```powershell
encrypt-o-matic.exe C:\Apps\MyTool ChaCha20 5 0-5000 30
encrypt-o-matic.exe decrypt C:\Apps\MyTool
```

## Compression

Original file bytes are gzip-compressed before encryption to reduce ciphertext size and demonstrate a realistic pipeline. Metadata records `"compressed": true` so decryption can decompress correctly.

## Custom operation

The `custom_variable` range (for example `0-100000`) triggers a safe CPU workload:

- For each integer in range, compute SHA-256 of the number as a string
- Print progress periodically
- Yield briefly during large ranges to avoid unnecessary freezing

This is intentionally time-consuming but harmless.

## Local data layout

```text
.encryptomatic/
  master.hash          # bcrypt hash of master password
  backups/             # pre-modification backups
  metadata/            # per-file JSON metadata
```

## Supported algorithms

| Algorithm | Mode | Status |
|-----------|------|--------|
| AES | AES-256-GCM | Implemented |
| ChaCha20 | ChaCha20-Poly1305 | Implemented |
| Twofish | Twofish-GCM (`golang.org/x/crypto/twofish`) | Implemented |

All three algorithms use the same encrypt/decrypt pipeline: gzip compress → encrypt → store nonce and salt in metadata → decrypt → decompress → verify SHA-256.

## Algorithm demonstration

All supported algorithms can be tested using the same workflow:

- AES-256-GCM
- ChaCha20-Poly1305
- Twofish-GCM

Example:

```powershell
encrypt-o-matic.exe demo.exe AES 5 0-3000 2
encrypt-o-matic.exe decrypt demo.exe
```

Replace `AES` with `ChaCha20` or `Twofish` to compare algorithms. The automated test suite validates all three.

## Automated testing

Run the Go test suite:

```bash
make test
```

Or:

```bash
go test -v ./...
```

Tests verify AES, ChaCha20, and Twofish round-trips on a minimal test binary in `tests/testdata/demo.exe`. The Web UI **Reviewer Mode** page can also run tests via the dashboard.

## Troubleshooting

### Verify the master password

```powershell
encrypt-o-matic.exe verify-password
```

Expected output:

```text
Password OK
```

If the password is wrong:

```text
Password INVALID
```

This command checks the password against `.encryptomatic/master.hash` in your **current working directory**.

### Inspect local state with debug-info

```powershell
encrypt-o-matic.exe debug-info
```

This command never prints the password or bcrypt hash. Use the Web UI **Debug** page for the same information in a visual format.

### Common password issues

| Symptom | Likely cause | What to do |
|---------|--------------|------------|
| `Password INVALID` from `verify-password` | Wrong password entered | Retry carefully; password input is hidden |
| `master password hash file not found` | No `.encryptomatic/master.hash` in current directory | Run from the directory where you first set the password |
| `Password verification failed.` | bcrypt mismatch during encrypt/decrypt | Use `verify-password` to confirm the correct password |
| `decryption failed after successful authentication` | Ciphertext or metadata issue | Confirm metadata matches the file and the same password was used for encryption |

### Reset the test environment

Delete the local config directory from your working directory:

```powershell
Remove-Item -Recurse -Force .\.encryptomatic
```

Decrypt files first if you still need them. To create a new master password, delete `.encryptomatic` (or use a new working directory) and run any encrypt or decrypt command.

## Error handling

The tool provides friendly errors for:

- Missing/invalid CLI arguments
- Invalid paths
- Unsupported algorithms
- Wrong password
- Invalid custom variable range
- Invalid duration or padding values
- File permission problems
- Corrupted metadata
- Hash mismatch after decryption

## Project highlights

- AES-256-GCM support
- ChaCha20-Poly1305 support
- Twofish-GCM support
- bcrypt password protection
- PBKDF2 key derivation with salt
- SHA-256 integrity verification
- Compression before encryption
- Timer-based file locking
- Directory encryption
- Backup and recovery system
- Automated testing
- Reviewer Dashboard Web UI
- Debug and verification tools
- Standard Go project structure

## License

Educational use only. Use responsibly.
