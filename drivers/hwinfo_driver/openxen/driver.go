package openxen

import (
	"github.com/dorzheh/deployer/utils"
	"github.com/dorzheh/deployer/utils/hwinfo/host"
	ssh "github.com/dorzheh/infra/comm/common"
)

type HostinfoDriver struct {
	c         *host.Collector
	cpuinfo   *host.CPU
	numanodes host.NUMANodes
	nics      host.NICList
}

func NewHostinfoDriver(conf *ssh.Config, lshwpath, hwinfoFile string) (hi *HostinfoDriver, err error) {
	hi = new(HostinfoDriver)
	hi.c, err = host.NewCollector(conf, lshwpath, hwinfoFile)
	if err != nil {
		hi = nil
		err = utils.FormatError(err)
		return
	}
	return
}

func (hi *HostinfoDriver) Init() error {
	return hi.c.Hwinfo2Json()
}

// Returns RAM size
func (hi *HostinfoDriver) RAMSize() (int, error) {
	return hi.c.RAMSize()
}

// Returns available CPUs
func (hi *HostinfoDriver) CPUs() (int, error) {
	return hi.c.CPUs()
}

// Returns information related to the host's CPU
func (hi *HostinfoDriver) CPUInfo() (*host.CPU, error) {
	if hi.cpuinfo != nil {
		return hi.cpuinfo, nil
	}

	var err error
	hi.cpuinfo, err = hi.c.CPUInfo()
	if err != nil {
		return nil, err
	}
	return hi.cpuinfo, nil
}

// Returns amount of NUMA nodes and appropriate CPUs per NUMA node
func (hi *HostinfoDriver) NUMAInfo() (host.NUMANodes, error) {
	if hi.numanodes != nil {
		return hi.numanodes, nil
	}

	var err error
	hi.numanodes, err = hi.c.NUMAInfo()
	if err != nil {
		return nil, err
	}
	return hi.numanodes, nil
}

// Returns info related to the host's NICs
func (hi *HostinfoDriver) NICs() (host.NICList, error) {
	if hi.nics != nil {
		return hi.nics, nil
	}

	var err error
	hi.nics, err = hi.c.NICInfo()
	if err != nil {
		return nil, err
	}
	return hi.nics, nil
}
