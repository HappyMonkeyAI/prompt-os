package tui

import (
	"strings"
	"testing"

	"github.com/HappyMonkeyAI/prompt-os/internal/hardware"
)

func TestWizardHardwareAddsDiskStepAndDetectsGPU(t *testing.T) {
	m := NewWizardModelWithHardware(hardware.HardwareInfo{
		GPU:   "NVIDIA Corporation Device",
		Disks: []hardware.BlockDevice{{Path: "/dev/sdb", Size: "25G", Model: "VMware Virtual SATA Hard Drive"}},
	})

	if len(m.steps) != len(wizardSteps)+1 {
		t.Fatalf("expected disk step to be inserted, got %d steps", len(m.steps))
	}
	if m.steps[2].key != "target_disk" {
		t.Fatalf("expected target_disk after base_distro, got %q", m.steps[2].key)
	}
	if got := choiceValue(m.steps[2].choices[m.steps[2].recommended]); got != "/dev/sdb" {
		t.Fatalf("expected detected disk value /dev/sdb, got %q", got)
	}

	gpuStep := m.steps[len(m.steps)-1]
	if gpuStep.key != "gpu" || gpuStep.recommended != 0 {
		t.Fatalf("expected nvidia GPU recommendation, got step=%q index=%d", gpuStep.key, gpuStep.recommended)
	}
	if !strings.Contains(gpuStep.choices[0], "detected") {
		t.Fatalf("expected detected annotation on GPU choice: %#v", gpuStep.choices)
	}
}

func TestGPUChoiceIndexDetectsVirtual(t *testing.T) {
	if got := gpuChoiceIndex("VMware SVGA II Adapter"); got != 3 {
		t.Fatalf("expected virtual GPU choice, got %d", got)
	}
}
