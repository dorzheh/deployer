package host_hwfilter

import (
	"errors"
	"strings"

	"github.com/dorzheh/deployer/config/xmlinput"
	"github.com/dorzheh/deployer/deployer"
	"github.com/dorzheh/deployer/utils"
	"github.com/dorzheh/deployer/utils/hwinfo/host"
)

// GetAllowedNICs responsible for fetching Network ports satisfying user configuration
func GetAllowedNICs(data *xmlinput.XMLInputData, hidriver deployer.HostinfoDriver) (host.NICList, error) {
	allowedNics := host.NewNICList()
	priorityNics := host.NewNICList()
	physicalNics := host.NewNICList()

	nics, err := hidriver.NICs()
	if err != nil {
		return nil, utils.FormatError(err)
	}
	for _, nic := range nics {
		if deniedNIC(nic, &data.HostNics) {
			continue
		}
		// if counfiguration exists, treat physical ports
		if nic.Type == host.NicTypePhys || nic.Type == host.NicTypePhysVF {
			if data.HostNics.Allowed != nil {
				for _, anic := range data.HostNics.Allowed {
					if (anic.Model == "" && strings.Contains(nic.Vendor, anic.Vendor)) || nic.Model == anic.Model {
						if anic.Priority {
							priorityNics.Add(nic)
						} else {
							physicalNics.Add(nic)
						}
					}
				}
			}
		} else {
			// allow all ports in case configuration is empty or the port is not physical
			allowedNics.Add(nic)
		}
	}
	// if found ports with priority append them
	if priorityNics.Length() > 0 {
		allowedNics.AppendList(priorityNics)
	} else {
		allowedNics.AppendList(physicalNics)
	}
	return allowedNics, nil
}

// NicsByType verifies that Network ports in the list satisfy user configuration
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

			case xmlinput.ConTypeSRIOV:
				if nic.Type == host.NicTypePhysVF {
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

func NicDisjunctionFound(hnic *host.NIC, anics []*xmlinput.Allow) bool {
	for _, anic := range anics {
		if (anic.Model == "" && strings.Contains(hnic.Vendor, anic.Vendor)) || hnic.Model == anic.Model {
			if anic.Disjunction {
				return true
			}
		}
	}
	return false
}

// deniedNIC verifies whether a given Network port is denied by configuration or not
func deniedNIC(hnic *host.NIC, nics *xmlinput.HostNics) bool {
	for _, nic := range nics.Denied {
		if nic.Model == "" {
			if strings.Contains(hnic.Vendor, nic.Vendor) {
				return true
			}
		} else if strings.Contains(hnic.Model, nic.Model) {
			return true
		}
	}
	return false
}
