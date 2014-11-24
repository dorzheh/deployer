// Intended for creating configuration related to those deployments
// where the target appliance assumed to be powered by libvirt API
package libvirt

import (
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/dorzheh/deployer/builder/common/image"
	"github.com/dorzheh/deployer/config/common"
	"github.com/dorzheh/deployer/config/common/xmlinput"
	"github.com/dorzheh/deployer/config/metadata"
	"github.com/dorzheh/deployer/deployer"
	hwinfodrvr "github.com/dorzheh/deployer/hwinfo_driver/libvirt"
	"github.com/dorzheh/deployer/post_processor/libvirt"
	"github.com/dorzheh/deployer/utils"
	"github.com/dorzheh/deployer/utils/hwinfo"
)

type meta struct{}

func CreateConfig(d *deployer.CommonData, i *metadata.InputData) (*metadata.Config, error) {
	var err error
	d.DefaultExportDir = "/var/lib/libvirt/images"
	c := &metadata.Config{common.CreateConfig(d), nil, nil, nil, "", nil}
	c.Hwdriver, err = hwinfodrvr.NewHostinfoDriver(filepath.Join(d.RootDir, "hwinfo.json"), i.Lshw, c.SshConfig)
	if err != nil {
		return nil, err
	}

	c.Metadata = new(metadata.CommonMetadata)
	postdriver := libvirt.NewDriver(c.SshConfig)
	if c.Metadata.EmulatorPath, err = postdriver.Emulator(d.Arch); err != nil {
		return nil, err
	}
	return metadata.CreateConfig(d, i, c, postdriver, &meta{})
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
		d.BlockDeviceSuffix = blockDevicesSuffix[i]
		tempData, err := utils.ProcessTemplate(TmpltStorage, d)
		if err != nil {
			return "", err
		}
		data += string(tempData) + "\n"
	}

	return data, nil
}

// --- metadata configuration: network --- //

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

// SetNetworkData is responsible for adding to the metadata appropriate entries
// related to the network configuration
func (m meta) SetNetworkData(mapping map[*xmlinput.Network]hwinfo.NICList, templatesDir string) (string, error) {
	var data string
	for network, list := range mapping {
		for _, mode := range network.Modes {
			for _, port := range list {
				switch port.Type {
				case hwinfo.NicTypePhys:
					if mode.Type == xmlinput.ConTypePassthrough || mode.Type == xmlinput.ConTypeDirect {
						out, err := treatPhysical(port, mode, templatesDir)
						if err != nil {
							return "", err
						}
						data += out
					}

				case hwinfo.NicTypeOVS:
					if mode.Type == xmlinput.ConTypeBridged {
						buf, err := ioutil.ReadFile(filepath.Join(templatesDir, TmpltFileBridgedOVS))
						if err == nil {
							TmpltBridgedOVS = string(buf)
						}
						tempData, err := utils.ProcessTemplate(TmpltBridgedOVS, &BridgedOVSData{port.Name, mode.VnicDriver})
						if err != nil {
							return "", err
						}
						data += string(tempData) + "\n"
					}

				case hwinfo.NicTypeBridge:
					if mode.Type == xmlinput.ConTypeBridged {
						buf, err := ioutil.ReadFile(filepath.Join(templatesDir, TmpltFileBridged))
						if err == nil {
							TmpltBridged = string(buf)
						}
						tempData, err := utils.ProcessTemplate(TmpltBridged, &BridgedData{port.Name, mode.VnicDriver})
						if err != nil {
							return "", err
						}
						data += string(tempData) + "\n"
					}
				}
			}
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

func treatPhysical(port *hwinfo.NIC, mode *xmlinput.Mode, templatesDir string) (string, error) {
	var err error
	var tempData []byte

	switch mode.Type {
	case xmlinput.ConTypePassthrough:
		tempData, err = ProcessTemplatePassthrough(port.PCIAddr)
		if err != nil {
			return "", err
		}
	case xmlinput.ConTypeDirect:
		buf, err := ioutil.ReadFile(filepath.Join(templatesDir, TmpltFileBridgedOVS))
		if err == nil {
			TmpltDirect = string(buf)
		}
		tempData, err = utils.ProcessTemplate(TmpltDirect, &DirectData{port.Name, mode.VnicDriver})
		if err != nil {
			return "", err
		}
	}
	return string(tempData) + "\n", nil
}
