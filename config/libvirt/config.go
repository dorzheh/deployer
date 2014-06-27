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

type CommonInputData struct {
	Networks []string
	LshwPath string
}

type UserMetaData struct {
	DomainName   string
	AmountOfRam  string
	Cpus         string
	EmulatorPath string
	ImagePath    string
	Interfaces   string
}

type Config struct {
	Common   *deployer.CommonConfig
	Networks map[string]*utils.NicInfo
}

type Configurator struct {
	Config *Config
	Driver deployer.Driver
	Data   *UserMetaData
}

func (c *Configurator) Create(d *deployer.CommonData, i *CommonInputData) (*Configurator, error) {
	var err error

	c.Config.Common = common.CreateConfig(d)
	c.Driver = libvirt.NewDriver(c.Config.Common.SshConfig)
	if c.Data.EmulatorPath, err = c.Driver.Emulator(); err != nil {
		return nil, err
	}
	if i.LshwPath == "" {
		if i.LshwPath, err = exec.LookPath("lshw"); err != nil {
			return nil, err
		}
	}
	info, err := utils.NewHwInfoParser(filepath.Join(d.RootDir, "hwinfo.json"), i.LshwPath, c.Config.Common.SshConfig)

	go info.Parse()

	d.VaName = gui.UiApplianceName(d.Ui, d.VaName, c.Driver)
	c.Data.DomainName = d.VaName
	ni, err := info.NicsInfo()
	if err != nil {
		return nil, err
	}

	c.Config.Networks, err = gui.UiNetworks(d.Ui, i.Networks, ni)
	if err != nil {
		return nil, err
	}
	return c, nil
}
