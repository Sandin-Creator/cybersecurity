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

## Setup

### Prerequisites

- Go 1.21 or newer
- Windows (primary target platform)

### Clone and fetch dependencies

```powershell
git clone <your-repo-url>
cd encrypt-o-matic
go mod tidy
make test
```

## Project layout

Source code lives under `cmd/` and `internal/`. Built binaries go to `dist/`. Runtime data (`.encryptomatic/`) is created in your **working directory** when you run the tool — not in the repository root.

See [docs/PROJECT_STRUCTURE.md](docs/PROJECT_STRUCTURE.md) for a full breakdown of each package and file.

## Windows build instructions

From the project directory:

```powershell
go build -o dist/encrypt-o-matic.exe ./cmd/encrypt-o-matic
```

Or use the Makefile:

```powershell
make build
```

Cross-compile from Linux/macOS:

```bash
GOOS=windows GOARCH=amd64 go build -o dist/encrypt-o-matic.exe ./cmd/encrypt-o-matic
```

Verify the binary:

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
|---|---|
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

## Encryption and decryption flow

### Encryption

1. Prompt for master password (hidden input)
2. Verify/create bcrypt hash in `.encryptomatic/master.hash`
3. Run the custom SHA-256 loop (`custom_variable`)
4. For each target file:
   - Create a backup in `.encryptomatic/backups/`
   - Read original bytes and compute SHA-256
   - gzip compress the bytes
   - Encrypt compressed data (AES-256-GCM or ChaCha20-Poly1305)
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

```
.encryptomatic/
  master.hash          # bcrypt hash of master password
  backups/             # pre-modification backups
  metadata/            # per-file JSON metadata
```

## Supported algorithms

| Algorithm | Mode | Status |
|---|---|---|
| AES | AES-256-GCM | Implemented |
| ChaCha20 | ChaCha20-Poly1305 | Implemented |
| Twofish | Twofish-GCM (`golang.org/x/crypto/twofish`) | Implemented |

All three algorithms use the same encrypt/decrypt pipeline: gzip compress → encrypt → store nonce and salt in metadata → decrypt → decompress → verify SHA-256.

## Automated test workflow

Run the Go tests to verify AES, ChaCha20, and Twofish round-trips on a small test binary:

```powershell
go test -v ./...
```

The tests:

1. Create a small test `.exe`-like binary payload in a temp directory
2. Encrypt it with **AES**, decrypt it, and verify SHA-256
3. Repeat with **ChaCha20**
4. Repeat with **Twofish**

You should see passing subtests for `AES`, `ChaCha20`, and `Twofish`.

## Twofish demo (manual)

Create a demo folder and test file:

```powershell
mkdir C:\encrypt-demo
Set-Content -Path C:\encrypt-demo\demo.exe -Value "MZ test payload for encrypt-o-matic" -NoNewline
cd C:\encrypt-demo
```

Encrypt with Twofish (1-minute lock, no extra padding, small custom range):

```powershell
C:\path\to\encrypt-o-matic.exe C:\encrypt-demo\demo.exe Twofish 0 0-1000 1
```

Confirm the file is no longer usable as the original executable, then decrypt immediately:

```powershell
C:\path\to\encrypt-o-matic.exe decrypt C:\encrypt-demo\demo.exe
```

Verify the restored bytes match the original:

```powershell
Get-FileHash C:\encrypt-demo\demo.exe -Algorithm SHA256
```

Repeat the same workflow with `AES` and `ChaCha20` to compare all three algorithms.

## Troubleshooting

### Verify the master password

Use `verify-password` to test authentication without encrypting or decrypting anything:

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

This command checks the password against `.encryptomatic/master.hash` in your **current working directory**. Run it from the same folder where you created the master password.

### Inspect local state with debug-info

```powershell
encrypt-o-matic.exe debug-info
```

Example output:

```text
Encrypt-O-Matic Debug Information

Config directory:
  C:\encrypt-demo\.encryptomatic

Master hash:
  Exists: Yes
  Path: C:\encrypt-demo\.encryptomatic\master.hash

Metadata files:
  abc123....json (C:\encrypt-demo\demo.exe)

Backup files:
  demo.exe.20260620-190622.bak
```

This command never prints the password or bcrypt hash.

### Common password issues

| Symptom | Likely cause | What to do |
|---|---|---|
| `Password INVALID` from `verify-password` | Wrong password entered | Retry carefully; password input is hidden |
| `master password hash file not found` | No `.encryptomatic/master.hash` in current directory | Run from the directory where you first set the password, or reset and create a new one |
| `master password hash file unreadable` | Permission problem | Check file permissions on `.encryptomatic/master.hash` |
| `Password verification failed.` | bcrypt mismatch during encrypt/decrypt | Use `verify-password` to confirm the correct password |
| `decryption failed after successful authentication` | Password auth succeeded but ciphertext/key derivation failed | Confirm you are decrypting with the same password used for encryption and that metadata matches the file |

During decrypt, successful authentication prints:

```text
Password verified successfully.
Starting decryption...
```

If you see the first line but decryption still fails, the problem is in the encrypted file or metadata — not the master password hash check.

### Reset the test environment

To start fresh in a lab folder, delete the local config directory from your working directory:

```powershell
Remove-Item -Recurse -Force .\.encryptomatic
```

This removes:

- `master.hash`
- metadata files
- backup copies

It does **not** restore encrypted executables automatically. Decrypt files first if you still need them, then delete `.encryptomatic`.

### Create a new master password

1. Delete `.encryptomatic` (see above), or use a new working directory
2. Run any encrypt or decrypt command
3. Enter and confirm a new password when prompted

You will see:

```text
Master password created successfully.
```

After that, `verify-password` should report `Password OK` for the new password.

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

## Review / demo instructions

For classroom or lab demos:

1. Build the binary on Windows
2. Create a test folder, for example `C:\encrypt-demo\`
3. Copy a harmless test executable into that folder
4. Run encryption with a short duration (for example `1` minute)
5. Confirm the executable no longer runs
6. Run `decrypt` with the master password and verify restoration
7. Wait for timer expiry and run encrypt mode again to observe auto-decrypt
8. Inspect `.encryptomatic/backups/` and `.encryptomatic/metadata/` to explain the pipeline

**Demo rule:** Only use executables and directories you created for the lab. Never point the tool at production software or shared systems without authorization.

## License

Educational use only. Use responsibly.
