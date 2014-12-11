package libvirt

import (
	"fmt"
	"testing"
)

const libvirtVersion = "1.3.4"

func TestMultiQueueSupported(t *testing.T) {
	yesno, err := MultiQueueSupportedMock(multiQueueKernelVersion, libvirtVersion, nil)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("Supported: %v\n", yesno)
}
