package execute

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/HappyMonkeyAI/prompt-os/internal/llm"
)

func TestBuildConfigDropPlanMapsUnderMount(t *testing.T) {
	bp := &llm.Blueprint{
		Configs: map[string]string{
			"/etc/environment.d/ai-keys.conf": "OPENAI_API_KEY=sk-test",
			"/opt/promptos/hello.txt":           "hi",
		},
	}
	files, err := BuildConfigDropPlan(ConfigDropOptions{Blueprint: bp, MountRoot: "/mnt/target"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(files) != 2 {
		t.Fatalf("expected 2 files, got %v", files)
	}
	want := filepath.Join("/mnt/target", "etc/environment.d/ai-keys.conf")
	if files[0] != want {
		t.Fatalf("expected sorted first file %s, got %s", want, files[0])
	}
}

func TestApplyConfigDropRejectsUnsafePath(t *testing.T) {
	bp := &llm.Blueprint{Configs: map[string]string{"/etc/../shadow": "bad"}}
	_, err := ApplyConfigDrop(ConfigDropOptions{Blueprint: bp, MountRoot: "/mnt"})
	if !errors.Is(err, llm.ErrUnsafePath) {
		t.Fatalf("expected ErrUnsafePath, got %v", err)
	}
}

func TestApplyConfigDropRejectsSymlinkTarget(t *testing.T) {
	mount := t.TempDir()
	targetDir := filepath.Join(mount, "etc")
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		t.Fatalf("setup mkdir failed: %v", err)
	}
	if err := os.Symlink("/tmp/outside", filepath.Join(targetDir, "foo.conf")); err != nil {
		t.Fatalf("setup symlink failed: %v", err)
	}

	bp := &llm.Blueprint{Configs: map[string]string{"/etc/foo.conf": "x"}}
	_, err := ApplyConfigDrop(ConfigDropOptions{Blueprint: bp, MountRoot: mount, Confirm: true})
	if !errors.Is(err, llm.ErrUnsafePath) {
		t.Fatalf("expected ErrUnsafePath, got %v", err)
	}
}

func TestApplyConfigDropWritesFiles(t *testing.T) {
	dir := t.TempDir()
	mount := filepath.Join(dir, "target")
	bp := &llm.Blueprint{
		Configs: map[string]string{
			"/etc/environment.d/ai-keys.conf": "KEY=1",
		},
	}
	res, err := ApplyConfigDrop(ConfigDropOptions{Blueprint: bp, MountRoot: mount, Confirm: true})
	if err != nil {
		t.Fatalf("apply failed: %v", err)
	}
	if !res.Applied {
		t.Fatal("expected applied")
	}
	gotPath := filepath.Join(mount, "etc/environment.d/ai-keys.conf")
	data, err := os.ReadFile(gotPath)
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}
	if string(data) != "KEY=1" {
		t.Fatalf("unexpected content: %s", string(data))
	}
	info, _ := os.Stat(gotPath)
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("expected mode 0600, got %o", info.Mode().Perm())
	}
}

func TestApplyConfigDropRealRunRequiresConfirm(t *testing.T) {
	bp := &llm.Blueprint{Configs: map[string]string{"/etc/foo.conf": "x"}}
	_, err := ApplyConfigDrop(ConfigDropOptions{Blueprint: bp, MountRoot: t.TempDir()})
	if !errors.Is(err, ErrConfirmRequired) {
		t.Fatalf("expected ErrConfirmRequired, got %v", err)
	}
}
