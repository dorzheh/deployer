package deployer

import (
	"github.com/dorzheh/deployer/builder/common/image"
	"github.com/dorzheh/deployer/config/common/xmlinput"
	"github.com/dorzheh/deployer/utils/hwinfo"
)

type MetadataConfigurator interface {
	// storage configuration and templates directory
	// returns metadata entry related to storage and error
	SetStorageData(*image.Config, string) (string, error)

	// network interfaces information, templates directory
	// returns metadata entry related to the network interfaces configuration and error
	SetNetworkData(*OutputNetworkData, string) (string, error)

	// default metadata is used by deployer in case user didn't provide any template
	// returns entry related to default metadata
	DefaultMetadata() []byte
}

type OutputNetworkData struct {
	NICLists []hwinfo.NICList
	Networks []*xmlinput.Network
}
