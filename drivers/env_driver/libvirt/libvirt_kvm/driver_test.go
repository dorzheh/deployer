package libvirt_kvm

import (
	"fmt"
	"testing"
)

func TestEmulator(t *testing.T) {
	d := NewDriver(nil)
	emu, err := d.Emulator("x86_64")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("DEBUG: x86_64 emulator => %s\n", emu)

	emu, err = d.Emulator("i686")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("DEBUG: i686 emulator => %s\n", emu)
}

func TestVersion(t *testing.T) {
	d := NewDriver(nil)
	v, err := d.Version()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("DEBUG: driver version => %s\n", v)
}

func TestAllCPUsPinned(t *testing.T) {
	d := NewDriver(nil)
	pinned, err := d.AllCPUsPinned()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("AllCPUsPinned result : %v\n", pinned)
}
