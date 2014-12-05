package libvirt

import (
	"fmt"
	"testing"

	"github.com/dorzheh/deployer/config/common/xmlinput"
	"github.com/dorzheh/deployer/utils/hwinfo"
)

const expectedSetNetworkDataDirect = `<interface type='direct'>
      <source dev='eth0' mode='private'/>
      <model type='virtio'/>
    </interface>
`

const expectedSetNetworkDataPassthrough = `<hostdev mode='subsystem' type='pci' managed='yes'>
    <source>
      <address type='pci' domain='0x0000' bus='0x05' slot='0x00' function='0x0'/>
    </source>
  </hostdev>
`

const expectedSetNetworkDataBridged = `<interface type='bridge'>
      <source bridge='br0'/>
	  <model type='virtio'/>
	  <driver name='vhost'/>
 </interface>
`

const expectedSetNetworkDataBridgedOVS = `<interface type='bridge'>
      <source bridge='br1-ovs'/>
      <virtualport type='openvswitch'/>
	  <model type='virtio'/>
	  <driver name='vhost'/>
</interface>
`

func TestSetNetworkDataDirect(t *testing.T) {
	info := map[string]*hwinfo.NIC{
		"test": &hwinfo.NIC{
			Name:    "eth0",
			Driver:  "tg3",
			Vendor:  "Broadcom",
			Model:   "",
			PCIAddr: "pci@0000:05:00.0",
			Desc:    "test1_direct",
			Type:    hwinfo.NicTypePhys,
		},
	}

	nics := []*xmlinput.Allow{&xmlinput.Allow{Vendor: "Broadcom", Model: ""}}
	d := new(meta)
	str, err := d.SetNetworkData(info, nics, "")
	if err != nil {
		t.Fatal(err)
	}

	fmt.Printf("Template: %s\n\n", TmpltDirect)
	fmt.Printf("Expected: %s\n", expectedSetNetworkDataDirect)
	fmt.Printf("Generated:%s\n", str)
}

func TestSetNetworkDataPassthrough(t *testing.T) {
	info := map[string]*hwinfo.NIC{
		"test": &hwinfo.NIC{
			Name:    "eth1",
			Driver:  "igb",
			Vendor:  "Intel",
			Model:   "I350 Gigabit Network Connection",
			PCIAddr: "pci@0000:05:00.0",
			Desc:    "test2_passthrough",
			Type:    hwinfo.NicTypePhys,
		},
	}

	nics := []*xmlinput.Allow{&xmlinput.Allow{Vendor: "Intel", Model: "", Mode: "passthrough"}}
	d := new(meta)
	str, err := d.SetNetworkData(info, nics, "")
	if err != nil {
		t.Fatal(err)
	}

	fmt.Printf("Template: %s\n\n", TmpltPassthrough)
	fmt.Printf("Expected: %s\n", expectedSetNetworkDataPassthrough)
	fmt.Printf("Generated:%s\n", str)
}

func TestSetNetworkDataBridged(t *testing.T) {
	info := map[string]*hwinfo.NIC{
		"test": &hwinfo.NIC{
			Name:    "br0",
			Driver:  "bridge",
			Vendor:  "",
			Model:   "",
			PCIAddr: "",
			Desc:    "test3_bridge",
			Type:    hwinfo.NicTypeBridge,
		},
	}

	d := new(meta)
	str, err := d.SetNetworkData(info, nil, "")
	if err != nil {
		t.Fatal(err)
	}

	fmt.Printf("Template: %s\n\n", TmpltBridged)
	fmt.Printf("Expected: %s\n", expectedSetNetworkDataBridged)
	fmt.Printf("Generated:%s\n", str)
}

func TestSetNetworkDataOVS(t *testing.T) {
	info := map[string]*hwinfo.NIC{
		"test": &hwinfo.NIC{
			Name:    "br1-ovs",
			Driver:  "openvswitch",
			Vendor:  "",
			Model:   "",
			PCIAddr: "",
			Desc:    "test4_ovs",
			Type:    hwinfo.NicTypeOVS,
		},
	}

	d := new(meta)
	str, err := d.SetNetworkData(info, nil, "")
	if err != nil {
		t.Fatal(err)
	}

	fmt.Printf("Template: %s\n\n", TmpltBridgedOVS)
	fmt.Printf("Expected: %s\n", expectedSetNetworkDataBridgedOVS)
	fmt.Printf("Generated:%s\n", str)
}
