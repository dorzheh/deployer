package common

import (
	"github.com/dorzheh/deployer/deployer"
	gui "github.com/dorzheh/deployer/ui"
	"github.com/dorzheh/infra/comm/ssh"
)

func CreateConfig(d *deployer.CommonData) *deployer.CommonConfig {
	c := new(deployer.CommonConfig)
	c.RemoteMode = gui.UiRemoteMode(d.Ui)
	if c.RemoteMode {
		c.SshConfig = new(ssh.Config)
		c.SshConfig.Host, c.SshConfig.Port, c.SshConfig.User, c.SshConfig.Passwd, c.SshConfig.PrvtKeyFile = gui.UiRemoteParams(d.Ui)
	}
	c.ExportDir = gui.UiImagePath(d.Ui, d.RootDir)
	return c
}
