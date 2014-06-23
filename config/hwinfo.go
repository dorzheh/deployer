package config

import (
	"io/ioutil"
	"os"
	"strings"

	"github.com/dorzheh/deployer/deployer"
	"github.com/dorzheh/infra/comm/ssh"
	"github.com/dorzheh/infra/utils/lshw"
	"github.com/dorzheh/mxj"
)

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
}

type HwInfo struct {
	run       func(string) (string, error)
	cacheFile string
	cmd       string
}

func NewHwInfoParser(cacheFile, lshwpath string, sshconf *ssh.Config) (*HwInfo, error) {
	lshwconf := &lshw.Config{[]lshw.Class{lshw.All}, lshw.FormatJSON}
	l, err := lshw.New(lshwpath, lshwconf)
	if err != nil {
		return nil, err
	}
	i := new(HwInfo)
	i.run = deployer.RunFunc(sshconf)
	i.cmd = l.Cmd()
	i.cacheFile = cacheFile
	return i, nil
}

func (i *HwInfo) Parse() error {
	out, err := i.run(i.cmd)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(i.cacheFile, []byte(out), 0)
}

func (i *HwInfo) NicsInfo() (map[NicType][]*NicInfo, error) {
	if _, err := os.Stat(i.cacheFile); err != nil {
		if err = i.Parse(); err != nil {
			return nil, err
		}
	}
	out, err := mxj.ReadMapsFromJsonFile(i.cacheFile)
	if err != nil {
		return nil, err
	}

	nicsMap := make(map[NicType][]*NicInfo)
	phys := make([]*NicInfo, 0)
	ovs := make([]*NicInfo, 0)

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
						nic.Driver = driver
						nic.Desc = "Open vSwitch interface"
						ovs = append(ovs, nic)
						nicsMap[NicTypeOVS] = ovs
					default:
						prod, ok := ch["product"].(string)
						if ok {
							vendor, _ := ch["vendor"].(string)
							nic.Driver = driver
							nic.Desc = vendor + " " + prod
							phys = append(phys, nic)
							nicsMap[NicTypePhys] = phys
						}
					}
				}
			}
		}
	}

	// lshw is unable to find linux bridges so let's do it manually
	res, err := i.run(`out="";for n in /sys/class/net/*;do [ -d $n/bridge ] && out="$out ${n##/sys/class/net/}";done;echo $out`)
	if err != nil {
		return nil, err
	}
	bridges := make([]*NicInfo, 0)
	if res != "" {
		for _, n := range strings.Split(res, " ") {
			br := &NicInfo{
				Name:   n,
				Driver: "bridge",
				Desc:   "Bridge interface",
			}
			bridges = append(bridges, br)
			nicsMap[NicTypeBridge] = bridges
		}
	}
	return nicsMap, nil
}
