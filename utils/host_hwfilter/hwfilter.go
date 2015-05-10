package host_hwfilter

import (
	"errors"
	"strings"

	"github.com/dorzheh/deployer/config/common/xmlinput"
	"github.com/dorzheh/deployer/deployer"
	"github.com/dorzheh/deployer/utils"
	"github.com/dorzheh/deployer/utils/hwinfo/host"
)

func GetAllowedNICs(data *xmlinput.XMLInputData, hidriver deployer.HostinfoDriver) (host.NICList, error) {
	allowedNics := host.NewNICList()
	nics, err := hidriver.NICs()
	if err != nil {
		return nil, utils.FormatError(err)
	}
	for _, n := range nics {
		if deniedNIC(n, &data.HostNics) {
			continue
		}
		if n.Type == host.NicTypePhys {
			if !allowedNIC(n, &data.HostNics) {
				continue
			}
		}
		allowedNics.Add(n)
	}
	return allowedNics, nil
}

func NicsByType(nics host.NICList, types []xmlinput.ConnectionMode) (host.NICList, error) {
	if len(nics) == 0 {
		return nil, utils.FormatError(errors.New("nics of type host.NICList is empty"))
	}

	list := host.NewNICList()
	for _, contype := range types {
		for _, nic := range nics {
			switch contype {
			case xmlinput.ConTypeBridged:
				if nic.Type == host.NicTypeBridge {
					list.Add(nic)
				}

			case xmlinput.ConTypeOVS:
				if nic.Type == host.NicTypeOVS {
					list.Add(nic)
				}

			case xmlinput.ConTypeDirect:
				if nic.Type == host.NicTypePhys {
					list.Add(nic)
				}

			case xmlinput.ConTypePassthrough:
				if nic.Type == host.NicTypePhys {
					list.Add(nic)
				}

			case xmlinput.ConTypeVirtualNetwork:
				if nic.Type == host.NicTypeVirtualNetwork {
					list.Add(nic)
				}

			default:
				return nil, utils.FormatError(errors.New("unexpected type " + string(contype)))
			}
		}
	}
	return list, nil
}

func deniedNIC(hwnics *host.NIC, nics *xmlinput.HostNics) bool {
	for _, nic := range nics.Denied {
		if nic.Model == "" {
			if strings.Contains(hwnics.Vendor, nic.Vendor) {
				return true
			}
		} else if strings.Contains(hwnics.Model, nic.Model) {
			return true
		}
	}
	return false
}

func allowedNIC(n *host.NIC, nics *xmlinput.HostNics) bool {
	if nics.Allowed == nil {
		return true
	}
	for _, nic := range nics.Allowed {
		if nic.Model == "" && strings.Contains(n.Vendor, nic.Vendor) {
			return true
		}
		if strings.Contains(n.Model, nic.Model) {
			return true
		}
	}
	return false
}
