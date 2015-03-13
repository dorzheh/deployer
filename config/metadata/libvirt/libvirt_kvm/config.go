// Intended for creating configuration related to those deployments
// where the target appliance assumed to be powered by libvirt API
package libvirt_kvm

import (
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/dorzheh/deployer/builder/image"
	"github.com/dorzheh/deployer/config/common"
	"github.com/dorzheh/deployer/config/common/xmlinput"
	"github.com/dorzheh/deployer/config/metadata"
	"github.com/dorzheh/deployer/deployer"
	envdriver "github.com/dorzheh/deployer/drivers/env_driver/libvirt/libvirt_kvm"
	hwinfodriver "github.com/dorzheh/deployer/drivers/hwinfo_driver/libvirt"
	"github.com/dorzheh/deployer/utils"
	"github.com/dorzheh/deployer/utils/hwinfo"
)

type meta struct{}

func CreateConfig(d *deployer.CommonData, i *metadata.InputData) (*metadata.Config, error) {
	var err error
	d.DefaultExportDir = "/var/lib/libvirt/images"
	c := &metadata.Config{common.CreateConfig(d), nil, nil, nil, "", nil}
	c.Hwdriver, err = hwinfodriver.NewHostinfoDriver(filepath.Join(d.RootDir, ".hwinfo.json"), i.Lshw, c.SshConfig)
	if err != nil {
		return nil, utils.FormatError(err)
	}

	c.Metadata = new(metadata.Metadata)
	drvr := envdriver.NewDriver(c.SshConfig)
	if c.Metadata.EmulatorPath, err = drvr.Emulator(d.Arch); err != nil {
		return nil, utils.FormatError(err)
	}
	return metadata.CreateConfig(d, i, c, drvr, &meta{})
}

func (m meta) DefaultMetadata() []byte {
	return defaultMetdata
}

// --- metadata configuration: storage --- //
const (
	TmpltFileStorage = "template_storage.xml"
)

type DiskData struct {
	ImagePath         string
	StorageType       image.StorageType
	BlockDeviceSuffix string
}

var blockDevicesSuffix = []string{"a", "b", "c", "d", "e", "f", "g", "h"}

// SetStorageData is responsible for adding to the metadata appropriate entries
// related to the storage configuration
func (m meta) SetStorageData(conf *image.Config, templatesDir string) (string, error) {
	var data string

	buf, err := ioutil.ReadFile(filepath.Join(templatesDir, TmpltFileStorage))
	if err == nil {
		TmpltStorage = string(buf)
	}

	for i, disk := range conf.Disks {
		d := new(DiskData)
		d.ImagePath = disk.Path
		d.StorageType = disk.Type
		d.BlockDeviceSuffix = blockDevicesSuffix[i]
		tempData, err := utils.ProcessTemplate(TmpltStorage, d)
		if err != nil {
			return "", utils.FormatError(err)
		}
		data += string(tempData) + "\n"
	}

	return data, nil
}

// --- metadata configuration: network --- //

type PassthroughData struct {
	Bus      string
	Slot     string
	Function string
}

type BridgedOVSData struct {
	OVSBridge string
	Driver    string
}

type BridgedData struct {
	Bridge string
	Driver string
}

type DirectData struct {
	IfaceName string
	Driver    string
}

type VirtNetwork struct {
	NetworkName string
	Driver      string
}

// SetNetworkData is responsible for adding to the metadata appropriate entries
// related to the network configuration
func (m meta) SetNetworkData(mapping *deployer.OutputNetworkData, templatesDir string) (string, error) {
	var data string
	for i, network := range mapping.Networks {
		list := mapping.NICLists[i]
		for _, mode := range network.Modes {
			for _, port := range list {
				switch port.Type {
				case hwinfo.NicTypePhys:
					if mode.Type == xmlinput.ConTypePassthrough || mode.Type == xmlinput.ConTypeDirect {
						out, err := treatPhysical(port, mode, templatesDir)
						if err != nil {
							return "", utils.FormatError(err)
						}
						data += out
					}

				case hwinfo.NicTypeOVS:
					if mode.Type == xmlinput.ConTypeOVS {
						tempData, err := metadata.ProcessNetworkTemplate(mode, TmpltBridgedOVS,
							&BridgedOVSData{port.Name, mode.VnicDriver}, templatesDir)
						if err != nil {
							return "", utils.FormatError(err)
						}
						data += tempData
					}

				case hwinfo.NicTypeBridge:
					if mode.Type == xmlinput.ConTypeBridged {
						tempData, err := metadata.ProcessNetworkTemplate(mode, TmpltBridged,
							&BridgedData{port.Name, mode.VnicDriver}, templatesDir)
						if err != nil {
							return "", utils.FormatError(err)
						}
						data += tempData
					}

				case hwinfo.NicTypeVirtualNetwork:
					if mode.Type == xmlinput.ConTypeVirtualNetwork {
						tempData, err := metadata.ProcessNetworkTemplate(mode, TmpltVirtNetwork,
							&VirtNetwork{port.Name, mode.VnicDriver}, templatesDir)
						if err != nil {
							return "", utils.FormatError(err)
						}
						data += tempData
					}
				}
			}
		}
	}
	return data, nil
}

func ProcessTemplatePassthrough(pci string) (string, error) {
	pciSlice := strings.Split(pci, ":")
	d := new(PassthroughData)
	d.Bus = pciSlice[1]
	temp := strings.Split(pciSlice[2], ".")
	d.Slot = temp[0]
	d.Function = temp[1]
	data, err := utils.ProcessTemplate(TmpltPassthrough, d)
	if err != nil {
		return "", utils.FormatError(err)
	}
	return string(data), nil
}

func treatPhysical(port *hwinfo.NIC, mode *xmlinput.Mode, templatesDir string) (string, error) {
	var err error
	var tempData string

	switch mode.Type {
	case xmlinput.ConTypePassthrough:
		if tempData, err = ProcessTemplatePassthrough(port.PCIAddr); err != nil {
			return "", utils.FormatError(err)
		}
	case xmlinput.ConTypeDirect:
		if tempData, err = metadata.ProcessNetworkTemplate(mode, TmpltDirect,
			&DirectData{port.Name, mode.VnicDriver}, templatesDir); err != nil {
			return "", utils.FormatError(err)
		}
	}
	return tempData, nil
}
