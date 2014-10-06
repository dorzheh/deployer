// Intended for creating configuration related to those deployments
// where the target appliance assumed to be powered by libvirt API
package libvirt

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/dorzheh/deployer/config/common"
	"github.com/dorzheh/deployer/config/libvirt/bundle"
	"github.com/dorzheh/deployer/config/libvirt/xmlinput"
	"github.com/dorzheh/deployer/deployer"
	"github.com/dorzheh/deployer/post_processor/libvirt"
	gui "github.com/dorzheh/deployer/ui"
)

// InputData provides a static data
type InputData struct {
	// Path to the lshw binary
	Lshw string

	// Path to the basic configuration file (see xmlinput package)
	InputDataXMLFile string

	// Path to the bundle configuration file (see bundle package)
	BundleDataXMLFile string

	// Path to directory containing appropriate templates intended
	// for overriding default configuration
	NetworkDataTemplatesDir string
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

	// Hostinfo driver
	Hwdriver deployer.HostinfoDriver

	// Common metadata stuff
	Metadata *commonMetadata

	// Path to metadata file (libvirt XML)
	DestMetadataFile string

	BundleConf *bundle.Config
}

func CreateConfig(d *deployer.CommonData, i *InputData) (*Config, error) {
	var err error
	d.DefaultExportDir = "/var/lib/libvirt/images"

	c := &Config{common.CreateConfig(d), nil, nil, "", nil}
	c.Metadata = new(commonMetadata)

	driver := libvirt.NewDriver(c.SshConfig)
	if c.Metadata.EmulatorPath, err = driver.Emulator(d.Arch); err != nil {
		return nil, err
	}

	d.VaName = gui.UiApplianceName(d.Ui, d.VaName, driver)
	c.Metadata.DomainName = d.VaName
	c.Metadata.ImagePath = filepath.Join(c.ExportDir, d.VaName)

	c.DestMetadataFile = filepath.Join(c.ExportDir, d.VaName+".xml")
	// always create default metadata
	if err := ioutil.WriteFile(c.DestMetadataFile, defaultMetdata, 0); err != nil {
		return nil, err
	}

	if fd, err := os.Stat(i.InputDataXMLFile); err == nil && !fd.IsDir() {
		xid, err := xmlinput.ParseXML(i.InputDataXMLFile)
		if err != nil {
			return nil, err
		}
		if err := ResourcesConfig(d, i, xid, c); err != nil {
			return nil, err
		}
	}
	return c, nil
}
