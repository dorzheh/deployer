package guest

import (
	"fmt"
	"testing"

	"github.com/dorzheh/deployer/utils/hwinfo/host"
)

func createTestListSingleNUMA() []NICList {
	phys1n1 := new(host.NIC)
	phys1n1.Desc = "test NIC #1"
	phys1n1.Driver = "e1000e"
	phys1n1.Name = "eth0"
	phys1n1.PCIAddr = "0000:00:16.1"
	phys1n1.NUMANode = 0
	phys1n1.Type = host.NicTypePhys
	phys1n1.Vendor = "Intel Corporation"
	phys1n1.Model = "82579LM Gigabit Network Connection"

	phys2n1 := new(host.NIC)
	phys2n1.Desc = "test NIC #2"
	phys2n1.Driver = "e1000e"
	phys2n1.Name = "eth1"
	phys2n1.PCIAddr = "0000:00:16.2"
	phys2n1.NUMANode = 0
	phys2n1.Type = host.NicTypePhys
	phys2n1.Vendor = "Intel Corporation"
	phys2n1.Model = "82579LM Gigabit Network Connection"

	phys3n1 := new(host.NIC)
	phys3n1.Desc = "test NIC #3"
	phys3n1.Driver = "e1000e"
	phys3n1.Name = "eth2"
	phys3n1.PCIAddr = "0000:00:16.3"
	phys3n1.NUMANode = 0
	phys3n1.Type = host.NicTypePhys
	phys3n1.Vendor = "Intel Corporation"
	phys3n1.Model = "82579LM Gigabit Network Connection"

	phys1n2 := new(host.NIC)
	phys1n2.Desc = "test NIC #4"
	phys1n2.Driver = "e1000e"
	phys1n2.Name = "eth3"
	phys1n2.PCIAddr = "0000:00:17.1"
	phys1n2.NUMANode = 0
	phys1n2.Type = host.NicTypePhys
	phys1n2.Vendor = "Intel Corporation"
	phys1n2.Model = "82579LM Gigabit Network Connection"

	phys2n2 := new(host.NIC)
	phys2n2.Desc = "test NIC #5"
	phys2n2.Driver = "e1000e"
	phys2n2.Name = "eth4"
	phys2n2.PCIAddr = "0000:00:17.2"
	phys2n2.NUMANode = 0
	phys2n2.Type = host.NicTypePhys
	phys2n2.Vendor = "Intel Corporation"
	phys2n2.Model = "82579LM Gigabit Network Connection"

	phys3n2 := new(host.NIC)
	phys3n2.Desc = "test NIC #6"
	phys3n2.Driver = "e1000e"
	phys3n2.Name = "eth5"
	phys3n2.PCIAddr = "0000:00:17.3"
	phys3n2.NUMANode = 0
	phys3n2.Type = host.NicTypePhys
	phys3n2.Vendor = "Intel Corporation"
	phys3n2.Model = "82579LM Gigabit Network Connection"

	gn1 := NewNIC()
	gn1.HostNIC = phys1n1
	gn1.Network = "net1"
	gn1.PCIAddr = &PCI{"0000", "00", "01", "0"}

	gn2 := NewNIC()
	gn2.HostNIC = phys2n1
	gn2.Network = "net1"
	gn2.PCIAddr = &PCI{"0000", "00", "01", "1"}

	gn3 := NewNIC()
	gn3.HostNIC = phys3n1
	gn3.Network = "net1"
	gn3.PCIAddr = &PCI{"0000", "00", "02", "0"}

	gn4 := NewNIC()
	gn4.HostNIC = phys1n2
	gn4.Network = "net2"
	gn4.PCIAddr = &PCI{"0000", "00", "01", "0"}

	gn5 := NewNIC()
	gn5.HostNIC = phys2n2
	gn5.Network = "net2"
	gn5.PCIAddr = &PCI{"0000", "00", "01", "1"}

	gn6 := NewNIC()
	gn6.HostNIC = phys3n2
	gn6.Network = "net1"
	gn6.PCIAddr = &PCI{"0000", "00", "02", "0"}

	list1 := NewNICList()
	list1.Add(gn1)
	list1.Add(gn2)
	list1.Add(gn3)

	list2 := NewNICList()
	list2.Add(gn4)
	list2.Add(gn5)
	list2.Add(gn6)
	return []NICList{list1, list2}
}

func createTestListMultipleNUMAs() []NICList {
	phys1n1 := new(host.NIC)
	phys1n1.Desc = "test NIC #1"
	phys1n1.Driver = "e1000e"
	phys1n1.Name = "eth0"
	phys1n1.PCIAddr = "0000:00:16.1"
	phys1n1.NUMANode = 0
	phys1n1.Type = host.NicTypePhys
	phys1n1.Vendor = "Intel Corporation"
	phys1n1.Model = "82579LM Gigabit Network Connection"

	phys2n1 := new(host.NIC)
	phys2n1.Desc = "test NIC #2"
	phys2n1.Driver = "e1000e"
	phys2n1.Name = "eth1"
	phys2n1.PCIAddr = "0000:00:16.2"
	phys2n1.NUMANode = 0
	phys2n1.Type = host.NicTypePhys
	phys2n1.Vendor = "Intel Corporation"
	phys2n1.Model = "82579LM Gigabit Network Connection"

	phys3n1 := new(host.NIC)
	phys3n1.Desc = "test NIC #3"
	phys3n1.Driver = "e1000e"
	phys3n1.Name = "eth2"
	phys3n1.PCIAddr = "0000:00:16.3"
	phys3n1.NUMANode = 0
	phys3n1.Type = host.NicTypePhys
	phys3n1.Vendor = "Intel Corporation"
	phys3n1.Model = "82579LM Gigabit Network Connection"

	phys1n2 := new(host.NIC)
	phys1n2.Desc = "test NIC #4"
	phys1n2.Driver = "e1000e"
	phys1n2.Name = "eth3"
	phys1n2.PCIAddr = "0000:00:17.1"
	phys1n2.NUMANode = 0
	phys1n2.Type = host.NicTypePhys
	phys1n2.Vendor = "Intel Corporation"
	phys1n2.Model = "82579LM Gigabit Network Connection"

	phys2n2 := new(host.NIC)
	phys2n2.Desc = "test NIC #5"
	phys2n2.Driver = "e1000e"
	phys2n2.Name = "eth4"
	phys2n2.PCIAddr = "0000:00:17.2"
	phys2n2.NUMANode = 1
	phys2n2.Type = host.NicTypePhys
	phys2n2.Vendor = "Intel Corporation"
	phys2n2.Model = "82579LM Gigabit Network Connection"

	phys3n2 := new(host.NIC)
	phys3n2.Desc = "test NIC #6"
	phys3n2.Driver = "e1000e"
	phys3n2.Name = "eth5"
	phys3n2.PCIAddr = "0000:00:17.3"
	phys3n2.NUMANode = 1
	phys3n2.Type = host.NicTypePhys
	phys3n2.Vendor = "Intel Corporation"
	phys3n2.Model = "82579LM Gigabit Network Connection"

	gn1 := NewNIC()
	gn1.HostNIC = phys1n1
	gn1.Network = "net1"
	gn1.PCIAddr = &PCI{"0000", "00", "01", "0"}

	gn2 := NewNIC()
	gn2.HostNIC = phys2n1
	gn2.Network = "net1"
	gn2.PCIAddr = &PCI{"0000", "00", "01", "1"}

	gn3 := NewNIC()
	gn3.HostNIC = phys3n1
	gn3.Network = "net1"
	gn3.PCIAddr = &PCI{"0000", "00", "02", "0"}

	gn4 := NewNIC()
	gn4.HostNIC = phys1n2
	gn4.Network = "net2"
	gn4.PCIAddr = &PCI{"0000", "00", "01", "0"}

	gn5 := NewNIC()
	gn5.HostNIC = phys2n2
	gn5.Network = "net2"
	gn5.PCIAddr = &PCI{"0000", "00", "01", "1"}

	gn6 := NewNIC()
	gn6.HostNIC = phys3n2
	gn6.Network = "net1"
	gn6.PCIAddr = &PCI{"0000", "00", "02", "0"}

	list1 := NewNICList()
	list1.Add(gn1)
	list1.Add(gn2)
	list1.Add(gn3)

	list2 := NewNICList()
	list2.Add(gn4)
	list2.Add(gn5)
	list2.Add(gn6)
	return []NICList{list1, list2}
	return []NICList{list1}
}

func TestSetTopologySingleNUMA(t *testing.T) {
	hn := new(host.NUMA)
	hn.CPUs = []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	hn.CellID = 0
	hn.FreeRAM = 6000
	hn.TotalRAM = 6000

	numas := host.NUMANodes{hn}

	c := NewConfig()
	c.CPUs = 6
	c.RamMb = 5
	if err := c.SetTopologySingleVirtualNUMA(numas, false); err != nil {
		t.Fatal(err)
	}
	for _, gn := range c.NUMAs {
		fmt.Println("\n===== Guest NUMA info =====")
		fmt.Printf("CellID: %v\n", gn.CellID)
		fmt.Printf("MemoryMb: %v\n", gn.MemoryMb)
		fmt.Printf("vCPUs: %v\n", gn.VCPUs)
	}
}

func TestSetTopologySingleVirtualNUMA(t *testing.T) {
	hn := new(host.NUMA)
	hn.CPUs = []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	hn.CellID = 0
	hn.FreeRAM = 6000
	hn.TotalRAM = 6000

	c := NewConfig()
	c.CPUs = 6
	c.RamMb = 5
	if err := c.SetTopologySingleVirtualNUMA(host.NUMANodes{hn}, true); err != nil {
		t.Fatal(err)
	}
	for _, gn := range c.NUMAs {
		fmt.Println("\n===== Guest NUMA info =====")
		fmt.Printf("CellID: %v\n", gn.CellID)
		fmt.Printf("MemoryMb: %v\n", gn.MemoryMb)
		fmt.Printf("vCPUs: %v\n", gn.VCPUs)
	}
}

func TestSetTopologyOneToOnePinningMultipleNUMA(t *testing.T) {

	hn1 := new(host.NUMA)
	hn1.CPUs = []int{0, 2, 4, 6, 8, 10, 12, 14, 16, 18}
	hn1.CellID = 0
	hn1.FreeRAM = 6000
	hn1.TotalRAM = 64375

	hn2 := new(host.NUMA)
	hn2.CPUs = []int{1, 3, 5, 7, 9, 11, 13, 15, 17, 19}
	hn2.CellID = 1
	hn2.FreeRAM = 6000
	hn2.TotalRAM = 64375

	hn3 := new(host.NUMA)
	hn3.CPUs = []int{20, 22, 24, 26, 28, 30, 32, 34, 36, 38}
	hn3.CellID = 2
	hn3.FreeRAM = 6000
	hn3.TotalRAM = 6437

	hn4 := new(host.NUMA)
	hn4.CPUs = []int{21, 23, 25, 27, 29, 31, 33, 35, 37, 39}
	hn4.CellID = 3
	hn4.FreeRAM = 6000
	hn4.TotalRAM = 6437

	numas := host.NUMANodes{hn1, hn2, hn3, hn4}

	c := NewConfig()
	c.CPUs = 10
	c.RamMb = 6000
	c.NICLists = createTestListMultipleNUMAs()
	if err := c.SetTopologyMultipleVirtualNUMAs(numas); err != nil {
		t.Fatal(err)
	}

	fmt.Println("\nTestSetTopologyMultipleVirtualNUMAs")
	for _, gn := range c.NUMAs {
		fmt.Println("\n===== Guest NUMA info =====")
		fmt.Printf("CellID: %v\n", gn.CellID)
		fmt.Printf("MemoryMb: %v\n", gn.MemoryMb)
		fmt.Printf("vCPUs: %v\n", gn.VCPUs)
		fmt.Printf("CPU pinning %v\n", gn.CPUPin)
	}
}
