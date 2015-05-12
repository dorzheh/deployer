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
	Desc map[string]interface{}
	Cap  map[string]interface{}
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
	cpu.Desc = make(map[string]interface{})
	cpu.Cap = make(map[string]interface{})
	for _, s := range out {
		r, _ := s.ValuesForPath("children.children")
		for _, n := range r {
			ch := n.(map[string]interface{})
			if ch["id"] == "cpu:0" {
				for k, v := range ch {
					if k != "capabilities" {
						cpu.Desc[k] = v
					}
				}
				for k, v := range ch["capabilities"].(map[string]interface{}) {
					cpu.Cap[k] = v
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
		val, err := m.ValuesForKey("class")
		if err != nil {
			continue
		}
		for _, name := range val {
			if name == "network" {
				nic := c.findInterface(m.Old())
				if nic != nil {
					list.Add(nic)
				}
			}
		}
	}

	deep := []string{"children.children.children.children.children",
		"children.children.children.children",
		"children.children.children",
		"children.children",
		"children"}
	for _, m := range out {
		for _, d := range deep {
			r, _ := m.ValuesForPath(d)
			for _, n := range r {
				nic := c.findInterface(n)
				if nic != nil {
					list.Add(nic)
				}
			}
		}
	}

	// lshw is unable to find linux bridges so let's do it manually
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

// findInterface parses appropriate data structure and gather
// information related to the Network interfaces
func (c *Collector) findInterface(json interface{}) *NIC {
	ch := json.(map[string]interface{})
	var nic *NIC
	if ch["description"] == "Ethernet interface" ||
		ch["description"] == "Ethernet controller" {
		name, ok := ch["logicalname"].(string)
		if !ok {
			name = "N/A"
		}
		if name == "ovs-system" {
			return nil
		}
		nic = new(NIC)
		nic.Name = name
		nic.PCIAddr = "N/A"
		nic.NUMANode = -1
		driver, ok := ch["configuration"].(map[string]interface{})["driver"].(string)
		if !ok {
			driver = "N/A"
		}
		switch driver {
		case "tun", "veth", "macvlan":
			return nil
		case "openvswitch":
			nic.Desc = "Open vSwitch interface"
			nic.Type = NicTypeOVS
		default:
			prod, ok := ch["product"].(string)
			if ok {
				vendor, _ := ch["vendor"].(string)
				if _, err := c.Run(fmt.Sprintf("[[ -d /sys/class/net/%s/master || -d /sys/class/net/%s/brport ]]", name, name)); err == nil {
					return nil
				}
				nic.PCIAddr = strings.Split(ch["businfo"].(string), "@")[1]
				nic.Vendor = vendor
				nic.Model = prod
				numa, err := numa4Nic(nic.PCIAddr)
				if err != nil {
					return nil
				}
				nic.NUMANode = numa
				nic.Desc = vendor + " " + prod
				nic.Type = NicTypePhys
			}
		}
		nic.Driver = driver
	}
	return nic
}

// numa4Nic fetchs NUMA node for appropriate PCI device
func numa4Nic(pciAddr string) (int, error) {
	numaInt := 0
	buf, err := ioutil.ReadFile("/sys/bus/pci/devices/" + pciAddr + "/numa_node")
	if err != nil {
		return 0, err
	}
	numaInt, err = strconv.Atoi(strings.Trim(string(buf), "\n"))
	if err != nil {
		return numaInt, err
	}
	if numaInt == -1 {
		numaInt = 0
	}
	return numaInt, nil
}
