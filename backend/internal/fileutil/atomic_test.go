package fileutil

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteAtomicCreatesDirectoriesAndPrivateFiles(t *testing.T) {
	path := filepath.Join(t.TempDir(), "deep", "er", "data.json")
	if err := WriteAtomic(path, []byte("one"), 0o600); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(path)
	if err != nil || string(got) != "one" {
		t.Fatalf("read back %q, %v", got, err)
	}
	info, _ := os.Stat(path)
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("mode %v, want 0600", info.Mode().Perm())
	}
	dirInfo, _ := os.Stat(filepath.Dir(path))
	if dirInfo.Mode().Perm() != 0o700 {
		t.Fatalf("directory mode %v, want 0700", dirInfo.Mode().Perm())
	}
}

func TestWriteAtomicReplacesWithoutLeavingTemporaryFiles(t *testing.T) {
	path := filepath.Join(t.TempDir(), "data.json")
	for _, content := range []string{"first", "second, longer"} {
		if err := WriteAtomic(path, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	if got, _ := os.ReadFile(path); string(got) != "second, longer" {
		t.Fatalf("got %q", got)
	}
	entries, _ := os.ReadDir(filepath.Dir(path))
	if len(entries) != 1 {
		t.Fatalf("temporary files left behind: %v", entries)
	}
}

func TestWriteAtomicFailsWhenTheDirectoryCannotBeCreated(t *testing.T) {
	blocker := filepath.Join(t.TempDir(), "file")
	os.WriteFile(blocker, nil, 0o600)
	if err := WriteAtomic(filepath.Join(blocker, "data.json"), []byte("x"), 0o600); err == nil {
		t.Fatal("expected an error when a parent is a file")
	}
}
