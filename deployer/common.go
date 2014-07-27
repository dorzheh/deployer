package deployer

import (
	ui "github.com/dorzheh/deployer/ui/dialog_ui"
	ssh "github.com/dorzheh/infra/comm/common"
)

// CommonData represents the data
// required by deployer at any stage
type CommonData struct {
	// Directory containing components required for
	// creating the target appliance
	RootDir string

	// Directory that will be used as a mount point
	// for
	RootfsMp string

	// DefaultExportDir - default directory for created artifact
	DefaultExportDir string

	// VaName - default name for the virtual appliance
	VaName string

	// Arch - archirecture we'r running on
	Arch string

	// Ui - user interface
	Ui *ui.DialogUi
}

// CommonConfig represents common configuration
// generated during either user input or pasing appropriate
// configuration file
type CommonConfig struct {
	// RemoteMode indicates whether the deployment occures remotely
	RemoteMode bool

	// ExportDir is a directory for storing appropriate artifacts
	ExportDir string

	// SshConfig represents ssh configuration
	SshConfig *ssh.Config

	// Data repesents common data
	Data *CommonData
}
