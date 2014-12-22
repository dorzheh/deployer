package libvirt

import (
	info "github.com/dorzheh/deployer/post_processor/libvirt"
	"github.com/dorzheh/deployer/utils/sysinfo"
	ssh "github.com/dorzheh/infra/comm/common"
)

const (
	multiQueueLibvirtVersion = "1.0.6"
	multiQueueKernelVersion  = "3.8"
)

func MultiQueueSupported(sshconfig *ssh.Config) (bool, error) {
	p := sysinfo.NewParser(sshconfig)
	if p.KernelMajorMinorEqualOrGreaterThan(multiQueueKernelVersion) {
		d := info.NewDriver(sshconfig)
		curVersion, err := d.Version()
		if err != nil {
			return false, err
		}
		if curVersion >= multiQueueLibvirtVersion {
			return true, nil
		}
	}
	return false, nil
}

// For testing purpose
func MultiQueueSupportedMock(kernelVersion, libvirtVersion string, sshconfig *ssh.Config) (bool, error) {
	p := sysinfo.NewParser(sshconfig)
	if p.KernelMajorMinorEqualOrGreaterThan(kernelVersion) {
		d := info.NewDriver(sshconfig)
		curVersion, err := d.Version()
		if err != nil {
			return false, err
		}
		if curVersion >= libvirtVersion {
			return true, nil
		}
	}
	return false, nil
}
