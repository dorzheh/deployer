// A wrapper for hwinfo.
// Provides a couple fo advantages over calling hwinfo directly:
// - ability to customize the parser according to deployment type
// - interface implementation
package libvirt

import (
	"github.com/dorzheh/deployer/utils"
	"github.com/dorzheh/deployer/utils/hwinfo"
	ssh "github.com/dorzheh/infra/comm/common"
)

type HostinfoDriver struct {
	parser *hwinfo.Parser
}

func NewHostinfoDriver(cacheFile, lshwpath string, conf *ssh.Config) (hi *HostinfoDriver, err error) {
	hi = new(HostinfoDriver)
	hi.parser, err = hwinfo.NewParser(cacheFile, lshwpath, conf)
	if err != nil {
		hi = nil
		err = utils.FormatError(err)
		return
	}
	return
}

func (hi *HostinfoDriver) Init() error {
	return hi.parser.Parse()
}

// Returns RAM size
func (hi *HostinfoDriver) RAMSize() (int, error) {
	return hi.parser.RAMSize()
}

// Returns available CPUs
func (hi *HostinfoDriver) CPUs() (int, error) {
	return hi.parser.CPUs()
}

// Returns information related to the host's CPU
func (hi *HostinfoDriver) CPUInfo() (*hwinfo.CPU, error) {
	return hi.parser.CPUInfo()
}

// Returns amount of NUMA nodes and appropriate CPUs per NUMA node
func (hi *HostinfoDriver) NUMANodes() (map[int][]int, error) {
	return hi.parser.NUMANodes()
}

// Returns info related to the host's NICs
func (hi *HostinfoDriver) NICs() ([]*hwinfo.NIC, error) {
	return hi.parser.NICInfo()
}
