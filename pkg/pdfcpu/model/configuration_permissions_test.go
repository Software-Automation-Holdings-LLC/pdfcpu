package model

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestConfigurationDirectoryPermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Unix directory permission bits")
	}
	for _, operation := range []string{"load", "reset", "legacy"} {
		t.Run(operation, func(t *testing.T) {
			preserveConfigurationGlobals(t)
			root := t.TempDir()
			var err error
			switch operation {
			case "load":
				_, err = LoadConfiguration(root)
			case "reset":
				_, err = ResetConfiguration(root)
			case "legacy":
				err = EnsureDefaultConfigAt(root, false)
			}
			if err != nil {
				t.Fatal(err)
			}
			for _, subdir := range []string{"pdfcpu", "pdfcpu/fonts", "pdfcpu/certs"} {
				info, err := os.Stat(filepath.Join(root, subdir))
				if err != nil {
					t.Fatal(err)
				}
				if got := info.Mode().Perm(); got&0o077 != 0 {
					t.Fatalf("%s permissions = %o, want owner-only access", subdir, got)
				}
			}
		})
	}
}

func TestConfigurationPreservesExistingDirectoryPermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Unix directory permission bits")
	}
	root := t.TempDir()
	configDir := filepath.Join(root, "pdfcpu")
	if err := os.Mkdir(configDir, 0o750); err != nil {
		t.Fatal(err)
	}
	before, err := os.Stat(configDir)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := LoadConfiguration(root); err != nil {
		t.Fatal(err)
	}
	after, err := os.Stat(configDir)
	if err != nil {
		t.Fatal(err)
	}
	if before.Mode().Perm() != after.Mode().Perm() {
		t.Fatalf("existing permissions changed: %o -> %o", before.Mode().Perm(), after.Mode().Perm())
	}
}
