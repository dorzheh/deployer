// Intended for creating configuration related to those deployments
// where the target appliance assumed to be powered by libvirt API
package libvirt

import (
	"path/filepath"

	"github.com/dorzheh/deployer/config/common"
	"github.com/dorzheh/deployer/deployer"
	"github.com/dorzheh/deployer/post_processor/libvirt"
	gui "github.com/dorzheh/deployer/ui"
	"github.com/dorzheh/deployer/utils/hwinfo"
)

// InputData provides a static data
type InputData struct {
	// Path to the lshw binary
	LshwPath string

	// Indicates if Ui function allowing CPU configuration will be called
	ConfigCPUs bool

	// Minimum vCPU requirement for VM
	MinCPUs uint

	// Indicates if Ui function allowing RAM configuration will be called
	ConfigRAM bool

	// Minimum RAM (in Mb)requirement for VM
	MinRAM uint

	// Indicates if Ui function allowing network configuration will be called
	ConfigNet bool

	// Networks contains a slice of networks
	// that the target appliance will be bound to
	Networks []string

	// Supported NIC vendors
	NICVendors []string
}

// commonMetadata contains elements that will processed
// by the template library and used by Libvirt XML metadata
type commonMetadata struct {
	DomainName   string
	CPUs         uint
	RAM          uint
	EmulatorPath string
	ImagePath    string
	Networks     string
}

// Config contains common configuration plus appropriate
// configuration required for appliances powered by environment
// based on libvirt API
type Config struct {
	// Common configuration
	*deployer.CommonConfig

	// Common metadata stuff
	Metadata *commonMetadata

	// Path to metadata file (libvirt XML)
	MetadataFile string

	// HWinfor parser
	HWInfo *hwinfo.Parser
}

func CreateConfig(d *deployer.CommonData, i *InputData) (*Config, error) {
	var err error
	d.DefaultExportDir = "/var/lib/libvirt/images"

	c := &Config{common.CreateConfig(d), nil, "", nil}
	c.Metadata = new(commonMetadata)

	driver := libvirt.NewDriver(c.SshConfig)
	if c.Metadata.EmulatorPath, err = driver.Emulator(); err != nil {
		return nil, err
	}

	d.VaName = gui.UiApplianceName(d.Ui, d.VaName, driver)
	c.Metadata.DomainName = d.VaName
	c.Metadata.ImagePath = filepath.Join(c.ExportDir, d.VaName)
	c.MetadataFile = filepath.Join(c.ExportDir, d.VaName+".xml")

	c.HWInfo, err = hwinfo.NewParser(filepath.Join(d.RootDir, "hwinfo.json"), i.LshwPath, c.SshConfig)
	if err := gui.UiGatherHWInfo(d.Ui, c.HWInfo, "1s", c.RemoteMode); err != nil {
		return nil, err
	}
	if i.ConfigCPUs {
		cpus, err := c.HWInfo.CPUs()
		if err != nil {
			return nil, err
		}
		c.Metadata.CPUs = gui.UiCPUs(d.Ui, cpus, i.MinCPUs)
	} else if i.MinCPUs > 0 {
		c.Metadata.CPUs = i.MinCPUs
	}
	if i.ConfigRAM {
		ram, err := c.HWInfo.RAMSize()
		if err != nil {
			return nil, err
		}
		c.Metadata.RAM = gui.UiRAMSize(d.Ui, ram, i.MinRAM)
	} else if i.MinRAM > 0 {
		c.Metadata.RAM = i.MinRAM
	}
	// Sometimes more complex network configuration is needed.
	// In this case set ConfigNet to false
	if i.ConfigNet {
		ni, err := c.HWInfo.NICInfo(i.NICVendors)
		if err != nil {
			return nil, err
		}

		nets, err := gui.UiNetworks(d.Ui, ni, i.Networks[0:]...)
		if err != nil {
			return nil, err
		}

		c.Metadata.Networks, err = SetNetworkData(nets)
		if err != nil {
			return nil, err
		}
	}
	return c, nil
}
