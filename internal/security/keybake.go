package security

import (
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

var (
	ErrEmptyKey       = errors.New("security: key is empty")
	ErrUnsupportedExt = errors.New("security: unsupported secret extension")
	ErrBackupFailed   = errors.New("security: backup failed")
	ErrUnsafePath     = errors.New("security: unsafe secret path")
)

const keyPerm = 0o600
const maxSecretBytes = 4096

type SecretKind string

const (
	SecretKindAPIKey SecretKind = "api_key"
	SecretKindSSH    SecretKind = "ssh_private"
	SecretKindCloud  SecretKind = "cloud_cred"
)

type BakeResult struct {
	Path    string
	Backup  string
	Backed  bool
	Kind    SecretKind
	Written bool
}

func defaultDest(kind SecretKind) (string, error) {
	switch kind {
	case SecretKindAPIKey:
		return "/etc/environment.d/ai-keys.conf", nil
	case SecretKindSSH:
		return "/root/.ssh/id_ed25519", nil
	case SecretKindCloud:
		return "/etc/promptos/cloud-cred", nil
	default:
		return "", ErrUnsupportedExt
	}
}

func BakeSecret(root string, kind SecretKind, data []byte) (BakeResult, error) {
	if len(data) == 0 {
		return BakeResult{}, ErrEmptyKey
	}
	if len(data) > maxSecretBytes {
		return BakeResult{}, fmt.Errorf("security: secret too large: %d > %d", len(data), maxSecretBytes)
	}

	dest, err := defaultDest(kind)
	if err != nil {
		return BakeResult{}, err
	}
	out := filepath.Join(root, dest)
	if err := rejectSymlinkPath(root, out); err != nil {
		return BakeResult{}, err
	}
	if err := mkdirp(filepath.Dir(out)); err != nil {
		return BakeResult{}, err
	}
	if err := rejectSymlinkPath(root, out); err != nil {
		return BakeResult{}, err
	}

	res := BakeResult{Kind: kind}
	if st, err := os.Stat(out); err == nil && !st.IsDir() {
		bak := out + ".bak"
		src, err := os.Open(out)
		if err != nil {
			return res, fmt.Errorf("open: %w", err)
		}
		defer src.Close()
		dst, err := os.OpenFile(bak, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, keyPerm)
		if err != nil {
			return res, fmt.Errorf("%w: %v", ErrBackupFailed, err)
		}
		defer dst.Close()
		if _, err := io.Copy(dst, src); err != nil {
			return res, fmt.Errorf("%w: %v", ErrBackupFailed, err)
		}
		res.Backup = bak
		res.Backed = true
	}

	f, err := os.CreateTemp(filepath.Dir(out), filepath.Base(out)+".tmp-*")
	if err != nil {
		return res, fmt.Errorf("create tmp: %w", err)
	}
	tmp := f.Name()

	n, writeErr := f.Write(data)
	closeErr := f.Close()
	if writeErr != nil {
		_ = os.Remove(tmp)
		return res, fmt.Errorf("write tmp: %w", writeErr)
	}
	if closeErr != nil {
		_ = os.Remove(tmp)
		return res, fmt.Errorf("close tmp: %w", closeErr)
	}
	if n != len(data) {
		_ = os.Remove(tmp)
		return res, fmt.Errorf("short write: %d != %d", n, len(data))
	}
	if err := os.Rename(tmp, out); err != nil {
		_ = os.Remove(tmp)
		return res, fmt.Errorf("rename: %w", err)
	}
	_ = os.Chmod(out, keyPerm)
	res.Path = out
	res.Written = true
	return res, nil
}

func rejectSymlinkPath(root, target string) error {
	root, err := filepath.Abs(filepath.Clean(root))
	if err != nil {
		return err
	}
	rootInfo, err := os.Lstat(root)
	if err == nil && rootInfo.Mode()&os.ModeSymlink != 0 {
		return ErrUnsafePath
	}
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}

	target, err = filepath.Abs(filepath.Clean(target))
	if err != nil {
		return err
	}
	rel, err := filepath.Rel(root, target)
	if err != nil {
		return err
	}
	if rel == "." {
		return nil
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) || filepath.IsAbs(rel) {
		return ErrUnsafePath
	}

	current := root
	for _, part := range strings.Split(rel, string(os.PathSeparator)) {
		current = filepath.Join(current, part)
		info, err := os.Lstat(current)
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		if err != nil {
			return err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return ErrUnsafePath
		}
	}
	return nil
}

func mkdirp(p string) error {
	if p == "" || p == "/" {
		return nil
	}
	if err := os.MkdirAll(p, 0o700); err != nil {
		return err
	}
	return nil
}

func Scrub(root string, kind SecretKind) error {
	dest, err := defaultDest(kind)
	if err != nil {
		return err
	}
	p := filepath.Join(root, dest)
	if _, err := os.Stat(p); errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err := os.Remove(p); err != nil {
		return err
	}
	_ = os.Remove(p + ".bak")
	return nil
}

func WipeOverwrite(b []byte) {
	if b == nil {
		return
	}
	_, _ = rand.Read(b)
	for i := range b {
		b[i] = 0
	}
}

func QuoteSurround(s string) string {
	s = strings.ReplaceAll(s, "'", "'\\''")
	return "'" + s + "'"
}
