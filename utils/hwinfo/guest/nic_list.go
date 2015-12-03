package guest

import (
	"errors"
	"fmt"
	"sort"

	"github.com/dorzheh/deployer/utils"
	"github.com/dorzheh/deployer/utils/hwinfo/host"
)

type NICList []*NIC

func NewNICList() NICList {
	return make(NICList, 0)
}

func (list *NICList) Add(n *NIC) {
	*list = append(*list, n)
}

func (list *NICList) AppendList(l NICList) {
	*list = append(*list, l[:]...)
}

func (list *NICList) Remove(index int) {
	temp := *list
	temp = append(temp[:index], temp[index+1:]...)
	*list = temp
}

func (list NICList) IndexByPCISlot(pcislot string) (int, error) {
	list.SortByPCISlot()
	f := func(i int) bool {
		return list[i].PCIAddr.Slot >= pcislot
	}
	if i := sort.Search(len(list), f); list[i].PCIAddr.Slot == pcislot {
		return i, nil
	}
	return -1, utils.FormatError(fmt.Errorf("index for PCI slot %s not found", pcislot))
}

func (list NICList) NicByHostNicObj(n *host.NIC) (*NIC, int, error) {
	for i, nic := range list {
		if nic.HostNIC == n {
			return nic, i, nil
		}
	}
	return nil, -1, utils.FormatError(errors.New("Host NIC object not found"))
}

func (list NICList) Length() int {
	return len(list)
}

func (list NICList) SortByPCISlot() {
	sort.Sort(SortByPCISlot(list))
}

func (list NICList) NicsByHostNicVendor(vendor string) NICList {
	var newList NICList

	for _, nic := range list {
		if nic.HostNIC.Vendor == vendor {
			if newList == nil {
				newList = NewNICList()
			}
			newList.Add(nic)
		}
	}
	return newList
}

func (list NICList) NicsByNUMAId(cellID int) NICList {
	var newList NICList

	for _, nic := range list {
		if nic.HostNIC.NUMANode == cellID {
			if newList == nil {
				newList = NewNICList()
			}
			newList.Add(nic)
		}
	}
	return newList
}

type SortByPCISlot NICList

func (list SortByPCISlot) Len() int           { return len(list) }
func (list SortByPCISlot) Swap(i, j int)      { list[i], list[j] = list[j], list[i] }
func (list SortByPCISlot) Less(i, j int) bool { return list[i].PCIAddr.Slot < list[j].PCIAddr.Slot }
