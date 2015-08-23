package host

import (
	"errors"
	"fmt"
	"sort"

	"github.com/dorzheh/deployer/utils"
)

type NICList []*NIC

func NewNICList() NICList {
	return make(NICList, 0)
}

func (list *NICList) Add(nics ...*NIC) {
	for _, nic := range nics {
		*list = append(*list, nic)
	}
}

func (list *NICList) AppendList(l NICList) {
	*list = append(*list, l[:]...)
}

func (list *NICList) Remove(index int) {
	temp := *list
	temp = append(temp[:index], temp[index+1:]...)
	*list = temp
}

func (list NICList) SearchIndexByPCI(pciaddr string) (int, error) {
	list.SortByPCI()
	f := func(i int) bool {
		return list[i].PCIAddr >= pciaddr
	}
	if i := sort.Search(len(list), f); list[i].PCIAddr == pciaddr {
		return i, nil
	}
	return -1, utils.FormatError(fmt.Errorf("index for PCIAddr %s not found", pciaddr))
}

func (list NICList) SearchIndexByName(name string) (int, error) {
	list.SortByName()
	f := func(i int) bool {
		return list[i].Name >= name
	}
	if i := sort.Search(len(list), f); list[i].Name == name {
		return i, nil
	}
	return -1, utils.FormatError(fmt.Errorf("index for Name %s not found", name))
}

func (list NICList) SearchIndexByObj(n *NIC) (int, error) {
	for i, nic := range list {
		if nic == n {
			return i, nil
		}
	}
	return -1, utils.FormatError(errors.New("NIC object not found"))
}

func (list NICList) Length() int {
	return len(list)
}

func (list NICList) SortByPCI() {
	sort.Sort(SortByPCI(list))
}

func (list NICList) SortByName() {
	sort.Sort(SortByName(list))
}

type SortByPCI NICList

func (list SortByPCI) Len() int           { return len(list) }
func (list SortByPCI) Swap(i, j int)      { list[i], list[j] = list[j], list[i] }
func (list SortByPCI) Less(i, j int) bool { return list[i].PCIAddr < list[j].PCIAddr }

type SortByName NICList

func (list SortByName) Len() int           { return len(list) }
func (list SortByName) Swap(i, j int)      { list[i], list[j] = list[j], list[i] }
func (list SortByName) Less(i, j int) bool { return list[i].Name < list[j].Name }
