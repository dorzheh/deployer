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
	CPUPin   map[int][]int
	NICs     NICList
}

type Config struct {
	CPUs                      int
	RamMb                     int
	Storage                   *image.Config
	LargeHugePagesSupported   bool
	NUMAs                     []*NUMA
	HostNUMAIds               []int
	NICLists                  []NICList
	Networks                  []*xmlinput.Network
	OptimizationFailureMemory bool
	OptimizationFailureCPU    bool
	OptimizationFailureMsg    string
}

func NewConfig() *Config {
	c := new(Config)
	c.CPUs = 0
	c.RamMb = 0
	c.Storage = nil
	c.LargeHugePagesSupported = false
	c.NUMAs = make([]*NUMA, 0)
	c.HostNUMAIds = make([]int, 0)
	c.NICLists = make([]NICList, 0)
	c.Networks = make([]*xmlinput.Network, 0)
	c.OptimizationFailureMemory = false
	c.OptimizationFailureCPU = false
	return c
}

func (c *Config) SetTopologySingleVirtualNUMA(numas host.NUMANodes, singleNuma bool) error {
	if c.CPUs == 0 {
		return errors.New("A data memeber (CPUs) is not initialized")
	}
	if c.RamMb == 0 {
		return errors.New("A data memeber (RamMb) is not initialized")
	}

	c.NUMAs = nil
	gn := new(NUMA)
	gn.CellID = 0
	gn.MemoryMb = c.RamMb
	gn.NICs = nil
	gn.CPUPin = make(map[int][]int, 0)
	nicNumaMapping, _ := numaNicsMapping(c, numas)

	for _, n := range numas {
		l, ok := nicNumaMapping[n.CellID]
		if ok {
			gn.NICs.AppendList(l)
		}
	}

	if singleNuma {
		cpusNuma0, err := numas.CpusOnNUMA(0)
		if err != nil {
			return err
		}
		if c.CPUs <= len(cpusNuma0) {
			for i := 0; i < c.CPUs; i++ {
				gn.CPUPin[i] = append(gn.CPUPin[i], i)
			}
			c.HostNUMAIds = []int{0}
		}
	} else {
		for i := 0; i < c.CPUs; i++ {
			for _, n := range numas {
				gn.CPUPin[i] = append(gn.CPUPin[i], n.CPUs...)
				c.HostNUMAIds = append(c.HostNUMAIds, n.CellID)
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

	if c.CPUs > numas.TotalCPUs() {
		c.OptimizationFailureCPU = true
		c.OptimizationFailureMemory = true
		c.OptimizationFailureMsg = fmt.Sprintf("Amount of requested CPUs (%d) is greater than installed(%d).", c.CPUs, numas.TotalCPUs())
		return c.SetTopologySingleVirtualNUMA(numas, false)
	}

	nicNumaMapping, totalAmountOfPorts := numaNicsMapping(c, numas)
	if totalAmountOfPorts == 0 {
		return c.SetTopologySingleVirtualNUMA(numas, true)
	}

	cellID := 0
	vcpuID := 0
	allocatedCpus := 0
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
			// fmt.Printf("amount of ports on NUMA %d = %d\n", n.CellID, len(portsList))
			// fmt.Printf("percentage = %d\n", percentage)
			// fmt.Printf("amount of MemoryMb = %d\n", amountOfMemoryMb)
			// fmt.Printf("amount of CPUS = %d\n", amountOfCpus)
			// fmt.Printf("required CPUS = %d\n", c.CPUs)

			cpusOnNuma := len(n.CPUs)
			switch {
			case amountOfCpus > cpusOnNuma && amountOfMemoryMb > n.TotalRAM:
				c.OptimizationFailureCPU = true
				c.OptimizationFailureMemory = true
				c.OptimizationFailureMsg = fmt.Sprintf("Not enough CPUs and memory on NUMA %d (requested CPUs %d, installed CPUs %d\nrequested memory %dMB , installed %dMB).", n.CellID, amountOfCpus, cpusOnNuma, amountOfMemoryMb, n.TotalRAM)
				return c.SetTopologySingleVirtualNUMA(numas, false)
			case amountOfCpus > cpusOnNuma && amountOfMemoryMb <= n.TotalRAM:
				c.OptimizationFailureCPU = true
				c.OptimizationFailureMsg = fmt.Sprintf("Not enough CPUs on NUMA %d (requested %d , installed %d).", n.CellID, amountOfCpus, cpusOnNuma)
				return c.SetTopologySingleVirtualNUMA(numas, false)
			case amountOfMemoryMb > n.TotalRAM && amountOfCpus <= cpusOnNuma:
				c.OptimizationFailureMemory = true
				c.OptimizationFailureMsg = fmt.Sprintf("Not enough memory on NUMA %d (requested memory %dMB , installed %dMB).", n.CellID, amountOfMemoryMb, n.TotalRAM)
				return c.SetTopologySingleVirtualNUMA(numas, true)
			case amountOfCpus == 0:
				c.OptimizationFailureCPU = true
				c.OptimizationFailureMsg = "The Virtual Machine is configured with a single vCPU."
				if amountOfMemoryMb > n.TotalRAM {
					c.OptimizationFailureMemory = true
				}
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
				vcpuID++
			}

			allocatedCpus += amountOfCpus
			c.NUMAs = append(c.NUMAs, gn)
			c.HostNUMAIds = append(c.HostNUMAIds, n.CellID)
			cellID++
		}
	}
	if allocatedCpus != c.CPUs {
		c.OptimizationFailureCPU = true
		c.OptimizationFailureMsg = "Cannot distribute vCPUs among the vNUMAS."
		return c.SetTopologySingleVirtualNUMA(numas, false)
	}
	return nil
}

func numaNicsMapping(c *Config, numas host.NUMANodes) (map[int][]*NIC, int) {
	nicNumaMapping := make(map[int][]*NIC)
	totalAmountOfPorts := 0

	for _, l := range c.NICLists {
		for _, n := range numas {
			pList := l.NicsByNUMAId(n.CellID)
			if pList.Length() > 0 {
				totalAmountOfPorts += pList.Length()
				nicNumaMapping[n.CellID] = append(nicNumaMapping[n.CellID], pList...)
			}
		}
	}
	return nicNumaMapping, totalAmountOfPorts
}
