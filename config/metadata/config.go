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
	"github.com/dorzheh/deployer/utils/hwfilter"
)

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

// Metadata contains elements that will processed
// by the template library and used by Libvirt XML metadata
type Metadata struct {
	DomainName   string
	CPUs         int
	RAM          int
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
	Metadata *Metadata

	// Path to metadata file
	DestMetadataFile string

	// Bundle config
	Bundle map[string]interface{}
}

func CreateConfig(d *deployer.CommonData, i *InputData, c *Config, driver deployer.Driver, metaconf deployer.MetadataConfigurator) (*Config, error) {
	var err error
	d.VaName = gui.UiApplianceName(d.Ui, d.VaName, driver)
	c.Metadata.DomainName = d.VaName
	c.DestMetadataFile = filepath.Join(c.ExportDir, d.VaName+"-metadata")
	// always create default metadata
	if err := ioutil.WriteFile(c.DestMetadataFile, metaconf.DefaultMetadata(), 0); err != nil {
		return nil, utils.FormatError(err)
	}

	xid, err := xmlinput.ParseXMLInput(i.InputDataConfigFile)
	if err != nil {
		return nil, utils.FormatError(err)
	}
	if err := gui.UiGatherHWInfo(d.Ui, c.Hwdriver, "1s", c.RemoteMode); err != nil {
		return nil, utils.FormatError(err)
	}

	var storageConfigIndex image.ConfigIndex = 0
	var m map[string]interface{}
	if i.BundleParser != nil {
		m, err = i.BundleParser.Parse(d, c.Hwdriver, xid)
		if err != nil {
			return nil, utils.FormatError(err)
		}
		if m != nil {
			c.Bundle = m
			c.Metadata.CPUs = m["cpus"].(int)
			c.Metadata.RAM = m["ram_mb"].(int) * 1024
			storageConfigIndex = m["storage_config_index"].(image.ConfigIndex)
		}
	}

	var conf *gui.VmConfig
	if m == nil {
		conf, err = gui.UiVmConfig(d.Ui, c.Hwdriver, xid)
		if err != nil {
			return nil, utils.FormatError(err)
		}
		c.Metadata.CPUs = conf.CPUs
		c.Metadata.RAM = conf.RamMb * 1024
	}

	// imagePath is a path to the main disk of the appliance.In case the appliance needs more than a single disk,
	// appropriate suffix will be added to each disk.For example, if path to the main disk is /mypath/disk
	// and the guest will be equipped with 3 disks , upcoming disks will be /mypath/disk_1 and /mypath/disk_2
	imagePath := filepath.Join(c.ExportDir, d.VaName)
	c.StorageConfig, err = common.StorageConfig(i.StorageConfigFile, imagePath, storageConfigIndex, conf.DisksMb)
	if err != nil {
		return nil, utils.FormatError(err)
	}

	c.Metadata.Storage, err = metaconf.SetStorageData(c.StorageConfig, i.TemplatesDir)
	if err != nil {
		return nil, utils.FormatError(err)
	}
	if xid.Networks.Configure {
		nics, err := hwfilter.GetAllowedNICs(xid, c.Hwdriver)
		if err != nil {
			return nil, utils.FormatError(err)
		}
		nets, err := gui.UiNetworks(d.Ui, xid, nics)
		if err != nil {
			return nil, utils.FormatError(err)
		}

		c.Metadata.Networks, err = metaconf.SetNetworkData(nets, i.TemplatesDir)
		if err != nil {
			return nil, utils.FormatError(err)
		}
	}
	return c, nil
}

func ProcessNetworkTemplate(mode *xmlinput.Mode, defaultTemplate string, tmpltData interface{}, templatesDir string) (string, error) {
	var customTemplate string

	if mode.Tmplt == nil {
		customTemplate = defaultTemplate
	} else {
		var templatePath string
		if templatesDir != "" {
			templatePath = filepath.Join(templatesDir, mode.Tmplt.FileName)
		} else {
			templatePath = filepath.Join(mode.Tmplt.Dir, mode.Tmplt.FileName)
		}

		buf, err := ioutil.ReadFile(templatePath)
		if err != nil {
			return "", utils.FormatError(err)
		}
		customTemplate = string(buf)
	}

	tempData, err := utils.ProcessTemplate(customTemplate, tmpltData)
	if err != nil {
		return "", utils.FormatError(err)
	}
	return string(tempData) + "\n", nil
}
