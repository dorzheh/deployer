package xen_xl

import (
	"fmt"
	"testing"
)

func TestVersion(t *testing.T) {
	d := NewDriver(nil)
	v, err := d.Version()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("DEBUG: driver version => %s\n", v)
}
