package deployer

import (
	ui "github.com/dorzheh/deployer/ui/dialog_ui"
	ssh "github.com/dorzheh/infra/comm/common"
)

// CommonData represents the data
// required by deployer at any stage
// RootDir  - directory containing components
// needed for creaing artifacts
// DefaultExportDir - default directory for created artifact
// VaName - default name for the virtual appliance
// Arch - archirecture we'r running on
// Ui - user interface
type CommonData struct {
	RootDir          string
	RootfsMp         string
	DefaultExportDir string
	VaName           string
	Arch             string
	Ui               *ui.DialogUi
}

// CommonConfig represents common configuration
// generated during either user input or pasing appropriate
// configuration file
// RemoteMode - indicates whether the deployment occures remotely or not
// ExportDir - directory for storing appropriate artifacts
// SshConfig - ssh configuration
// Data - common data (above)
type CommonConfig struct {
	RemoteMode bool
	ExportDir  string
	SshConfig  *ssh.Config
	Data       *CommonData
}
