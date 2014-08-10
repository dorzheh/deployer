// Intended for creating a common configuration
// required by either target
package common

import (
	"github.com/dorzheh/deployer/deployer"
	gui "github.com/dorzheh/deployer/ui"
)

func CreateConfig(d *deployer.CommonData) *deployer.CommonConfig {
	c := new(deployer.CommonConfig)
	c.RemoteMode = gui.UiRemoteMode(d.Ui)
	if c.RemoteMode {
		c.SshConfig = gui.UiSshConfig(d.Ui)
	}
	c.ExportDir = gui.UiImagePath(d.Ui, d.DefaultExportDir, c.RemoteMode)
	c.Data = new(deployer.CommonData)
	c.Data = d
	return c
}
