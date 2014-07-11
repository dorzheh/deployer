package utils

import (
	"io/ioutil"
	"os"
	"strings"

	"github.com/dorzheh/deployer/deployer"
	ssh "github.com/dorzheh/infra/comm/common"
	"github.com/dorzheh/infra/utils/lshw"
	"github.com/dorzheh/mxj"
)

// CpuInfo contains CPU description and properties
type CpuInfo struct {
	Desc map[string]interface{}
	Cap  map[string]interface{}
}

type NicType string

const (
	NicTypePhys   NicType = "physical"
	NicTypeOVS    NicType = "openvswitch"
	NicTypeBridge NicType = "bridge"
)

type NicInfo struct {
	Name   string
	Driver string
	Desc   string
	Type   NicType
}

type HwInfoParser struct {
	run       func(string) (string, error)
	cacheFile string
	cmd       string
}

func NewHwInfoParser(cacheFile, lshwpath string, sshconf *ssh.Config) (*HwInfoParser, error) {
	lshwconf := &lshw.Config{[]lshw.Class{lshw.All}, lshw.FormatJSON}
	l, err := lshw.New(lshwpath, lshwconf)
	if err != nil {
		return nil, err
	}
	i := new(HwInfoParser)
	i.run = deployer.RunFunc(sshconf)
	i.cmd = l.Cmd()
	i.cacheFile = cacheFile
	return i, nil
}

func (i *HwInfoParser) Parse() error {
	out, err := i.run(i.cmd)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(i.cacheFile, []byte(out), 0)
}

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

func (i *HwInfoParser) NicsInfo() (map[int]*NicInfo, error) {
	if _, err := os.Stat(i.cacheFile); err != nil {
		if err = i.Parse(); err != nil {
			return nil, err
		}
	}
	out, err := mxj.ReadMapsFromJsonFile(i.cacheFile)
	if err != nil {
		return nil, err
	}

	nicsMap := make(map[int]*NicInfo)
	index := 0
	deep := []string{"children.children.children.children", "children"}
	for _, m := range out {
		for _, d := range deep {
			r, _ := m.ValuesForPath(d)
			for _, n := range r {
				ch := n.(map[string]interface{})
				if ch["description"] == "Ethernet interface" {
					nic := new(NicInfo)
					nic.Name = ch["logicalname"].(string)
					driver := ch["configuration"].(map[string]interface{})["driver"].(string)
					switch driver {
					case "openvswitch":
						nic.Desc = "Open vSwitch interface"
						nic.Type = NicTypeOVS
					default:
						prod, ok := ch["product"].(string)
						if ok {
							vendor, _ := ch["vendor"].(string)
							nic.Desc = vendor + " " + prod
							nic.Type = NicTypePhys
						}
					}
					nic.Driver = driver
					nicsMap[index] = nic
					index++
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
			nicsMap[index] = br
			index++
		}
	}
	return nicsMap, nil
}
