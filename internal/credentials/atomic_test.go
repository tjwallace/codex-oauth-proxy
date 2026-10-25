package credentials

import (
	"os"
	"path/filepath"
	"testing"
)

func TestHomeOpsAtomicCredentials(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "auth.json")
	if err := atomicWriteCredentials(path, []byte("old")); err != nil {
		t.Fatal(err)
	}
	// Holding the old inode open distinguishes replacement from truncation.
	old, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer old.Close()
	if err := atomicWriteCredentials(path, []byte("new")); err != nil {
		t.Fatal(err)
	}
	buffer := make([]byte, 3)
	if _, err := old.Read(buffer); err != nil || string(buffer) != "old" {
		t.Fatalf("old inode modified: %q %v", buffer, err)
	}
	got, err := os.ReadFile(path)
	if err != nil || string(got) != "new" {
		t.Fatalf("new contents: %q %v", got, err)
	}
	info, err := os.Stat(path)
	if err != nil || info.Mode().Perm() != 0600 {
		t.Fatalf("credentials must have mode 0600: %v", err)
	}
	entries, err := os.ReadDir(dir)
	if err != nil || len(entries) != 1 {
		t.Fatalf("temporary files left behind: %v", err)
	}
}
