package libvirt_kvm

import (
	envdriver "github.com/dorzheh/deployer/drivers/env_driver/libvirt/libvirt_kvm"
	"github.com/dorzheh/deployer/utils"
	"github.com/dorzheh/deployer/utils/sysinfo"
	ssh "github.com/dorzheh/infra/comm/common"
)

const (
	multiQueueLibvirtVersion = "1.0.6"
	multiQueueKernelVersion  = "3.8"
)

func MultiQueueSupported(sshconfig *ssh.Config) (bool, error) {
	c := sysinfo.NewCollector(sshconfig)
	if c.KernelMajorMinorEqualOrGreaterThan(multiQueueKernelVersion) {
		d := envdriver.NewDriver(sshconfig)
		curVersion, err := d.Version()
		if err != nil {
			return false, utils.FormatError(err)
		}
		if curVersion >= multiQueueLibvirtVersion {
			return true, nil
		}
	}
	return false, nil
}

// For testing purpose
func MultiQueueSupportedMock(kernelVersion, libvirtVersion string, sshconfig *ssh.Config) (bool, error) {
	c := sysinfo.NewCollector(sshconfig)
	if c.KernelMajorMinorEqualOrGreaterThan(kernelVersion) {
		d := envdriver.NewDriver(sshconfig)
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
