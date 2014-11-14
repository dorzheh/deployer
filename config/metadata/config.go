package metadata

import (
	//"errors"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/dorzheh/deployer/builder/common/image"
	"github.com/dorzheh/deployer/config/common"
	"github.com/dorzheh/deployer/config/common/bundle"
	"github.com/dorzheh/deployer/config/common/xmlinput"
	"github.com/dorzheh/deployer/deployer"
	gui "github.com/dorzheh/deployer/ui"
	"github.com/dorzheh/deployer/utils/hwinfo"
)

type Metaconfigurator interface {
	// storage configuration and templates directory
	// returns metadata entry related to storage and error
	SetStorageData(*image.Config, string) (string, error)

	// network interfaces information, allowed NICs, templates directory
	// returns metadata entry related to the network interfaces configuration and error
	SetNetworkData(map[string]*hwinfo.NIC, []*xmlinput.Allow, string) (string, error)

	// default metadata is used by the deployer in case user didn't provide any template
	// returns entry related to default metadata
	DefaultMetadata() []byte
}

// InputData provides a static data
type InputData struct {
	// Path to the lshw binary
	Lshw string

	// Path to the basic configuration file (see xmlinput package)
	InputDataConfigFile string

	// Path to the bundle configuration file (see bundle package)
	BundleDataConfigFile string

	// Path to the storage configuration file (see image package)
	StorageConfigFile string

	// Path to directory containing appropriate templates intended
	// for overriding default configuration
	TemplatesDir string
}

// commonMetadata contains elements that will processed
// by the template library and used by Libvirt XML metadata
type CommonMetadata struct {
	DomainName   string
	CPUs         uint
	RAM          uint
	EmulatorPath string
	Storage      string
	Networks     string
	CustomData   interface{}
}

// Config contains common configuration plus appropriate
// configuration required for appliances powered by environment
// based on libvirt API
type Config struct {
	// Common configuration
	*deployer.CommonConfig

	// Hostinfo driver
	Hwdriver deployer.HostinfoDriver

	// Common metadata stuff
	Metadata *CommonMetadata

	// Path to metadata file
	DestMetadataFile string

	// Storage configuration
	StorageConfig *image.Config

	// bundle configuration(in case predefined configuration is used)
	BundleConfig *bundle.Config
}

func CreateConfig(d *deployer.CommonData, i *InputData,
	c *Config, driver deployer.Driver, metaconf Metaconfigurator) (*Config, error) {

	var err error

	d.VaName = gui.UiApplianceName(d.Ui, d.VaName, driver)
	c.Metadata.DomainName = d.VaName
	c.DestMetadataFile = filepath.Join(c.ExportDir, d.VaName+"-metadata")
	// always create default metadata
	if err := ioutil.WriteFile(c.DestMetadataFile, metaconf.DefaultMetadata(), 0); err != nil {
		return nil, err
	}

	if fd, err := os.Stat(i.InputDataConfigFile); err != nil || fd.IsDir() {
		return nil, err
	}

	xid, err := xmlinput.ParseXML(i.InputDataConfigFile)
	if err != nil {
		return nil, err
	}
	if err := gui.UiGatherHWInfo(d.Ui, c.Hwdriver, "1s", c.RemoteMode); err != nil {
		return nil, err
	}

	var storageConfigIndex image.ConfigIndex = 0
	if fd, err := os.Stat(i.BundleDataConfigFile); err == nil && !fd.IsDir() {
		c.BundleConfig, err = bundle.GetConfig(d, c.Hwdriver, xid, i.BundleDataConfigFile)
		if err != nil {
			return nil, err
		}
		c.Metadata.CPUs = c.BundleConfig.CPUs
		c.Metadata.RAM = c.BundleConfig.RAM * 1024
		storageConfigIndex = c.BundleConfig.StorageConfigIndex
	} else {
		if xid.CPU.Config {
			cpus, err := c.Hwdriver.CPUs()
			if err != nil {
				return nil, err
			}
			c.Metadata.CPUs = gui.UiCPUs(d.Ui, cpus, xid.CPU.Min, xid.CPU.Max)
		} else if xid.CPU.Default > 0 {
			c.Metadata.CPUs = xid.CPU.Default
		}
		if xid.RAM.Config {
			ram, err := c.Hwdriver.RAMSize()
			if err != nil {
				return nil, err
			}
			c.Metadata.RAM = gui.UiRAMSize(d.Ui, ram, xid.RAM.Min, xid.RAM.Max)
			c.Metadata.RAM *= 1024
		} else if xid.RAM.Default > 0 {
			c.Metadata.RAM = xid.RAM.Default * 1024
		}
	}

	c.StorageConfig, err = common.StorageConfig(i.StorageConfigFile, filepath.Join(c.ExportDir, d.VaName), storageConfigIndex)
	c.Metadata.Storage, err = metaconf.SetStorageData(c.StorageConfig, i.TemplatesDir)
	if err != nil {
		return nil, err
	}

	if xid.Networks.Config {
		nets, err := gui.UiNetworks(d.Ui, xid, c.Hwdriver)
		if err != nil {
			return nil, err
		}

		c.Metadata.Networks, err = metaconf.SetNetworkData(nets, xid.Allowed, i.TemplatesDir)
		if err != nil {
			return nil, err
		}
	}
	return c, nil
}
