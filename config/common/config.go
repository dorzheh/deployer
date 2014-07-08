package common

import (
	"github.com/dorzheh/deployer/deployer"
	gui "github.com/dorzheh/deployer/ui"
	ssh "github.com/dorzheh/infra/comm/common"
)

func CreateConfig(d *deployer.CommonData) *deployer.CommonConfig {
	c := new(deployer.CommonConfig)
	c.RemoteMode = gui.UiRemoteMode(d.Ui)
	if c.RemoteMode {
		c.SshConfig = new(ssh.Config)
		c.SshConfig.Host, c.SshConfig.Port, c.SshConfig.User, c.SshConfig.Password, c.SshConfig.PrvtKeyFile = gui.UiRemoteParams(d.Ui)
	}
	c.ExportDir = gui.UiImagePath(d.Ui, d.DefaultExportDir)
	c.Data = new(deployer.CommonData)
	c.Data = d
	return c
}
