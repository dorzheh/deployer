package libvirt

import (
	"strings"

	"github.com/dorzheh/deployer/deployer"
	"github.com/dorzheh/deployer/utils"
)

type passthroughData struct {
	Bus      string
	Slot     string
	Function string
}

type bridgedOVSData struct {
	OVSBridge string
}

type bridgedData struct {
	Bridge string
}

func SetNetworkData(ni map[string]*utils.NicInfo) (string, error) {
	var data string
	for _, port := range ni {
		switch port.Type {
		case utils.NicTypePhys:
			tempData, err := processPassthroughTemplate(port.PCIAddr)
			if err != nil {
				return "", err
			}
			data += string(tempData) + "\n"

		case utils.NicTypeOVS:
			tempData, err := deployer.ProcessTemplate(bridgedOVS, &bridgedOVSData{port.Name})
			if err != nil {
				return "", err
			}
			data += string(tempData) + "\n"

		case utils.NicTypeBridge:
			tempData, err := deployer.ProcessTemplate(bridged, &bridgedData{port.Name})
			if err != nil {
				return "", err
			}
			data += string(tempData) + "\n"
		}
	}
	return data, nil
}

func processPassthroughTemplate(pci string) ([]byte, error) {
	pciSlice := strings.Split(pci, ":")
	d := new(passthroughData)
	d.Bus = pciSlice[1]
	temp := strings.Split(pciSlice[2], ".")
	d.Slot = temp[0]
	d.Function = temp[1]
	return deployer.ProcessTemplate(passthrough, d)
}

var bridged = `<interface type='bridge'>
      <source bridge='{{ .Bridge }}'/>
 </interface>`

var bridgedOVS = `<interface type='bridge'>
      <source bridge='{{ .OVSBridge }}'/>
      <virtualport type='openvswitch'/>
	  <model type='virtio'/>
</interface>`

var passthrough = `<interface type='hostdev' managed='yes'>
      <source>
        <address type='pci' domain='0x0000' bus='0x{{ .Bus }}' slot='0x{{ .Slot }}' function='0x{{ .Function }}'/>
      </source>
	 <model type='virtio'/>
    </interface>
`
