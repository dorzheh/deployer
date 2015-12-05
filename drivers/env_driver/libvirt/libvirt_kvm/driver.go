package libvirt_kvm

import (
	"fmt"
	"strings"
	"sync"

	"github.com/clbanning/mxj"
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
	d.Lock()
	defer d.Unlock()

	if _, err := d.Run("virsh define " + domainConfig); err != nil {
		return utils.FormatError(err)
	}
	return nil
}

func (d *Driver) StartDomain(name string) error {
	d.Lock()
	defer d.Unlock()

	if _, err := d.Run("virsh start " + name); err != nil {
		return utils.FormatError(err)
	}
	return nil
}

func (d *Driver) DestroyDomain(name string) error {
	d.Lock()
	defer d.Unlock()

	if _, err := d.Run("virsh destroy " + name); err != nil {
		return utils.FormatError(err)
	}
	return nil
}

func (d *Driver) UndefineDomain(name string) error {
	d.Lock()
	defer d.Unlock()

	if _, err := d.Run("virsh undefine " + name); err != nil {
		return utils.FormatError(err)
	}
	return nil
}

func (d *Driver) SetAutostart(name string) error {
	d.Lock()
	defer d.Unlock()

	if _, err := d.Run("virsh autostart " + name); err != nil {
		return utils.FormatError(err)
	}
	return nil
}

func (d *Driver) DomainExists(name string) bool {
	d.Lock()
	defer d.Unlock()

	if _, err := d.Run("virsh dominfo " + name); err != nil {
		return false
	}
	return true
}

// Emulator returns appropriate path to QEMU emulator for a given architecture
func (d *Driver) Emulator(arch string) (string, error) {
	switch arch {
	case "x86_64":
	case "i686":
	default:
		return "", utils.FormatError(fmt.Errorf("Unsupported architecture(%s).Only i686 and x86_64 supported", arch))
	}

	out, err := d.Run("virsh capabilities")
	if err != nil {
		return "", utils.FormatError(err)
	}

	m, err := mxj.NewMapXml([]byte(out))
	if err != nil {
		return "", utils.FormatError(err)
	}

	v, _ := m.ValuesForPath("capabilities.guest.arch", "-name:"+arch)
	return v[0].(map[string]interface{})["emulator"].(string), nil
}

// Version returns libvirt API version
func (d *Driver) Version() (string, error) {
	out, err := d.Run("virsh version|grep \"Using library\"")
	if err != nil {
		return "", utils.FormatError(err)
	}
	return strings.Split(out, " ")[3], nil

}

// Returns maximal Virtual CPUs per guest
func (d *Driver) MaxVCPUsPerGuest() int {
	return 64
}

// Returns true if all domains configured for CPU affinity
func (d *Driver) AllCPUsPinned() (bool, error) {
	out, err := d.Run("virsh list --name --all")
	if err != nil {
		return false, err
	}

	amountOfdomains := 0
	amountOfPinnedDomains := 0
	for _, s := range strings.Split(out, "\n") {
		if s != "" {
			out, err := d.Run("virsh vcpuinfo " + s)
			if err != nil {
				return false, err
			}

			amountOfdomains++
			if strings.Contains(out, "-") {
				amountOfPinnedDomains++
			}
		}
	}
	if amountOfPinnedDomains == amountOfdomains {
		return true, nil
	}
	return false, nil
}
