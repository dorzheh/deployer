// Intended for creating configuration related to those deployments
// where the target appliance assumed to be powered by libvirt API
package libvirt

import (
	"path/filepath"

	"github.com/dorzheh/deployer/config/common"
	"github.com/dorzheh/deployer/deployer"
	"github.com/dorzheh/deployer/post_processor/libvirt"
	gui "github.com/dorzheh/deployer/ui"
	"github.com/dorzheh/deployer/utils"
)

// InputData provides a static data
type InputData struct {
	// Networks contains a slice of networks
	// that the target appliance will be bound to
	Networks []string

        // Supported NIC vendors
	NICVendors []string
	
	// Path to the lshw binary
	LshwPath string
}

// CommonMetadata contains elements that will processed
// by the template library and used by Libvirt XML metadata
type CommonMetadata struct {
	DomainName   string
	EmulatorPath string
	ImagePath    string
	Networks     string
}

// Config contains common configuration plus appropriate
// configuration required for appliances powered by environment
// based on libvirt API
type Config struct {
	Common       *deployer.CommonConfig
	Networks     map[string]*utils.NicInfo
	HwInfo       *utils.HwInfoParser
	MetadataPath string
	Data         *CommonMetadata
}

func CreateConfig(d *deployer.CommonData, i *InputData) (*Config, error) {
	var err error
	d.DefaultExportDir = "/var/lib/libvirt/images"

	c := new(Config)
	c.Common = common.CreateConfig(d)
	driver := libvirt.NewDriver(c.Common.SshConfig)
	c.Data = new(CommonMetadata)
	if c.Data.EmulatorPath, err = driver.Emulator(); err != nil {
		return nil, err
	}
	c.HwInfo, err = utils.NewHwInfoParser(filepath.Join(d.RootDir, "hwinfo.json"), i.LshwPath, c.Common.SshConfig)

	errCh := make(chan error)
	defer close(errCh)
	go func() {
		errCh <- c.HwInfo.Parse()
	}()

	d.VaName = gui.UiApplianceName(d.Ui, d.VaName, driver)
	c.Data.DomainName = d.VaName
	c.Data.ImagePath = filepath.Join(c.Common.ExportDir, c.Data.DomainName)
	c.MetadataPath = filepath.Join(c.Common.ExportDir, c.Data.DomainName+".xml")

	if err = utils.WaitForResult(errCh, 1); err != nil {
		return nil, err
	}

	// Sometimes more complex network configuration is needed.
	// In this case -  pass empty slice and overwrite appropriate
	// logic at a higher implementation level
	if len(i.Networks) > 0 {
		ni, err := c.HwInfo.NicsInfo(NicVendors)
		if err != nil {
			return nil, err
		}

		c.Networks, err = gui.UiNetworks(d.Ui, i.Networks, ni)
		if err != nil {
			return nil, err
		}

		c.Data.Networks, err = SetNetworkData(c.Networks)
		if err != nil {
			return nil, err
		}
	}
	return c, nil
}
