// Responsible for parsing lshw output

package host

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/clbanning/mxj"
	"github.com/dorzheh/deployer/utils"
	ssh "github.com/dorzheh/infra/comm/common"
)

type Collector struct {
	Run        func(string) (string, error)
	prepare    func() (string, error)
	hwinfoFile string
}

// NewCollector constructs new lshw Collector
// The output will be represented in JSON format
func NewCollector(sshconf *ssh.Config, lshwpath, hwinfoFile string) (*Collector, error) {
	c := new(Collector)
	c.Run = utils.RunFunc(sshconf)
	c.hwinfoFile = hwinfoFile
	c.prepare = prepareFunc(c, sshconf, lshwpath)
	return c, nil
}

// Parse parses lshw output
func (c *Collector) Hwinfo2Json() error {
	lshwNewPath, err := c.prepare()
	if err != nil {
		return utils.FormatError(err)
	}
	out, err := c.Run(lshwNewPath + " -class network -class cpu -json")
	if err != nil {
		return utils.FormatError(err)
	}
	return ioutil.WriteFile(c.hwinfoFile, []byte(out), 0)
}

// CPU contains CPU description and properties
type CPU struct {
	Type     string
	Capacity float64
	Clock    float64
	Config   map[string]interface{}
	Cap      map[string]interface{}
}

// CPUInfo gathers information related to installed CPUs
func (c *Collector) CPUInfo() (*CPU, error) {
	if _, err := os.Stat(c.hwinfoFile); err != nil {
		if err = c.Hwinfo2Json(); err != nil {
			return nil, utils.FormatError(err)
		}
	}

	out, err := mxj.NewMapsFromJsonFile(c.hwinfoFile)
	if err != nil {
		return nil, utils.FormatError(err)
	}

	cpu := new(CPU)
	cpu.Config = make(map[string]interface{})
	cpu.Cap = make(map[string]interface{})
	for _, m := range out {
		if val, err := m.ValueForPathString("id"); err == nil {
			if val == "cpu" || val == "cpu:0" {
				if val, err := m.ValueForPathString("product"); err == nil {
					cpu.Type = val
				}
				if val, err := m.ValueForPath("capacity"); err == nil {
					cpu.Capacity = val.(float64)
				}
				if val, err := m.ValueForPath("clock"); err == nil {
					cpu.Clock = val.(float64)
				}

				conf, err := m.ValueForPath("configuration")
				if err == nil {
					for k, v := range conf.(map[string]interface{}) {
						cpu.Config[k] = v
					}
				}
				caps, err := m.ValueForPath("capabilities")
				if err == nil {
					for k, v := range caps.(map[string]interface{}) {
						cpu.Cap[k] = v
					}
				}
			}
		}
	}
	return cpu, nil
}

func (c *Collector) CPUs() (int, error) {
	cpustr, err := c.Run(`grep -c ^processor /proc/cpuinfo`)
	if err != nil {
		return 0, utils.FormatError(err)
	}

	cpus, err := strconv.Atoi(strings.Trim(cpustr, "\n"))
	if err != nil {
		return 0, utils.FormatError(err)
	}
	return int(cpus), nil
}

// supported NIC types
type NicType string

const (
	NicTypePhys           NicType = "physical"
	NicTypePhysVF         NicType = "virtualfunc"
	NicTypeOVS            NicType = "openvswitch"
	NicTypeBridge         NicType = "bridge"
	NicTypeVirtualNetwork NicType = "virtnetwork"
)

const NoNUMA = -1

// NIC information
type NIC struct {
	// port name (eth0,br0...)
	Name string

	// NIC driver(bridge,openvswitch...)
	Driver string

	// Vendor
	Vendor string

	// Model
	Model string

	// PCI Address
	PCIAddr string

	// NUMA Node
	NUMANode int

	// Description
	Desc string

	// Port type
	Type NicType
}

// NICInfo gathers information related to installed NICs
func (c *Collector) NICInfo() (NICList, error) {
	if _, err := os.Stat(c.hwinfoFile); err != nil {
		if err = c.Hwinfo2Json(); err != nil {
			return nil, utils.FormatError(err)
		}
	}
	out, err := mxj.NewMapsFromJsonFile(c.hwinfoFile)
	if err != nil {
		return nil, utils.FormatError(err)
	}

	list := NewNICList()
	for _, m := range out {
		if val, err := m.ValueForPathString("id"); err == nil && strings.HasPrefix(val, "network") {
			if desc, err := m.ValueForPathString("description"); err == nil &&
				(desc == "Ethernet interface" || desc == "Ethernet controller" ||
					desc == "Network controller" || desc == "interface") {
				var name string
				name, err := m.ValueForPathString("logicalname")
				if err != nil {
					name = "N/A"
				} else {
					if name == "ovs-system" {
						continue
					}
				}

				nic := new(NIC)
				nic.Name = name
				nic.PCIAddr = "N/A"
				nic.NUMANode = NoNUMA

				conf, err := m.ValueForPath("configuration")
				if err != nil {
					nic.Driver = "N/A"
				} else {
					if val, ok := conf.(map[string]interface{})["driver"]; ok {
						nic.Driver = val.(string)
					}
				}
				switch nic.Driver {
				case "tun", "veth", "macvlan":
					continue
				case "openvswitch":
					nic.Desc = "Open vSwitch interface"
					nic.Type = NicTypeOVS
				default:
					if prod, err := m.ValueForPathString("product"); err == nil {
						if vendor, err := m.ValueForPathString("vendor"); err == nil {
							if _, err := c.Run(fmt.Sprintf("[[ -d /sys/class/net/%s/master || -d /sys/class/net/%s/brport ]]", nic.Name, nic.Name)); err == nil {
								// the interface is part of a bridge
								continue
							}
							if businfo, err := m.ValueForPathString("businfo"); err == nil {
								nic.PCIAddr = strings.Split(businfo, "@")[1]
							}

							nic.Model = prod
							nic.Vendor = vendor
							numa, err := c.numa4Nic(nic.PCIAddr)
							if err != nil {
								return nil, utils.FormatError(err)
							}
							nic.NUMANode = numa
							nic.Desc = vendor + " " + prod
							nic.Type = NicTypePhys

							// find Virtual Functions for appropriate physical NIC
							vfs, err := c.virtualFunctions(nic)
							if err == nil {
								list.Add(vfs[0:]...)
							}
						}
					}
				}
				list.Add(nic)
			}
		}
	}

	// lshw cannot treat linux bridges so let's do it manually
	res, err := c.Run(`out="";for n in /sys/class/net/*;do [ -d $n/bridge ] && out="$out ${n##/sys/class/net/}";done;echo $out`)
	if err != nil {
		return nil, utils.FormatError(err)
	}
	if res != "" {
		for _, n := range strings.Split(res, " ") {
			br := &NIC{
				Name:     n,
				Driver:   "bridge",
				Desc:     "Bridge interface",
				NUMANode: NoNUMA,
				PCIAddr:  "N/A",
				Type:     NicTypeBridge,
			}
			list.Add(br)
		}
	}
	list.SortByPCI()
	return list, nil
}

// numa4Nic fetchs NUMA node for appropriate PCI device
func (c *Collector) numa4Nic(pciAddr string) (int, error) {
	numaInt := 0
	numaStr, err := c.Run("cat /sys/bus/pci/devices/" + pciAddr + "/numa_node")
	if err != nil {
		return numaInt, utils.FormatError(err)
	}
	if numaStr == "-1" {
		numaInt = 0
	} else {
		numaInt, err = strconv.Atoi(numaStr)
		if err != nil {
			return numaInt, utils.FormatError(err)
		}
	}
	return numaInt, nil
}

func (c *Collector) virtualFunctions(nic *NIC) ([]*NIC, error) {
	out, err := c.Run(fmt.Sprintf("path=$(find /sys/devices  -name \"%s\" );ls -d $path/virtf* 2>/dev/null", nic.PCIAddr))
	if err != nil {
		return nil, err
	}

	var list []*NIC

	for _, line := range strings.SplitAfter(out, "\n") {
		entry, _ := c.Run("readlink -f " + strings.TrimSpace(line))
		if err != nil {
			return nil, err
		}
		s := strings.Split(entry, "/")
		vf := new(NIC)
		vf.Name = "N/A"
		vf.Vendor = nic.Vendor
		vf.Model = "Ethernet Controller Virtual Function"
		vf.Desc = nic.Vendor + " " + vf.Model
		vf.Type = NicTypePhysVF
		vf.NUMANode = nic.NUMANode
		vf.PCIAddr = s[len(s)-1]
		vf.Driver = nic.Driver
		list = append(list, vf)
	}
	return list, nil
}

// RAMSize gathers information related to the installed amount of RAM in MB
func (c *Collector) RAMSize() (int, error) {
	out, err := c.Run("grep MemTotal /proc/meminfo")
	if err != nil {
		return 0, utils.FormatError(err)
	}
	var ramsize int
	fmt.Sscanf(out, "MemTotal: %d %s", &ramsize)
	return ramsize / 1024, nil
}

type NUMA struct {
	CellID   int
	TotalRAM int
	FreeRAM  int
	CPUs     []int
}

type NUMANodes []*NUMA

func (c *Collector) NUMANodes() (NUMANodes, error) {
	out, err := c.Run("ls -d  /sys/devices/system/node/node[0-9]*")
	if err != nil {
		return nil, utils.FormatError(err)
	}

	numas := make([]*NUMA, 0)
	for i, _ := range strings.SplitAfter(out, "\n") {
		content, err := ioutil.ReadFile(fmt.Sprintf("/sys/devices/system/node/node%d/meminfo", i))
		if err != nil {
			return nil, utils.FormatError(err)
		}

		n := new(NUMA)
		n.CellID = i
		var numaNum int
		var memStatus string
		var memAmount int
		for _, line := range strings.Split(string(content), "\n") {
			fmt.Sscanf(line, "Node%d%s%dkB", &numaNum, &memStatus, &memAmount)
			if memStatus == "MemTotal:" {
				n.TotalRAM = memAmount
			} else if memStatus == "MemFree:" {
				n.FreeRAM = memAmount
			}
		}
		out, err := c.Run(fmt.Sprintf("ls -d  /sys/devices/system/node/node%d/cpu[0-9]*", i))
		if err != nil {
			return nil, utils.FormatError(err)
		}

		n.CPUs = make([]int, 0)
		for _, line := range strings.SplitAfter(out, "\n") {
			cpu := strings.TrimSpace(strings.SplitAfter(line, "cpu")[1])
			cpuInt, err := strconv.Atoi(cpu)
			if err != nil {
				return nil, utils.FormatError(err)
			}
			n.CPUs = append(n.CPUs, cpuInt)
		}
		numas = append(numas, n)
	}
	return numas, nil
}

// TotalNUMAs returns amount of NUMA nodes installed on the host
func (n NUMANodes) TotalNUMAs() int {
	return len(n)
}

// TotalCPUs returns amount of CPUs on all NUMAs
func (n NUMANodes) TotalCPUs() int {
	var cpus int
	for _, node := range n {
		cpus += len(node.CPUs)
	}
	return cpus
}

// CpusOnNUMA returns a slice of CPUs bound to the given NUMA node
func (n NUMANodes) CpusOnNUMA(cellID int) ([]int, error) {
	for _, node := range n {
		if node.CellID == cellID {
			return node.CPUs, nil
		}
	}
	return nil, fmt.Errorf("NUMA %d not found", cellID)
}

func (n NUMANodes) NUMAByCellID(cellID int) *NUMA {
	for _, node := range n {
		if node.CellID == cellID {
			return node
		}
	}
	return nil
}

func prepareFunc(c *Collector, sshconf *ssh.Config, lshwpath string) func() (string, error) {
	return func() (string, error) {
		if lshwpath == "" {
			out, err := c.Run("which lshw")
			if err != nil {
				return "", utils.FormatError(fmt.Errorf("%s [%v]", out, err))
			}
			return out, nil
		}
		if sshconf != nil {
			dir, err := utils.UploadBinaries(sshconf, lshwpath)
			if err != nil {
				return "", utils.FormatError(err)
			}
			return filepath.Join(dir, filepath.Base(lshwpath)), nil
		}
		return lshwpath, nil
	}
}
