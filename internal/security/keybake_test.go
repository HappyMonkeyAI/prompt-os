package security

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writePrivate(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

func TestBakeSecretWritesModeAndPath(t *testing.T) {
	root := t.TempDir()
	res, err := BakeSecret(root, SecretKindAPIKey, []byte("OPENAI_API_KEY=sk-test"))
	if err != nil {
		t.Fatalf("bake failed: %v", err)
	}
	if !strings.HasSuffix(res.Path, filepath.Join("etc", "environment.d", "ai-keys.conf")) {
		t.Fatalf("unexpected path: %s", res.Path)
	}
	info, err := os.Stat(res.Path)
	if err != nil {
		t.Fatalf("stat failed: %v", err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("expected mode 0600, got %o", info.Mode().Perm())
	}
}

func TestBakeSecretRejectsEmpty(t *testing.T) {
	_, err := BakeSecret(t.TempDir(), SecretKindAPIKey, []byte{})
	if err == nil {
		t.Fatal("expected error for empty key")
	}
}

func TestBakeSecretBacksUpExisting(t *testing.T) {
	root := t.TempDir()
	dest := filepath.Join(root, "etc", "environment.d", "ai-keys.conf")
	if err := writePrivate(dest, []byte("old")); err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	res, err := BakeSecret(root, SecretKindAPIKey, []byte("OPENAI_API_KEY=sk-test"))
	if err != nil {
		t.Fatalf("bake failed: %v", err)
	}
	if !res.Backed {
		t.Fatalf("expected existing backup")
	}
	if _, err := os.Stat(res.Backup); err != nil {
		t.Fatalf("backup missing: %v", err)
	}
}

func TestBakeSecretRejectsSymlinkTarget(t *testing.T) {
	root := t.TempDir()
	dest := filepath.Join(root, "etc", "environment.d", "ai-keys.conf")
	if err := os.MkdirAll(filepath.Dir(dest), 0o700); err != nil {
		t.Fatalf("setup mkdir failed: %v", err)
	}
	if err := os.Symlink("/tmp/outside", dest); err != nil {
		t.Fatalf("setup symlink failed: %v", err)
	}

	_, err := BakeSecret(root, SecretKindAPIKey, []byte("OPENAI_API_KEY=sk-test"))
	if !errors.Is(err, ErrUnsafePath) {
		t.Fatalf("expected ErrUnsafePath, got %v", err)
	}
}

func TestScrubRemovesFiles(t *testing.T) {
	root := t.TempDir()
	if _, err := BakeSecret(root, SecretKindAPIKey, []byte("x=1")); err != nil {
		t.Fatalf("setup bake failed: %v", err)
	}
	if err := Scrub(root, SecretKindAPIKey); err != nil {
		t.Fatalf("scrub failed: %v", err)
	}
	p := filepath.Join(root, "etc", "environment.d", "ai-keys.conf")
	if _, err := os.Stat(p); !os.IsNotExist(err) {
		t.Fatalf("expected secret removed, got err=%v", err)
	}
}

func TestQuoteSurroundEscape(t *testing.T) {
	in := "it's ok"
	want := "'it'\\''s ok'"
	if got := QuoteSurround(in); got != want {
		t.Fatalf("unexpected quote: %s", got)
	}
}
