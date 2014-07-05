package libvirt

import (
	"fmt"
	"testing"
)

func TestEmulator(t *testing.T) {
	d := NewDriver(nil)
	emu, err := d.Emulator()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("DEBUG: emulator => %s\n", emu)
}
