package metadata

import (
	"io/ioutil"
	"path/filepath"

	"github.com/dorzheh/deployer/builder/common/image"
	"github.com/dorzheh/deployer/config/common"
	"github.com/dorzheh/deployer/config/common/bundle"
	"github.com/dorzheh/deployer/config/common/xmlinput"
	"github.com/dorzheh/deployer/deployer"
	gui "github.com/dorzheh/deployer/ui"
	"github.com/dorzheh/deployer/utils"
	"github.com/dorzheh/deployer/utils/hwinfo"
)

type MetadataConfigurator interface {
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

	// Bundle parser
	BundleParser *bundle.Parser

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

	// Storage configuration
	StorageConfig *image.Config

	// Common metadata stuff
	Metadata *CommonMetadata

	// Path to metadata file
	DestMetadataFile string

	// Bundle config
	Bundle map[string]interface{}
}

func CreateConfig(d *deployer.CommonData, i *InputData,
	c *Config, driver deployer.Driver, metaconf MetadataConfigurator) (*Config, error) {

	var err error

	d.VaName = gui.UiApplianceName(d.Ui, d.VaName, driver)
	c.Metadata.DomainName = d.VaName
	c.DestMetadataFile = filepath.Join(c.ExportDir, d.VaName+"-metadata")
	// always create default metadata
	if err := ioutil.WriteFile(c.DestMetadataFile, metaconf.DefaultMetadata(), 0); err != nil {
		return nil, err
	}

	xmlout, err := utils.ParseXMLFile(i.InputDataConfigFile, new(xmlinput.XMLInputData))
	if err != nil {
		return nil, err
	}

	xid := xmlout.(*xmlinput.XMLInputData)
	if err := gui.UiGatherHWInfo(d.Ui, c.Hwdriver, "1s", c.RemoteMode); err != nil {
		return nil, err
	}

	var storageConfigIndex image.ConfigIndex = 0
	var m map[string]interface{}
	if i.BundleParser != nil {
		m, err = i.BundleParser.Parse(d, c.Hwdriver, xid)
		if err != nil {
			return nil, err
		}
		if m != nil {
			c.Bundle = m
			c.Metadata.CPUs = m["cpus"].(uint)
			c.Metadata.RAM = m["ram"].(uint) * 1024
			storageConfigIndex = m["storage_config_index"].(image.ConfigIndex)
		}
	}
	if m == nil {
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

	// imagePath is a path to the main disk of the appliance.In case the appliance needs more than a single disk,
	// appropriate suffix will be added to each disk.For example, if path to the main disk is /mypath/disk
	// and the guest will be equipped with 3 disks , upcoming disks will be /mypath/disk_1 and /mypath/disk_2
	imagePath := filepath.Join(c.ExportDir, d.VaName)
	c.StorageConfig, err = common.StorageConfig(i.StorageConfigFile, imagePath, storageConfigIndex)
	if err != nil {
		return nil, err
	}

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
