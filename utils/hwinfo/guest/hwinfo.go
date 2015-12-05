package guest

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/dorzheh/deployer/builder/image"
	"github.com/dorzheh/deployer/config/xmlinput"
	"github.com/dorzheh/deployer/utils"
	"github.com/dorzheh/deployer/utils/hwinfo/host"
)

type NUMA struct {
	CellID   int
	MemoryMb int
	VCPUs    []int
	CPUPin   map[int][]int
	NICs     NICList
}

type Config struct {
	CPUs        int
	RamMb       int
	Storage     *image.Config
	ConfigNUMAs bool
	NUMAs       []*NUMA
	NICLists    []NICList
	Networks    []*xmlinput.Network
}

func NewConfig() *Config {
	c := new(Config)
	c.CPUs = 0
	c.RamMb = 0
	c.Storage = nil
	c.ConfigNUMAs = true
	c.NUMAs = make([]*NUMA, 0)
	c.NICLists = make([]NICList, 0)
	c.Networks = make([]*xmlinput.Network, 0)
	return c
}

func (c *Config) SetTopologySingleVirtualNUMA(numas host.NUMANodes, pinning bool) error {
	if c.CPUs == 0 {
		return errors.New("A data memeber (CPUs) is not initialized")
	}
	if c.RamMb == 0 {
		return errors.New("A data memeber (RamMb) is not initialized")
	}

	gn := new(NUMA)
	gn.CellID = 0
	gn.MemoryMb = c.RamMb
	gn.CPUPin = make(map[int][]int, 0)
	nicNumaMapping, _ := numaNicsMapping(c, numas)

	for i := 0; i < c.CPUs; i++ {
		gn.VCPUs = append(gn.VCPUs, i)
		for _, n := range numas {
			l, ok := nicNumaMapping[n.CellID]
			if ok {
				gn.NICs.AppendList(l)
			}
			if pinning {
				switch {
				case c.CPUs <= len(n.CPUs):
					gn.CPUPin[i] = append(gn.CPUPin[i], n.CPUs[i])
					break
				default:
					break
				}
			} else {
				gn.CPUPin[i] = append(gn.CPUPin[i], n.CPUs...)
			}
		}
	}

	c.NUMAs = append(c.NUMAs, gn)
	return nil
}

func (c *Config) SetTopologyMultipleVirtualNUMAs(numas host.NUMANodes) error {
	if c.CPUs == 0 {
		return errors.New("A data memeber (CPUs) is not initialized")
	}
	if c.RamMb == 0 {
		return errors.New("A data memeber (RamMb) is not initialized")
	}
	if len(c.NICLists) == 0 {
		return errors.New("A data memeber (NICLists) is not initialized")
	}

	nicNumaMapping, totalAmountOfPorts := numaNicsMapping(c, numas)

	cellID := 0
	vcpuID := 0
	for _, n := range numas {
		if portsList, ok := nicNumaMapping[n.CellID]; ok {
			percentage, err := utils.FloatStringsSliceToInt(strings.Split(fmt.Sprintf("%0.1f", float32(len(portsList))*float32(100)/float32(totalAmountOfPorts)), "."))
			if err != nil {
				return nil
			}

			amountOfMemoryMb, err := utils.FloatStringsSliceToInt(strings.Split(fmt.Sprintf("%0.1f", float32(c.RamMb)/float32(100)*float32(percentage)), "."))
			if err != nil {
				return err
			}

			amountOfCpus, err := utils.FloatStringsSliceToInt(strings.Split(fmt.Sprintf("%0.1f", float32(c.CPUs)/float32(100)*float32(percentage)), "."))
			if err != nil {
				return err
			}

			// fmt.Printf("total amount of ports = %d\n", totalAmountOfPorts)
			// fmt.Printf("amount of ports on NUMA = %d\n", len(portsList))
			// fmt.Printf("percentage = %d\n", percentage)
			// fmt.Printf("amount of MemoryMb = %d\n", amountOfMemoryMb)
			// fmt.Printf("amount of CPUS = %d\n", amountOfCpus)
			// fmt.Printf("required CPUS = %d\n", c.CPUs)

			switch {
			case amountOfCpus > len(n.CPUs) && amountOfMemoryMb > n.TotalRAM:
				c.ConfigNUMAs = false
				return c.SetTopologySingleVirtualNUMA(numas, false)
			case amountOfCpus > len(n.CPUs) && amountOfMemoryMb <= n.TotalRAM:
				return c.SetTopologySingleVirtualNUMA(numas, false)
			case amountOfMemoryMb > n.TotalRAM && amountOfCpus <= len(n.CPUs):
				c.ConfigNUMAs = false
				return c.SetTopologySingleVirtualNUMA(numas, true)
			}

			gn := new(NUMA)
			gn.CellID = cellID
			gn.CPUPin = make(map[int][]int, 0)
			gn.NICs = portsList

			sort.Ints(n.CPUs)
			vcpusPerNUMA := len(n.CPUs[0:amountOfCpus])
			gn.MemoryMb = amountOfMemoryMb

			for x := 0; x < vcpusPerNUMA; x++ {
				gn.CPUPin[vcpuID] = append(gn.CPUPin[vcpuID], n.CPUs[x])
				gn.VCPUs = append(gn.VCPUs, vcpuID)
				vcpuID++
			}

			c.NUMAs = append(c.NUMAs, gn)
			cellID++
		}
	}
	return nil
}

func numaNicsMapping(c *Config, numas host.NUMANodes) (map[int][]*NIC, int) {
	nicNumaMapping := make(map[int][]*NIC)
	totalAmountOfPorts := 0

	for _, l := range c.NICLists {
		for _, n := range numas {
			cellID := 0
			if n.CellID != host.NoNUMA {
				cellID = n.CellID
			}

			pList := l.NicsByNUMAId(cellID)
			if pList.Length() > 0 {
				totalAmountOfPorts += pList.Length()
				nicNumaMapping[cellID] = append(nicNumaMapping[cellID], pList...)
			}
		}
	}

	return nicNumaMapping, totalAmountOfPorts
}
