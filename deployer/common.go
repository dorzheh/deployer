package deployer

import (
	ui "github.com/dorzheh/deployer/ui/dialog_ui"
	ssh "github.com/dorzheh/infra/comm/common"
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
