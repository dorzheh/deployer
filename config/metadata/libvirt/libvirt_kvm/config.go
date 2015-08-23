// Intended for creating configuration related to those deployments
// where the target appliance assumed to be powered by libvirt API
package libvirt_kvm

import (
	"io/ioutil"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/dorzheh/deployer/builder/image"
	"github.com/dorzheh/deployer/config/common"
	"github.com/dorzheh/deployer/config/common/xmlinput"
	"github.com/dorzheh/deployer/config/metadata"
	"github.com/dorzheh/deployer/controller"
	"github.com/dorzheh/deployer/deployer"
	envdriver "github.com/dorzheh/deployer/drivers/env_driver/libvirt/libvirt_kvm"
	hwinfodriver "github.com/dorzheh/deployer/drivers/hwinfo_driver/libvirt"
	"github.com/dorzheh/deployer/utils"
	"github.com/dorzheh/deployer/utils/hwinfo/guest"
	"github.com/dorzheh/deployer/utils/hwinfo/host"
)

type meta struct{}

func CreateConfig(d *deployer.CommonData, i *metadata.InputData) (*metadata.Config, error) {
	d.DefaultExportDir = "/var/lib/libvirt/images"
	c := common.RegisterSteps(d)
	m := &metadata.Config{c, nil, nil, nil, nil, "", nil, nil}
	controller.RegisterSteps(func() func() error {
		return func() error {
			var err error
			m.Hwdriver, err = hwinfodriver.NewHostinfoDriver(m.SshConfig, i.Lshw, filepath.Join(d.RootDir, ".hwinfo.json"))
			if err != nil {
				return utils.FormatError(err)
			}

			m.Metadata = new(metadata.Metadata)
			m.EnvDriver = envdriver.NewDriver(m.SshConfig)
			if m.Metadata.EmulatorPath, err = m.EnvDriver.Emulator(d.Arch); err != nil {
				return utils.FormatError(err)
			}
			return controller.SkipStep
		}
	}())

	if err := metadata.RegisterSteps(d, i, m, &meta{}); err != nil {
		return nil, utils.FormatError(err)
	}
	return m, nil
}

func (m meta) DefaultMetadata() []byte {
	return defaultMetdata
}

// --- metadata configuration: cpu tuning --- //
const (
	TmpltFileCpuTune = "template_cpu_tuning.xml"
)

type CpuTuneData struct {
	CpuTuneData string
}

func (m meta) SetCpuTuneData(cpus map[int][]string, templatesDir string) (string, error) {
	// assume <cputune> </cputune> exists in metadata
	buf, err := ioutil.ReadFile(filepath.Join(templatesDir, TmpltFileCpuTune))
	if err == nil {
		TmpltCpuTune = string(buf)
	}

	tempData, err := utils.ProcessTemplate(TmpltCpuTune, &CpuTuneData{SetCpuTuneData(cpus)})
	if err != nil {
		return "", utils.FormatError(err)
	}
	return string(tempData) + "\n", nil
}

func SetCpuTuneData(cpus map[int][]string) string {
	var cpuTuneData string
	for vcpu, pcpus := range cpus {
		var cpuset string
		for i, pcpu := range pcpus {
			if i == 0 {
				cpuset = pcpu
			} else {
				cpuset += "," + pcpu
			}
		}
		cpuTuneData += "\n" + `<vcpupin vcpu='` + strconv.Itoa(vcpu) + `' cpuset='` + cpuset + `'/>`
	}
	return cpuTuneData
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
	HostNicBus       string
	HostNicSlot      string
	HostNicFunction  string
	GuestNicDomain   string
	GuestNicBus      string
	GuestNicSlot     string
	GuestNicFunction string
}

type BridgedOVSData struct {
	OVSBridge        string
	Driver           string
	GuestNicDomain   string
	GuestNicBus      string
	GuestNicSlot     string
	GuestNicFunction string
}

type BridgedData struct {
	Bridge           string
	Driver           string
	GuestNicDomain   string
	GuestNicBus      string
	GuestNicSlot     string
	GuestNicFunction string
}

type DirectData struct {
	IfaceName        string
	Driver           string
	GuestNicDomain   string
	GuestNicBus      string
	GuestNicSlot     string
	GuestNicFunction string
}

type VirtNetwork struct {
	NetworkName      string
	Driver           string
	GuestNicDomain   string
	GuestNicBus      string
	GuestNicSlot     string
	GuestNicFunction string
}

// SetNetworkData is responsible for adding to the metadata appropriate entries
// related to the network configuration
func (m meta) SetNetworkData(mapping *deployer.OutputNetworkData, templatesDir string) (string, error) {
	var data string
	for i, network := range mapping.Networks {
		list := mapping.NICLists[i]
		for _, mode := range network.Modes {
			for _, port := range list {
				switch port.HostNIC.Type {
				case host.NicTypePhys:
					if mode.Type == xmlinput.ConTypePassthrough || mode.Type == xmlinput.ConTypeDirect {
						out, err := treatPhysical(port, mode, templatesDir)
						if err != nil {
							return "", utils.FormatError(err)
						}
						data += out
					}

				case host.NicTypePhysVF:
					if mode.Type == xmlinput.ConTypeSRIOV {
						out, err := treatPhysical(port, mode, templatesDir)
						if err != nil {
							return "", utils.FormatError(err)
						}
						data += out
					}

				case host.NicTypeOVS:
					if mode.Type == xmlinput.ConTypeOVS {
						tempData, err := metadata.ProcessNetworkTemplate(mode, TmpltBridgedOVS,
							&BridgedOVSData{port.HostNIC.Name, mode.VnicDriver, port.PCIAddr.Domain,
								port.PCIAddr.Bus, port.PCIAddr.Slot, port.PCIAddr.Function}, templatesDir)
						if err != nil {
							return "", utils.FormatError(err)
						}
						data += tempData
					}

				case host.NicTypeBridge:
					if mode.Type == xmlinput.ConTypeBridged {
						tempData, err := metadata.ProcessNetworkTemplate(mode, TmpltBridged,
							&BridgedData{port.HostNIC.Name, mode.VnicDriver, port.PCIAddr.Domain,
								port.PCIAddr.Bus, port.PCIAddr.Slot, port.PCIAddr.Function}, templatesDir)
						if err != nil {
							return "", utils.FormatError(err)
						}
						data += tempData
					}

				case host.NicTypeVirtualNetwork:
					if mode.Type == xmlinput.ConTypeVirtualNetwork {
						tempData, err := metadata.ProcessNetworkTemplate(mode, TmpltVirtNetwork,
							&VirtNetwork{port.HostNIC.Name, mode.VnicDriver, port.PCIAddr.Domain,
								port.PCIAddr.Bus, port.PCIAddr.Slot, port.PCIAddr.Function}, templatesDir)
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

func ProcessTemplatePassthrough(port *guest.NIC, tmplt string) (string, error) {
	pciSlice := strings.Split(port.HostNIC.PCIAddr, ":")
	d := new(PassthroughData)
	d.HostNicBus = pciSlice[1]
	temp := strings.Split(pciSlice[2], ".")
	d.HostNicSlot = temp[0]
	d.HostNicFunction = temp[1]
	d.GuestNicDomain = port.PCIAddr.Domain
	d.GuestNicBus = port.PCIAddr.Bus
	d.GuestNicSlot = port.PCIAddr.Slot
	d.GuestNicFunction = port.PCIAddr.Function
	data, err := utils.ProcessTemplate(tmplt, d)
	if err != nil {
		return "", utils.FormatError(err)
	}
	return string(data), nil
}

func treatPhysical(port *guest.NIC, mode *xmlinput.Mode, templatesDir string) (string, error) {
	var err error
	var tempData string

	switch mode.Type {
	case xmlinput.ConTypePassthrough:
		if tempData, err = ProcessTemplatePassthrough(port, TmpltPassthrough); err != nil {
			return "", utils.FormatError(err)
		}
	case xmlinput.ConTypeSRIOV:
		if tempData, err = ProcessTemplatePassthrough(port, TmpltSriovPassthrough); err != nil {
			return "", utils.FormatError(err)
		}
	case xmlinput.ConTypeDirect:
		if tempData, err = metadata.ProcessNetworkTemplate(mode, TmpltDirect,
			&DirectData{port.HostNIC.Name, mode.VnicDriver, port.PCIAddr.Domain, port.PCIAddr.Bus,
				port.PCIAddr.Slot, port.PCIAddr.Function}, templatesDir); err != nil {
			return "", utils.FormatError(err)
		}
	}
	return tempData, nil
}
