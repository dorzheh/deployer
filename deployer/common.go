package deployer

import (
	ui "github.com/dorzheh/deployer/ui/dialog_ui"
	"github.com/dorzheh/infra/comm/ssh"
)

type CommonData struct {
	RootDir  string
	RootfsMp string
	VaName   string
	Ui       *ui.DialogUi
}

type CommonConfig struct {
	RemoteMode bool
	ExportDir  string
	SshConfig  *ssh.Config
}
