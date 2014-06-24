package deployer

import ui "github.com/dorzheh/deployer/ui/dialog_ui"

type FlowCreator interface {
	CreateConfig(*CommonData) error
	CreateBuilders() ([]Builder, error)
	CreateProvisioner() (Provisioner, error)
	CreatePostProcessor() (PostProcessor, error)
}

type commonPaths struct {
	RootDir  string
	RootfsMp string
}

type CommonData struct {
	commonPaths
	VaName string
	Ui     *ui.DialogUi
}
