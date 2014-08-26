package libvirt

import (
	"path/filepath"

	"github.com/dorzheh/deployer/config/libvirt/xmlinput"
	"github.com/dorzheh/deployer/deployer"
	gui "github.com/dorzheh/deployer/ui"
	"github.com/dorzheh/deployer/utils/hwinfo"
)

func ResourcesConfig(d *deployer.CommonData, i *InputData, c *Config, xid *xmlinput.XMLInputData) error {
	var err error
	c.HWInfo, err = hwinfo.NewParser(filepath.Join(d.RootDir, "hwinfo.json"), i.Lshw, c.SshConfig)
	if err != nil {
		return err
	}
	if err := gui.UiGatherHWInfo(d.Ui, c.HWInfo, "1s", c.RemoteMode); err != nil {
		return err
	}
	if xid.CPU.Config {
		cpus, err := c.HWInfo.CPUs()
		if err != nil {
			return err
		}
		c.Metadata.CPUs = gui.UiCPUs(d.Ui, cpus, xid.CPU.Min, xid.CPU.Max)
	} else if xid.CPU.Default > 0 {
		c.Metadata.CPUs = xid.CPU.Default
	}

	if xid.RAM.Config {
		ram, err := c.HWInfo.RAMSize()
		if err != nil {
			return err
		}
		c.Metadata.RAM = gui.UiRAMSize(d.Ui, ram, xid.RAM.Min, xid.RAM.Max)
	} else if xid.RAM.Default > 0 {
		c.Metadata.RAM = xid.RAM.Default
	}

	if xid.Networks.Config {
		ni, err := c.HWInfo.NICInfo()
		if err != nil {
			return err
		}

		nets, err := gui.UiNetworks(d.Ui, xid, ni)
		if err != nil {
			return err
		}

		c.Metadata.Networks, err = SetNetworkData(nets, i.NetworkDataTemplatesDir)
		if err != nil {
			return err
		}
	}
	return nil
}
