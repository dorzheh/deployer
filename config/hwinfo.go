package config

import (
	"io/ioutil"
	"os"

	"github.com/dorzheh/deployer/deployer"
	"github.com/dorzheh/infra/comm/ssh"
	"github.com/dorzheh/infra/utils/lshw"
	"github.com/dorzheh/mxj"
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

func NewHwInfoParser(lshwpath string, sshconf *ssh.Config) (*HwInfo, error) {
	lshwconf := &lshw.Config{[]lshw.Class{lshw.All}, lshw.FormatXML}
	l, err := lshw.New(lshwpath, lshwconf)
	if err != nil {
		return nil, err
	}
	i := new(HwInfo)
	i.run = deployer.RunFunc(sshconf)
	i.cmd = l.Cmd()
	return i, nil
}

func (i *HwInfo) Parse() error {
	out, err := i.run(i.cmd)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(i.cacheFile, []byte(out), 0)
}

func (i *HwInfo) NicsInfo() ([]*NicInfo, error) {
	if _, err := os.Stat(i.cacheFile); err != nil {
		if err = i.Parse(); err != nil {
			return nil, err
		}
	}
	out, err := mxj.ReadMapsFromXmlFile(i.cacheFile)
	if err != nil {
		return nil, err
	}

	nics := make([]*NicInfo, 0)
	for _, m := range out {
		v, _ := m.ValuesForPath("list.node")
		for _, n := range v {
			nic := new(NicInfo)
			nic.Name, _ = n.(map[string]interface{})["logicalname"].(string)
			vendor, ok := n.(map[string]interface{})["vendor"].(string)
			if ok {
				product, _ := n.(map[string]interface{})["vendor"].(string)
				nic.Desc = vendor + " " + product
			}
			conf := n.(map[string]interface{})["configuration"].(map[string]interface{})
			for _, c := range conf {
				for _, i := range c.([]interface{}) {
					if i.(map[string]interface{})["-id"] == "driver" {
						nic.Driver = i.(map[string]interface{})["-value"].(string)
					}
				}
			}
			nics = append(nics, nic)
		}
	}
	return nics, nil
}
