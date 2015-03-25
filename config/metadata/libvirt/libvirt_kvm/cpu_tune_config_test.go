package libvirt_kvm

import (
	"fmt"
	"testing"
)

var numas = map[int][]string{
	0: {"0", "1", "2", "3"},
}

func SetCpuTuneDataTest(t *testing.T) {
	fmt.Printf(SetCpuTuneData(numas))
}
