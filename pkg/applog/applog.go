package applog

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

// Init wires the standard logger to write to both stderr and a file named
// pokertips.log. The file lives next to the built binary — on macOS that
// means next to pokertips.app, not inside it — so users can inspect it
// without digging into the bundle.
//
// It returns the resolved log file path and a close func to defer.
func Init() (string, func(), error) {
	dir, err := LogDir()
	if err != nil {
		return "", nil, err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", nil, fmt.Errorf("mkdir log dir %q: %w", dir, err)
	}

	path := filepath.Join(dir, "pokertips.log")
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return "", nil, fmt.Errorf("open log file %q: %w", path, err)
	}

	log.SetOutput(io.MultiWriter(os.Stderr, f))
	log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.Lshortfile)
	log.Printf("=== pokertips started at %s (%s/%s) ===",
		time.Now().Format(time.RFC3339), runtime.GOOS, runtime.GOARCH)

	return path, func() { _ = f.Close() }, nil
}

// LogDir resolves the directory where the log file should be written. On
// macOS, when the binary is inside an .app bundle, it walks up past the
// bundle so the log ends up next to pokertips.app (typically build/bin/).
// On other platforms, or outside a bundle, it uses the executable's
// directory. If the directory isn't writable, it falls back to the user
// config dir.
func LogDir() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("resolve executable path: %w", err)
	}
	if resolved, err := filepath.EvalSymlinks(exe); err == nil {
		exe = resolved
	}

	dir := filepath.Dir(exe)
	if runtime.GOOS == "darwin" {
		if bundleParent, ok := macOSBundleParent(dir); ok {
			dir = bundleParent
		}
	}

	if writable(dir) {
		return dir, nil
	}
	if fallback, err := os.UserConfigDir(); err == nil {
		alt := filepath.Join(fallback, "pokertips")
		return alt, nil
	}
	return dir, nil
}

// macOSBundleParent detects whether dir is the MacOS/ folder inside a .app
// bundle and, if so, returns the directory that contains the bundle itself.
func macOSBundleParent(dir string) (string, bool) {
	if filepath.Base(dir) != "MacOS" {
		return "", false
	}
	contents := filepath.Dir(dir)
	if filepath.Base(contents) != "Contents" {
		return "", false
	}
	bundle := filepath.Dir(contents)
	if filepath.Ext(bundle) != ".app" {
		return "", false
	}
	return filepath.Dir(bundle), true
}

func writable(dir string) bool {
	probe, err := os.CreateTemp(dir, ".write-check-*")
	if err != nil {
		return false
	}
	name := probe.Name()
	_ = probe.Close()
	_ = os.Remove(name)
	return true
}
