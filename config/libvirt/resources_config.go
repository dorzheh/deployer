package libvirt

import (
	"path/filepath"

	"github.com/dorzheh/deployer/config/libvirt/xmlinput"
	"github.com/dorzheh/deployer/deployer"
	"github.com/dorzheh/deployer/hwinfo_driver/libvirt"
	gui "github.com/dorzheh/deployer/ui"
)

func ResourcesConfig(d *deployer.CommonData, i *InputData, c *Config, xid *xmlinput.XMLInputData) error {
	var err error
	c.Hwdriver, err = libvirt.NewHostinfoDriver(filepath.Join(d.RootDir, "hwinfo.json"), i.Lshw, c.SshConfig)
	if err != nil {
		return err
	}
	if err := gui.UiGatherHWInfo(d.Ui, c.Hwdriver, "1s", c.RemoteMode); err != nil {
		return err
	}
	if xid.CPU.Config {
		cpus, err := c.Hwdriver.CPUs()
		if err != nil {
			return err
		}
		c.Metadata.CPUs = gui.UiCPUs(d.Ui, cpus, xid.CPU.Min, xid.CPU.Max)
	} else if xid.CPU.Default > 0 {
		c.Metadata.CPUs = xid.CPU.Default
	}

	if xid.RAM.Config {
		ram, err := c.Hwdriver.RAMSize()
		if err != nil {
			return err
		}
		c.Metadata.RAM = gui.UiRAMSize(d.Ui, ram, xid.RAM.Min, xid.RAM.Max)
		c.Metadata.RAM *= 1024
	} else if xid.RAM.Default > 0 {
		c.Metadata.RAM = xid.RAM.Default
		c.Metadata.RAM *= 1024
	}

	if xid.Networks.Config {
		nets, err := gui.UiNetworks(d.Ui, xid, c.Hwdriver)
		if err != nil {
			return err
		}

		c.Metadata.Networks, err = SetNetworkData(nets, xid.Allowed, i.NetworkDataTemplatesDir)
		if err != nil {
			return err
		}
	}
	return nil
}
