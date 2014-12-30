package hwfilter

import (
	"errors"
	"strings"

	"github.com/dorzheh/deployer/config/common/xmlinput"
	"github.com/dorzheh/deployer/deployer"
	"github.com/dorzheh/deployer/utils/hwinfo"
)

func GetAllowedNICs(data *xmlinput.XMLInputData, hidriver deployer.HostinfoDriver) (hwinfo.NICList, error) {
	allowedNics := hwinfo.NewNICList()
	nics, err := hidriver.NICs()
	if err != nil {
		return nil, err
	}
	for _, n := range nics {
		if deniedNIC(n, &data.NICs) {
			continue
		}
		if n.Type == hwinfo.NicTypePhys {
			if !allowedNIC(n, &data.NICs) {
				continue
			}
		}
		allowedNics.Add(n)
	}
	return allowedNics, nil
}

func NicsByType(nics hwinfo.NICList, types []xmlinput.ConnectionMode) (hwinfo.NICList, bool, error) {
	modePassthrough := false
	if len(nics) == 0 {
		return nil, modePassthrough, errors.New("nics of type hwinfo.NICList is empty")
	}

	list := hwinfo.NewNICList()
	for _, contype := range types {
		for _, nic := range nics {
			switch contype {
			case xmlinput.ConTypeBridged:
				if nic.Type == hwinfo.NicTypeBridge {
					list.Add(nic)
				}

			case xmlinput.ConTypeOVS:
				if nic.Type == hwinfo.NicTypeOVS {
					list.Add(nic)
				}

			case xmlinput.ConTypeDirect:
				if nic.Type == hwinfo.NicTypePhys {
					list.Add(nic)
				}

			case xmlinput.ConTypePassthrough:
				if nic.Type == hwinfo.NicTypePhys {
					list.Add(nic)
				}
				modePassthrough = true

			default:
				return nil, modePassthrough, errors.New("unexpected type " + string(contype))
			}
		}
	}
	return list, modePassthrough, nil
}

func deniedNIC(hwnics *hwinfo.NIC, nics *xmlinput.NICs) bool {
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

func allowedNIC(n *hwinfo.NIC, nics *xmlinput.NICs) bool {
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
