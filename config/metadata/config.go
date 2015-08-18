package metadata

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/dorzheh/deployer/builder/image"
	"github.com/dorzheh/deployer/config/common"
	"github.com/dorzheh/deployer/config/common/bundle"
	"github.com/dorzheh/deployer/config/common/xmlinput"
	"github.com/dorzheh/deployer/deployer"
	gui "github.com/dorzheh/deployer/ui"
	"github.com/dorzheh/deployer/utils"
	"github.com/dorzheh/deployer/utils/host_hwfilter"
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
	CPUTune      string
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

	// Network data
	NetData *deployer.OutputNetworkData

	// Bundle config
	Bundle map[string]interface{}
}

func CreateConfig(d *deployer.CommonData, i *InputData, c *Config, driver deployer.EnvDriver, metaconf deployer.MetadataConfigurator) (*Config, error) {
	xid, err := xmlinput.ParseXMLInput(i.InputDataConfigFile)
	if err != nil {
		return nil, utils.FormatError(err)
	}

	c.Ctrl.RegisterSteps(func(*deployer.CommonData, *Config, deployer.EnvDriver) func() error {
		return func() error {
			var err error
			if c.Metadata.DomainName, err = gui.UiApplianceName(d.Ui, d.VaName, driver); err != nil {
				return err
			}
			d.VaName = c.Metadata.DomainName
			if err := gui.UiGatherHWInfo(d.Ui, c.Hwdriver, "1s", c.RemoteMode); err != nil {
				return utils.FormatError(err)
			}
			return nil
		}
	}(d, c, driver))

	// Network configuration
	if xid.Networks.Configure {
		c.Ctrl.RegisterSteps(func(*deployer.CommonData, *InputData, *Config, *xmlinput.XMLInputData) func() error {
			return func() error {
				nics, err := host_hwfilter.GetAllowedNICs(xid, c.Hwdriver)
				if err != nil {
					return utils.FormatError(err)
				}
				c.NetData, err = gui.UiNetworks(d.Ui, xid, nics)
				if err != nil {
					return err
				}
				c.Metadata.Networks, err = metaconf.SetNetworkData(c.NetData, i.TemplatesDir)
				if err != nil {
					return utils.FormatError(err)
				}
				return nil
			}

		}(d, i, c, xid))
	}

	conf := new(gui.VmConfig)
	conf.DisksMb = make([]int, 0)
	var storageConfigIndex image.ConfigIndex = 0

	c.Ctrl.RegisterSteps(func(*deployer.CommonData, *InputData, *Config, deployer.EnvDriver, *xmlinput.XMLInputData) func() error {
		return func() error {
			if i.BundleParser != nil {
				m, err := i.BundleParser.Parse(d, c.Hwdriver, xid)
				if err != nil {
					return utils.FormatError(err)
				}
				if m != nil {
					c.Bundle = m
					c.Metadata.CPUs = m["cpus"].(int)
					c.Metadata.RAM = m["ram_mb"].(int) * 1024
					storageConfigIndex = m["storage_config_index"].(image.ConfigIndex)
				}
			}
			if len(c.Bundle) == 0 {
				if xid.CPU.Max == xmlinput.UnlimitedAlloc {
					xid.CPU.Max = driver.MaxVCPUsPerGuest()
				}
				if conf, err = gui.UiVmConfig(d.Ui, c.Hwdriver, xid); err != nil {
					return err
				}
				c.Metadata.CPUs = conf.CPUs
				c.Metadata.RAM = conf.RamMb * 1024
			}
			imagePath := filepath.Join(c.ExportDir, d.VaName)
			c.StorageConfig, err = common.StorageConfig(i.StorageConfigFile, imagePath, storageConfigIndex, conf.DisksMb)
			if err != nil {
				return utils.FormatError(err)
			}

			c.Metadata.Storage, err = metaconf.SetStorageData(c.StorageConfig, i.TemplatesDir)
			if err != nil {
				return utils.FormatError(err)
			}

			c.DestMetadataFile = fmt.Sprintf("/tmp/%s-temp-metadata.%d", d.VaName, os.Getpid())
			// always create default metadata
			if err := ioutil.WriteFile(c.DestMetadataFile, metaconf.DefaultMetadata(), 0); err != nil {
				return utils.FormatError(err)
			}
			return nil
		}
	}(d, i, c, driver, xid))

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
