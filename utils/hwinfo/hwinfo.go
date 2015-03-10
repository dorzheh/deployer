// Responsible for parsing lshw output

package hwinfo

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

type Parser struct {
	Run       func(string) (string, error)
	parse     func() error
	cacheFile string
}

// NewParser constructs new lshw parser
// The output will be represented in JSON format
func NewParser(cacheFile, lshwpath string, sshconf *ssh.Config) (*Parser, error) {
	i := new(Parser)
	i.Run = utils.RunFunc(sshconf)
	i.parse = parseFunc(i, cacheFile, lshwpath, sshconf)
	return i, nil
}

// Parse parses lshw output
func (i *Parser) Parse() error {
	if err := i.parse(); err != nil {
		return utils.FormatError(err)
	}
	out, err := i.Run("lshw -class network -class cpu -json")
	if err != nil {
		return utils.FormatError(err)
	}
	return ioutil.WriteFile(i.cacheFile, []byte(out), 0)
}

// CPU contains CPU description and properties
type CPU struct {
	Desc map[string]interface{}
	Cap  map[string]interface{}
}

// CPUInfo gathers information related to installed CPUs
func (i *Parser) CPUInfo() (*CPU, error) {
	if _, err := os.Stat(i.cacheFile); err != nil {
		if err = i.Parse(); err != nil {
			return nil, utils.FormatError(err)
		}
	}

	out, err := mxj.NewMapsFromJsonFile(i.cacheFile)
	if err != nil {
		return nil, utils.FormatError(err)
	}

	c := new(CPU)
	c.Desc = make(map[string]interface{})
	c.Cap = make(map[string]interface{})
	for _, s := range out {
		r, _ := s.ValuesForPath("children.children")
		for _, n := range r {
			ch := n.(map[string]interface{})
			if ch["id"] == "cpu:0" {
				for k, v := range ch {
					if k != "capabilities" {
						c.Desc[k] = v
					}
				}
				for k, v := range ch["capabilities"].(map[string]interface{}) {
					c.Cap[k] = v
				}
			}
		}
	}
	return c, nil
}

func (p *Parser) CPUs() (int, error) {
	cpustr, err := p.Run(`grep -c ^processor /proc/cpuinfo`)
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

	// Description
	Desc string

	// Port type
	Type NicType
}

// NICInfo gathers information related to installed NICs
func (p *Parser) NICInfo() (NICList, error) {
	if _, err := os.Stat(p.cacheFile); err != nil {
		if err = p.Parse(); err != nil {
			return nil, utils.FormatError(err)
		}
	}
	out, err := mxj.NewMapsFromJsonFile(p.cacheFile)
	if err != nil {
		return nil, utils.FormatError(err)
	}

	list := NewNICList()
	for _, m := range out {
		val, err := m.ValuesForKey("id")
		if err != nil {
			continue
		}
		for _, name := range val {
			if name == "network" {
				nic := p.findInterface(m.Old())
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
				nic := p.findInterface(n)
				if nic != nil {
					list.Add(nic)
				}
			}
		}
	}

	// lshw is unable to find linux bridges so let's do it manually
	res, err := p.Run(`out="";for n in /sys/class/net/*;do [ -d $n/bridge ] && out="$out ${n##/sys/class/net/}";done;echo $out`)
	if err != nil {
		return nil, utils.FormatError(err)
	}
	if res != "" {
		for _, n := range strings.Split(res, " ") {
			br := &NIC{
				Name:   n,
				Driver: "bridge",
				Desc:   "Bridge interface",
				Type:   NicTypeBridge,
			}
			list.Add(br)
		}
	}
	list.SortByPCI()
	return list, nil
}

func (p *Parser) NUMANodes() (map[int][]int, error) {
	out, err := p.Run("ls -d  /sys/devices/system/node/node[0-9]*")
	if err != nil {
		return nil, utils.FormatError(err)
	}

	numaMap := make(map[int][]int)
	for i, _ := range strings.SplitAfter(out, "\n") {
		out, err := p.Run(fmt.Sprintf("ls -d  /sys/devices/system/node/node%d/cpu[0-9]*", i))
		if err != nil {
			return nil, utils.FormatError(err)
		}
		cpus := make([]int, 0)
		for _, line := range strings.SplitAfter(out, "\n") {
			cpuStr := strings.SplitAfter(line, "cpu")[1]
			cpu, err := strconv.Atoi(strings.TrimSpace(cpuStr))
			if err != nil {
				return nil, utils.FormatError(err)
			}
			cpus = append(cpus, int(cpu))
		}
		numaMap[int(i)] = cpus
	}
	return numaMap, nil
}

// RAMSize gathers information related to the installed amount of RAM in MB
func (p *Parser) RAMSize() (int, error) {
	out, err := p.Run("grep MemTotal /proc/meminfo")
	if err != nil {
		return 0, utils.FormatError(err)
	}
	var ramsize int
	fmt.Sscanf(out, "MemTotal: %d %s", &ramsize)
	return ramsize / 1024, nil
}

func parseFunc(i *Parser, cacheFile, lshwpath string, sshconf *ssh.Config) func() error {
	return func() error {
		if lshwpath == "" {
			out, err := i.Run("which lshw")
			if err != nil {
				return utils.FormatError(fmt.Errorf("%s [%v]", out, err))
			}
			lshwpath = out
		} else {
			if sshconf != nil {
				dir, err := utils.UploadBinaries(sshconf, lshwpath)
				if err != nil {
					return utils.FormatError(err)
				}
				lshwpath = filepath.Join(dir, filepath.Base(lshwpath))
			}
		}
		i.cacheFile = cacheFile
		return nil
	}
}

func (p *Parser) findInterface(json interface{}) *NIC {
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
				if _, err := p.Run(fmt.Sprintf("[[ -d /sys/class/net/%s/master || -d /sys/class/net/%s/brport ]]", name, name)); err == nil {
					return nil
				}
				nic.PCIAddr = ch["businfo"].(string)
				nic.Vendor = vendor
				nic.Model = prod
				nic.Desc = vendor + " " + prod
				nic.Type = NicTypePhys
			}
		}
		nic.Driver = driver
	}
	return nic
}
