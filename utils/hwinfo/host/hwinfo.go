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
	NicTypeOVS            NicType = "openvswitch"
	NicTypeBridge         NicType = "bridge"
	NicTypeVirtualNetwork NicType = "virtnetwork"
)

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
				nic.NUMANode = -1

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
				NUMANode: -1,
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
	}
	numaInt, err = strconv.Atoi(numaStr)
	if err != nil {
		return numaInt, utils.FormatError(err)
	}
	return numaInt, nil
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

type NUMANodes map[int][]string

func (c *Collector) NUMAInfo() (NUMANodes, error) {
	out, err := c.Run("ls -d  /sys/devices/system/node/node[0-9]*")
	if err != nil {
		return nil, utils.FormatError(err)
	}

	numaMap := make(map[int][]string)
	for i, _ := range strings.SplitAfter(out, "\n") {
		out, err := c.Run(fmt.Sprintf("ls -d  /sys/devices/system/node/node%d/cpu[0-9]*", i))
		if err != nil {
			return nil, utils.FormatError(err)
		}
		cpus := make([]string, 0)
		for _, line := range strings.SplitAfter(out, "\n") {
			cpu := strings.SplitAfter(line, "cpu")[1]
			cpus = append(cpus, strings.TrimSpace(cpu))
		}
		numaMap[int(i)] = cpus
	}
	return numaMap, nil
}

// TotalNUMAs returns amount of NUMA nodes installed on the host
func (n NUMANodes) TotalNUMAs() int {
	return len(n)
}

// TotalCpus returns amount of CPUs bound to the NUMA
func (n NUMANodes) TotalCpus() int {
	amount := 0
	for _, v := range n {
		amount += len(v)
	}
	return amount
}

// CpusOnNUMA returns a slice of CPUs bound to the given NUMA node
func (n NUMANodes) CpusOnNUMA(numa int) ([]string, error) {
	if cpus, ok := n[numa]; ok {
		return cpus, nil
	}
	return nil, fmt.Errorf("NUMA %d not found", numa)
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
