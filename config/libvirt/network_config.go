package libvirt

import (
	"errors"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/dorzheh/deployer/config/libvirt/xmlinput"
	"github.com/dorzheh/deployer/utils"
	"github.com/dorzheh/deployer/utils/hwinfo"
)

const (
	TmpltFileBridgedOVS = "template_bridged_ovs.xml"
	TmpltFileBridged    = "template_bridged.xml"
	TmpltFileDirect     = "template_direct.xml"
)

type PassthroughData struct {
	Bus      string
	Slot     string
	Function string
}

type BridgedOVSData struct {
	OVSBridge string
}

type BridgedData struct {
	Bridge string
}

type DirectData struct {
	IfaceName string
}

// SetNetworkData is responsible for adding to the metadata appropriate entries
// related to the network configuration
func SetNetworkData(ni map[string]*hwinfo.NIC, nics []xmlinput.Allow, templatesDir string) (string, error) {
	var data string
	for _, port := range ni {
		switch port.Type {
		case hwinfo.NicTypePhys:
			var tempData []byte
			var err error

			for _, nic := range nics {
				switch bindingMode(port, &nic) {
				case "passthrough":
					tempData, err = ProcessTemplatePassthrough(port.PCIAddr)
					if err != nil {
						return "", err
					}
				case "direct":
					buf, err := ioutil.ReadFile(filepath.Join(templatesDir, TmpltFileBridgedOVS))
					if err == nil {
						TmpltDirect = string(buf)
					}
					tempData, err = utils.ProcessTemplate(TmpltDirect, &DirectData{port.Name})
					if err != nil {
						return "", err
					}
				default:
					return "", errors.New("supported modes - direct or passthrough")
				}
			}
			data += string(tempData) + "\n"

		case hwinfo.NicTypeOVS:
			buf, err := ioutil.ReadFile(filepath.Join(templatesDir, TmpltFileBridgedOVS))
			if err == nil {
				TmpltBridgedOVS = string(buf)
			}
			tempData, err := utils.ProcessTemplate(TmpltBridgedOVS, &BridgedOVSData{port.Name})
			if err != nil {
				return "", err
			}
			data += string(tempData) + "\n"

		case hwinfo.NicTypeBridge:
			buf, err := ioutil.ReadFile(filepath.Join(templatesDir, TmpltFileBridged))
			if err == nil {
				TmpltBridged = string(buf)
			}
			tempData, err := utils.ProcessTemplate(TmpltBridged, &BridgedData{port.Name})
			if err != nil {
				return "", err
			}
			data += string(tempData) + "\n"
		}
	}
	return data, nil
}

func ProcessTemplatePassthrough(pci string) ([]byte, error) {
	pciSlice := strings.Split(pci, ":")
	d := new(PassthroughData)
	d.Bus = pciSlice[1]
	temp := strings.Split(pciSlice[2], ".")
	d.Slot = temp[0]
	d.Function = temp[1]
	return utils.ProcessTemplate(TmpltPassthrough, d)
}

func bindingMode(port *hwinfo.NIC, nconf *xmlinput.Allow) string {
	if nconf.Model == "" && strings.Contains(port.Vendor, nconf.Vendor) ||
		strings.Contains(port.Model, nconf.Model) {
		return nconf.Mode
	}
	return ""
}
