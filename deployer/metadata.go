package deployer

import (
	"github.com/dorzheh/deployer/builder/image"
	"github.com/dorzheh/deployer/config/common/xmlinput"
	"github.com/dorzheh/deployer/utils/hwinfo"
)

// MetadataConfigurator is the interface that has to be implemented
// in order to manipulate appropriate metadata.
type MetadataConfigurator interface {
	// Storage configuration and templates directory.
	// Returns storage related metadata entry and error.
	SetStorageData(*image.Config, string) (string, error)

	// Network interfaces information, templates directory.
	// Returns metadata entry related to the network interfaces configuration and error.
	SetNetworkData(*OutputNetworkData, string) (string, error)

	// Default metadata is used by deployer in case user didn't provide any template.
	// Returns entry related to default metadata.
	DefaultMetadata() []byte
}

type OutputNetworkData struct {
	NICLists []hwinfo.NICList
	Networks []*xmlinput.Network
}
