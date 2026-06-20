package main

// encrypt-o-matic is an educational Windows CLI file encryption tool.
// It only processes paths explicitly provided by the user on the command line.
// For authorized educational testing only — never use on systems or files
// without explicit permission.
import (
	"fmt"
	"os"
	"strconv"

	"encrypt-o-matic/internal/auth"
	"encrypt-o-matic/internal/crypto"
	"encrypt-o-matic/internal/custom"
	"encrypt-o-matic/internal/debug"
	"encrypt-o-matic/internal/fileops"
	"encrypt-o-matic/internal/web"
)

const usageText = `# Encrypt-O-Matic

Educational File Encryption Tool


Description:
  Encrypt-O-Matic demonstrates secure file encryption using
  AES, ChaCha20, and Twofish.

  The tool only operates on files and directories explicitly
  provided by the user.


Safety Notice:
  • Educational and authorized testing only
  • No persistence
  • No privilege escalation
  • No self-spreading behavior
  • No automatic file targeting


Usage:
  encrypt-o-matic.exe <target_app> <algorithm> <size_mb> <custom_variable> <duration>

  encrypt-o-matic.exe decrypt <target_path>
  encrypt-o-matic.exe verify-password
  encrypt-o-matic.exe debug-info
  encrypt-o-matic.exe server [--port 8080]
  encrypt-o-matic.exe --help


Arguments:

  <target_app>
      Path to executable or directory

  <algorithm>
      AES
      ChaCha20
      Twofish

  <size_mb>
      Random padding size in megabytes

  <custom_variable>
      Range used by the SHA-256 workload
      Example: 0-100000

  <duration>
      Encryption duration in minutes


Commands:

  decrypt
      Immediately decrypt a file

  verify-password
      Verify configured master password

  debug-info
      Display diagnostic information

  server
      Start the local Web UI dashboard (bonus feature)


Examples:
  encrypt-o-matic.exe app.exe AES 10 0-100000 60

  encrypt-o-matic.exe app.exe ChaCha20 5 0-5000 30

  encrypt-o-matic.exe decrypt app.exe

  encrypt-o-matic.exe verify-password


Options:
  -h, --help
      Show this help message


Features:
  ✓ AES-256-GCM encryption
  ✓ ChaCha20-Poly1305 encryption
  ✓ Twofish-GCM encryption
  ✓ Password protection
  ✓ Compression before encryption
  ✓ File integrity verification
  ✓ Timer-based decryption
  ✓ Directory encryption support
`

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Error: missing arguments")
		fmt.Fprint(os.Stderr, usageText)
		os.Exit(1)
	}

	switch os.Args[1] {
	case "-h", "--help", "help":
		fmt.Print(usageText)
		return
	case "decrypt":
		if len(os.Args) != 3 {
			fmt.Fprintln(os.Stderr, "Error: decrypt requires exactly one target path")
			fmt.Fprint(os.Stderr, usageText)
			os.Exit(1)
		}
		if err := runDecryptCommand(os.Args[2]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	case "verify-password":
		if len(os.Args) != 2 {
			fmt.Fprintln(os.Stderr, "Error: verify-password takes no additional arguments")
			fmt.Fprint(os.Stderr, usageText)
			os.Exit(1)
		}
		if err := debug.RunVerifyPasswordCommand(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	case "debug-info":
		if len(os.Args) != 2 {
			fmt.Fprintln(os.Stderr, "Error: debug-info takes no additional arguments")
			fmt.Fprint(os.Stderr, usageText)
			os.Exit(1)
		}
		if err := debug.RunDebugInfoCommand(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	case "server", "web":
		port := "8080"
		for i := 2; i < len(os.Args); i++ {
			if os.Args[i] == "--port" && i+1 < len(os.Args) {
				port = os.Args[i+1]
				break
			}
		}
		srv := web.NewServer(":" + port)
		if err := srv.Start(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	if len(os.Args) != 6 {
		fmt.Fprintln(os.Stderr, "Error: encrypt mode requires 5 arguments")
		fmt.Fprint(os.Stderr, usageText)
		os.Exit(1)
	}

	target := os.Args[1]
	algorithmInput := os.Args[2]
	paddingInput := os.Args[3]
	customRange := os.Args[4]
	durationInput := os.Args[5]

	if err := runEncryptCommand(target, algorithmInput, paddingInput, customRange, durationInput); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runDecryptCommand(targetPath string) error {
	password, err := auth.Authenticate()
	if err != nil {
		return err
	}

	fmt.Println("Starting decryption...")
	return fileops.DecryptTarget(targetPath, password)
}

func runEncryptCommand(targetPath, algorithmInput, paddingInput, customRange, durationInput string) error {
	algorithm, err := crypto.NormalizeAlgorithm(algorithmInput)
	if err != nil {
		return err
	}

	paddingMB, err := strconv.Atoi(paddingInput)
	if err != nil || paddingMB < 0 {
		return fmt.Errorf("invalid size_manipulation: must be a non-negative integer (megabytes)")
	}

	durationMinutes, err := strconv.Atoi(durationInput)
	if err != nil || durationMinutes <= 0 {
		return fmt.Errorf("invalid duration: must be a positive integer (minutes)")
	}

	if _, _, err := custom.ParseRange(customRange); err != nil {
		return err
	}

	password, err := auth.Authenticate()
	if err != nil {
		return err
	}

	if err := fileops.CheckExpiredAutoDecrypt(targetPath, password); err != nil {
		return err
	}

	return fileops.EncryptTarget(targetPath, algorithm, password, paddingMB, durationMinutes, customRange)
}
