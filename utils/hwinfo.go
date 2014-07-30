// Responsible for parsing lshw output

package utils

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	ssh "github.com/dorzheh/infra/comm/common"
	"github.com/dorzheh/infra/utils/lshw"
	"github.com/dorzheh/mxj"
)

// CpuInfo contains CPU description and properties
type CpuInfo struct {
	Cpus int
	Desc map[string]interface{}
	Cap  map[string]interface{}
}

// supported NIC types
type NicType string

const (
	NicTypePhys   NicType = "physical"
	NicTypeOVS    NicType = "openvswitch"
	NicTypeBridge NicType = "bridge"
)

// NIC information
type NicInfo struct {
	// port name (eth0,br0...)
	Name string

	// NIC driver(bridge,openvswitch...)
	Driver string

	// Description
	Desc string

	// PCI Address
	PCIAddr string

	// Port type
	Type NicType
}

type HwInfoParser struct {
	run       func(string) (string, error)
	cacheFile string
	cmd       string
}

// NewHwInfoParser constructs new lshw parser
// The output will be represented in JSON format
func NewHwInfoParser(cacheFile, lshwpath string, sshconf *ssh.Config) (*HwInfoParser, error) {
	i := new(HwInfoParser)
	i.run = RunFunc(sshconf)
	if lshwpath == "" {
		out, err := i.run("which lshw")
		if err != nil {
			return nil, fmt.Errorf("%s [%v]", out, err)
		}
		lshwpath = out
	} else {
		if sshconf != nil {
			dir, err := UploadBinaries(sshconf, lshwpath)
			if err != nil {
				return nil, err
			}
			lshwpath = filepath.Join(dir, filepath.Base(lshwpath))

		}
	}

	lshwconf := &lshw.Config{[]lshw.Class{lshw.All}, lshw.FormatJSON}
	l, err := lshw.New(lshwpath, lshwconf)
	if err != nil {
		return nil, err
	}
	i.cmd = l.Cmd()
	i.cacheFile = cacheFile
	return i, nil
}

// Parse parses lshw output
func (i *HwInfoParser) Parse() error {
	out, err := i.run(i.cmd)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(i.cacheFile, []byte(out), 0)
}

// CpuInfo gathers information related to installed CPUs
func (i *HwInfoParser) CpuInfo() (*CpuInfo, error) {
	if _, err := os.Stat(i.cacheFile); err != nil {
		if err = i.Parse(); err != nil {
			return nil, err
		}
	}

	out, err := mxj.ReadMapsFromJsonFile(i.cacheFile)
	if err != nil {
		return nil, err
	}

	c := new(CpuInfo)
	cpustr, err := i.run(`grep -c processor /proc/cpuinfo`)
	if err != nil {
		return nil, err
	}

	c.Cpus, err = strconv.Atoi(strings.Trim(cpustr, "\n"))
	if err != nil {
		return nil, err
	}

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

// NicInfo gathers information related to installed NICs
func (i *HwInfoParser) NicsInfo(supNicVendors []string) ([]*NicInfo, error) {
	if _, err := os.Stat(i.cacheFile); err != nil {
		if err = i.Parse(); err != nil {
			return nil, err
		}
	}
	out, err := mxj.ReadMapsFromJsonFile(i.cacheFile)
	if err != nil {
		return nil, err
	}

	nics := make([]*NicInfo, 0)
	deep := []string{"children.children.children.children", "children"}
	for _, m := range out {
		for _, d := range deep {
			r, _ := m.ValuesForPath(d)
			for _, n := range r {
				ch := n.(map[string]interface{})
				if ch["description"] == "Ethernet interface" {
					name := ch["logicalname"].(string)
					if name == "ovs-system" {
						continue
					}
					nic := new(NicInfo)
					nic.Name = name
					driver := ch["configuration"].(map[string]interface{})["driver"].(string)
					switch driver {
					case "tun":
						continue
					case "openvswitch":
						nic.Desc = "Open vSwitch interface"
						nic.Type = NicTypeOVS
					default:
						prod, ok := ch["product"].(string)
						if ok {
							vendor, _ := ch["vendor"].(string)
							if len(supNicVendors) > 0 {
								found := false
								for _, v := range supNicVendors {
									if v == vendor {
										found = true
										break
									}
								}
								if !found {
									continue
								}
							}
							nic.PCIAddr = ch["businfo"].(string)
							nic.Desc = vendor + " " + prod
							nic.Type = NicTypePhys
						}
					}
					nic.Driver = driver
					nics = append(nics, nic)
				}
			}
		}
	}

	// lshw is unable to find linux bridges so let's do it manually
	res, err := i.run(`out="";for n in /sys/class/net/*;do [ -d $n/bridge ] && out="$out ${n##/sys/class/net/}";done;echo $out`)
	if err != nil {
		return nil, err
	}
	if res != "" {
		for _, n := range strings.Split(res, " ") {
			br := &NicInfo{
				Name:   n,
				Driver: "bridge",
				Desc:   "Bridge interface",
				Type:   NicTypeBridge,
			}
			nics = append(nics, br)
		}
	}
	return nics, nil
}

// RAMSize gathers information related to the installed RAM pools
func (i *HwInfoParser) RAMSize() (uint, error) {
	var ramsize uint = 0
	if _, err := os.Stat(i.cacheFile); err != nil {
		if err = i.Parse(); err != nil {
			return ramsize, err
		}
	}
	out, err := mxj.ReadMapsFromJsonFile(i.cacheFile)
	if err != nil {
		return ramsize, err
	}
	for _, s := range out {
		r, _ := s.ValuesForPath("children.children")
		for _, n := range r {
			ch := n.(map[string]interface{})
			if ch["id"] == "memory" {
				ramsize = uint(ch["size"].(float64))
			}
		}
	}
	return uint(ramsize), nil
}
