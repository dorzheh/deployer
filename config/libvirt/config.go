package libvirt

import (
	"os/exec"
	"path/filepath"

	"github.com/dorzheh/deployer/config/common"
	"github.com/dorzheh/deployer/deployer"
	"github.com/dorzheh/deployer/post_processor/libvirt"
	gui "github.com/dorzheh/deployer/ui"
	"github.com/dorzheh/deployer/utils"
)

type InputData struct {
	Networks []string
	LshwPath string
}

type commonOutputData struct {
	DomainName   string
	EmulatorPath string
	ImagePath    string
}

type Config struct {
	Common   *deployer.CommonConfig
	Networks map[string]*utils.NicInfo
	MetadataPath string
	Data     *commonOutputData
}

func CreateConfig(d *deployer.CommonData, i *InputData) (*Config, error) {
	var err error

	c := new(Config)
	c.Common = common.CreateConfig(d)
	driver := libvirt.NewDriver(c.Common.SshConfig)
	if c.Data.EmulatorPath, err = driver.Emulator(); err != nil {
		return nil, err
	}
	if i.LshwPath == "" {
		if i.LshwPath, err = exec.LookPath("lshw"); err != nil {
			return nil, err
		}
	}
	info, err := utils.NewHwInfoParser(filepath.Join(d.RootDir, "hwinfo.json"), i.LshwPath, c.Common.SshConfig)

	go info.Parse()

	d.VaName = gui.UiApplianceName(d.Ui, d.VaName, driver)
	c.Data.DomainName = d.VaName
	c.Data.ImagePath = filepath.Join(c.Common.ExportDir,c.Data.DomainName)
	c.MetadataPath = filepath.Join(c.Common.ExportDir,c.Data.DomainName,".xml")

	ni, err := info.NicsInfo()
	if err != nil {
		return nil, err
	}

	c.Networks, err = gui.UiNetworks(d.Ui, i.Networks, ni)
	if err != nil {
		return nil, err
	}
	return c, nil
}
