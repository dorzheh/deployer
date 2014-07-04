package libvirt

import (
	"sync"

	"github.com/dorzheh/deployer/deployer"
	ssh "github.com/dorzheh/infra/comm/common"
	"github.com/dorzheh/mxj"
)

type Driver struct {
	sync.Mutex
	run func(string) (string, error)
}

func NewDriver(config *ssh.Config) *Driver {
	d := new(Driver)
	d.run = deployer.RunFunc(config)
	return d
}

func (d *Driver) DefineDomain(domainConfig string) error {
	d.Lock()
	defer d.Unlock()
	if _, err := d.run("virsh define " + domainConfig); err != nil {
		return err
	}
	return nil
}

func (d *Driver) StartDomain(name string) error {
	d.Lock()
	defer d.Unlock()
	if _, err := d.run("virsh start " + name); err != nil {
		return err
	}
	return nil
}

func (d *Driver) DestroyDomain(name string) error {
	d.Lock()
	defer d.Unlock()
	if _, err := d.run("virsh destroy " + name); err != nil {
		return err
	}
	return nil
}

func (d *Driver) UndefineDomain(name string) error {
	d.Lock()
	defer d.Unlock()
	if _, err := d.run("virsh undefine " + name); err != nil {
		return err
	}
	return nil
}

// Define is responsible for creating a new domain acording to provided XML template
func (d *Driver) SetAutostart(name string) error {
	d.Lock()
	defer d.Unlock()
	if _, err := d.run("virsh autostart " + name); err != nil {
		return err
	}
	return nil
}

func (d *Driver) DomainExists(name string) bool {
	d.Lock()
	defer d.Unlock()
	if _, err := d.run("virsh dominfo " + name); err != nil {
		return false
	}
	return true
}

func (d *Driver) Emulator() (string, error) {
	d.Lock()
	defer d.Unlock()
	out, err := d.run("virsh capabilities")
	if err != nil {
		return "", err
	}

	m, err := mxj.NewMapXml([]byte(out))
	if err != nil {
		return "", err
	}

	v, _ := m.ValuesForPath("capabilities.guest.arch", "-name:x86_64")
	return v[0].(map[string]interface{})["emulator"].(string), nil
}
