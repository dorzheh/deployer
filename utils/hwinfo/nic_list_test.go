package hwinfo

import (
	"testing"
)

func createTestList() NICList {
	phys1 := new(NIC)
	phys1.Desc = "test NIC #1"
	phys1.Driver = "e1000e"
	phys1.Name = "eth2"
	phys1.PCIAddr = "0000:00:16.1"
	phys1.Type = NicTypePhys
	phys1.Vendor = "Intel Corporation"
	phys1.Model = "82579LM Gigabit Network Connection"

	phys2 := new(NIC)
	phys2.Desc = "test NIC #2"
	phys2.Driver = "e1000e"
	phys2.Name = "eth1"
	phys2.PCIAddr = "0000:00:16.2"
	phys2.Type = NicTypePhys
	phys2.Vendor = "Intel Corporation"
	phys2.Model = "82579LM Gigabit Network Connection"

	phys3 := new(NIC)
	phys3.Desc = "test NIC #3"
	phys3.Driver = "e1000e"
	phys3.Name = "eth0"
	phys3.PCIAddr = "0000:00:16.3"
	phys3.Type = NicTypePhys
	phys3.Vendor = "Intel Corporation"
	phys3.Model = "82579LM Gigabit Network Connection"

	list := NewNICList()
	list.Add(phys3)
	list.Add(phys1)
	list.Add(phys2)
	return list
}

func TestAdd(t *testing.T) {
	l := createTestList()
	phys4 := new(NIC)
	phys4.Desc = "test NIC #4"
	phys4.Driver = "e1000e"
	phys4.Name = "eth3"
	phys4.PCIAddr = "0000:00:16.4"
	phys4.Type = NicTypePhys
	phys4.Vendor = "Intel Corporation"
	phys4.Model = "82579LM Gigabit Network Connection"
	l.Add(phys4)
	if len(l) != 4 {
		t.Fatalf("expected 4 elements, got %d", len(l))
	}
}

func TestSortByPCI(t *testing.T) {
	l := createTestList()
	l.SortByPCI()
	if l[0].PCIAddr != "0000:00:16.1" {
		t.Fatalf("expected 0000:00:16.1, got %s", l[0].PCIAddr)
	}
	if l[1].PCIAddr != "0000:00:16.2" {
		t.Fatalf("expected 0000:00:16.2, got %s", l[1].PCIAddr)
	}
	if l[2].PCIAddr != "0000:00:16.3" {
		t.Fatalf("expected 0000:00:16.3, got %s", l[2].PCIAddr)
	}
}

func TestSortByName(t *testing.T) {
	l := createTestList()
	l.SortByName()
	if l[0].Name != "eth0" {
		t.Fatalf("expected eth0 got %s", l[0].Name)
	}
	if l[1].Name != "eth1" {
		t.Fatalf("expected eth1 got %s", l[1].Name)
	}
	if l[2].Name != "eth2" {
		t.Fatalf("expected eth2 got %s", l[2].Name)
	}
}

func TestRemove(t *testing.T) {
	l := createTestList()
	i, err := l.SearchIndexByPCI("0000:00:16.2")
	if err != nil {
		t.Fatal(err)
	}
	l.Remove(i)
	i, err = l.SearchIndexByPCI("0000:00:16.2")
	if err == nil {
		t.Fatal("NIC with PCIAddr 0000:00:16.2 still exists")
	}
}
