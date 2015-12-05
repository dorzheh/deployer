package xen_xl

import (
	"fmt"
	"strings"
	"sync"

	"github.com/dorzheh/deployer/utils"
	ssh "github.com/dorzheh/infra/comm/common"
)

type Driver struct {
	sync.Mutex

	Run func(string) (string, error)
}

func NewDriver(config *ssh.Config) *Driver {
	d := new(Driver)
	d.Run = utils.RunFunc(config)
	return d
}

func (d *Driver) DefineDomain(domainConfig string) error {
	return nil
}

func (d *Driver) StartDomain(domainConfig string) error {
	d.Lock()
	defer d.Unlock()

	if _, err := d.Run("xl create " + domainConfig); err != nil {
		return utils.FormatError(err)
	}
	return nil
}

func (d *Driver) DestroyDomain(name string) error {
	d.Lock()
	defer d.Unlock()

	if _, err := d.Run("xl destroy " + name); err != nil {
		return utils.FormatError(err)
	}
	return nil
}

func (d *Driver) UndefineDomain(name string) error {
	return nil
}

func (d *Driver) SetAutostart(name string) error {
	d.Lock()
	defer d.Unlock()

	if _, err := d.Run(fmt.Sprintf("ln -fs /etc/xen/%s.cfg /etc/xen/auto/%s.cfg", name, name)); err != nil {
		return utils.FormatError(err)
	}
	return nil
}

func (d *Driver) DomainExists(name string) bool {
	d.Lock()
	defer d.Unlock()

	if _, err := d.Run("xl list " + name); err != nil {
		return false
	}
	return true
}

func (d *Driver) Emulator(arch string) (string, error) {
	return "", nil
}

// Version returns Xen version
func (d *Driver) Version() (string, error) {
	out, err := d.Run("xl info|grep xen_version")
	if err != nil {
		return "", utils.FormatError(err)
	}
	return strings.Split(out, ":")[1], nil

}

// Returns maximal Virtual CPUs per guest
func (d *Driver) MaxVCPUsPerGuest() int {
	return 64
}

// Returns true if all domains configured for CPU affinity
func (d *Driver) AllCPUsPinned() (bool, error) {
	return true, nil
}
