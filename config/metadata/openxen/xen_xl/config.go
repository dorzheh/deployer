package xen_xl

// Intended for creating configuration related to those deployments
// where the target appliance assumed to be powered by OpenXen xl toolstack

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/dorzheh/deployer/builder/image"
	"github.com/dorzheh/deployer/config/metadata"
	"github.com/dorzheh/deployer/config/xmlinput"
	"github.com/dorzheh/deployer/controller"
	"github.com/dorzheh/deployer/deployer"
	envdriver "github.com/dorzheh/deployer/drivers/env_driver/openxen/xen_xl"
	hwinfodriver "github.com/dorzheh/deployer/drivers/hwinfo_driver/openxen"
	"github.com/dorzheh/deployer/utils"
	"github.com/dorzheh/deployer/utils/hwinfo/guest"
	"github.com/dorzheh/deployer/utils/hwinfo/host"
)

type meta struct{}

func CreateConfig(d *deployer.CommonData, i *metadata.InputData) (*metadata.Config, error) {
	if d.DefaultExportDir == "" {
		d.DefaultExportDir = "/var/lib/xen"
	}

	m, err := metadata.NewMetdataConfig(d, i.StorageConfigFile)
	if err != nil {
		return nil, utils.FormatError(err)
	}

	controller.RegisterSteps(func() func() error {
		return func() error {
			var err error
			m.Hwdriver, err = hwinfodriver.NewHostinfoDriver(m.SshConfig, i.Lshw, filepath.Join(d.RootDir, ".hwinfo.json"))
			if err != nil {
				return utils.FormatError(err)
			}

			m.Metadata = new(metadata.Metadata)
			m.EnvDriver = envdriver.NewDriver(m.SshConfig)
			return controller.SkipStep
		}
	}())

	if err := metadata.RegisterSteps(d, i, m, &meta{}); err != nil {
		return nil, utils.FormatError(err)
	}
	return m, nil
}

func (m meta) DefaultMetadata() []byte {
	return defaultMetdataPVHVM
}

func (m meta) SetCpuTuneData(*guest.Config, string) (string, error) {
	return "", nil
}

// --- metadata configuration: storage --- //

var blockDevicesSuffix = []string{"a", "b", "c", "d", "e", "f", "g", "h"}

// SetStorageData is responsible for adding to the metadata appropriate entries
// related to the storage configuration
func (m meta) SetStorageData(conf *guest.Config, templatesDir string) (string, error) {
	var e []string
	for i, disk := range conf.Storage.Disks {
		switch disk.Type {
		case image.StorageTypeQCOW2:
			e = append(e, "'tap:qcow2:"+disk.Path+",xvd"+blockDevicesSuffix[i]+",w'")
		case image.StorageTypeRAW:
			e = append(e, "'file:"+disk.Path+",xvd"+blockDevicesSuffix[i]+",w'")
		default:
			return "", errors.New("Unsupported Virtual Disk format")
		}
	}

	if len(e) == 0 {
		return "", errors.New("Virtual Disk configuration not found")
	}
	return "disk = [ " + strings.Join(e, ",") + " ]", nil
}

// --- metadata configuration: network --- //

// SetNetworkData is responsible for adding to the metadata appropriate entries
// related to the network configuration
func (m meta) SetNetworkData(mapping *guest.Config, templatesDir string) (string, error) {
	var e []string
	for i, network := range mapping.Networks {
		list := mapping.NICLists[i]
		for _, mode := range network.Modes {
			for _, port := range list {
				switch port.HostNIC.Type {
				case host.NicTypeOVS:
					if mode.Type == xmlinput.ConTypeOVS {
						if mode.VnicDriver != "" {
							e = append(e, fmt.Sprintf("'script=vif-openvswitch,bridge=%s,model=%s'", port.HostNIC.Name, mode.VnicDriver))
						} else {
							e = append(e, fmt.Sprintf("'script=vif-openvswitch,bridge=%s'", port.HostNIC.Name))
						}
					}

				case host.NicTypeBridge:
					if mode.Type == xmlinput.ConTypeBridged {
						if mode.VnicDriver != "" {
							e = append(e, fmt.Sprintf("'bridge=%s,model=%s'", port.HostNIC.Name, mode.VnicDriver))
						} else {
							e = append(e, fmt.Sprintf("'bridge=%s'", port.HostNIC.Name))
						}
					}
				}
			}
		}
	}

	data := "vif = ["
	if len(e) == 0 {
		data += "' ']"
	} else {
		data += strings.Join(e, ",") + "]"
	}
	return data, nil
}

func (m meta) SetCpuConfigData(*guest.Config, string) (string, error) {
	return "", nil
}

func (m meta) SetNUMATuneData(*guest.Config, string) (string, error) {
	return "", nil
}

func (m meta) SetCustomData(*guest.Config, string) (string, error) {
	return "", nil
}
