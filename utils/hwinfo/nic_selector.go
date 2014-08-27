package hwinfo

import (
	"strings"

	"github.com/dorzheh/deployer/config/libvirt/xmlinput"
)

func DeniedNIC(n *NIC, nics *xmlinput.NIC) bool {
	for _, nic := range nics.Denied {
		if nic.Model == "" {
			if strings.Contains(n.Desc, nic.Vendor) {
				return true
			}
		} else if strings.Contains(n.Desc, nic.Model) {
			return true
		}
	}
	return false
}

func AllowedNIC(n *NIC, nics *xmlinput.NIC) bool {
	for _, nic := range nics.Allowed {
		if nic.Model == "" && strings.Contains(n.Desc, nic.Vendor) {
			return true
		}
		if strings.Contains(n.Desc, nic.Model) {
			return true
		}
	}
	return false
}
