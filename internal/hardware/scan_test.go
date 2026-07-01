package hardware

import "testing"

func TestParseBlockDevicesHandlesQuotedModels(t *testing.T) {
	out := "NAME=\"sda\" SIZE=\"25G\" MODEL=\"VMware Virtual SATA Hard Drive\" TYPE=\"disk\" RM=\"0\" RO=\"0\" TRAN=\"sata\"\n" +
		"NAME=\"sr0\" SIZE=\"1G\" MODEL=\"DVD\" TYPE=\"rom\" RM=\"1\" RO=\"0\" TRAN=\"sata\"\n"

	devices := parseBlockDevices(out)
	if len(devices) != 1 {
		t.Fatalf("expected one disk, got %d", len(devices))
	}
	got := devices[0]
	if got.Path != "/dev/sda" || got.Size != "25G" || got.Model != "VMware Virtual SATA Hard Drive" || got.Transport != "sata" {
		t.Fatalf("unexpected disk parsed: %+v", got)
	}
	if got.Removable || got.ReadOnly {
		t.Fatalf("expected writable non-removable disk: %+v", got)
	}
}
