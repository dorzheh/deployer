package libvirt_kvm

import (
	"fmt"
	"testing"

	"github.com/dorzheh/deployer/config/common/xmlinput"
	"github.com/dorzheh/deployer/deployer"
	"github.com/dorzheh/deployer/utils/hwinfo"
)

const expectedSetNetworkDataDirect = `<interface type='direct'>
      <source dev='eth0' mode='private'/>
      <model type='virtio'/>
 </interface>

 <interface type='direct'>
      <source dev='eth0' mode='private'/>
      <model type='virtio'/>
 </interface>
`

const expectedSetNetworkDataPassthrough = `<hostdev mode='subsystem' type='pci' managed='yes'>
    <source>
      <address type='pci' domain='0x0000' bus='0x05' slot='0x00' function='0x0'/>
    </source>
  </hostdev>
  <hostdev mode='subsystem' type='pci' managed='yes'>
	 <source>
	   <address type='pci' domain='0x0000' bus='0x06' slot='0x00' function='0x0'/>
	 </source>
   </hostdev>
`

const expectedSetNetworkDataBridged = `<interface type='bridge'>
      <source bridge='br0'/>
	  <model type='virtio'/>
	  <driver name='vhost'/>
 </interface>
 <interface type='bridge'>
	  <source bridge='br1'/>
	  <model type=''/>
	  <driver name='vhost'/>
 </interface>
`

const expectedSetNetworkDataBridgedOVS = `<interface type='bridge'>
      <source bridge='br1-ovs'/>
      <virtualport type='openvswitch'/>
	  <model type='virtio'/>
	  <driver name='vhost'/>
</interface>
<interface type='bridge'>
      <source bridge='br2-ovs'/>
      <virtualport type='openvswitch'/>
	  <model type='virtio'/>
	  <driver name='vhost'/>
</interface>
`

const expectedSetNetworkDataVirtNetwork = `<interface type='network'>
	      <source network='network1'/>
	      <model type='virtio'/>
	      <driver name='vhost'/>
	    </interface>
	 <interface type='network'>
	      <source network='network2'/>
	      <model type='virtio'/>
	      <driver name='vhost'/>
	    </interface>
`

func TestSetNetworkDataDirect(t *testing.T) {
	nic1 := &hwinfo.NIC{
		Name:    "eth0",
		Driver:  "tg3",
		Vendor:  "Broadcom",
		Model:   "",
		PCIAddr: "pci@0000:05:00.0",
		Desc:    "test1_direct",
		Type:    hwinfo.NicTypePhys,
	}

	list1 := hwinfo.NewNICList()
	list1.Add(nic1)

	mode1 := new(xmlinput.Mode)
	mode1.Type = xmlinput.ConTypeDirect
	mode1.VnicDriver = "virtio"

	net1 := new(xmlinput.Network)
	net1.Mandatory = true
	net1.MaxIfaces = 1
	net1.Modes = []*xmlinput.Mode{mode1}
	net1.Name = "net1"

	nic2 := &hwinfo.NIC{
		Name:    "eth1",
		Driver:  "tg3",
		Vendor:  "Broadcom",
		Model:   "",
		PCIAddr: "pci@0000:06:00.0",
		Desc:    "test1_direct",
		Type:    hwinfo.NicTypePhys,
	}

	list2 := hwinfo.NewNICList()
	list2.Add(nic2)

	mode2 := new(xmlinput.Mode)
	mode2.Type = xmlinput.ConTypeDirect
	mode2.VnicDriver = "virtio"

	net2 := new(xmlinput.Network)
	net2.Mandatory = true
	net2.MaxIfaces = 1
	net2.Modes = []*xmlinput.Mode{mode2}
	net2.Name = "net2"

	lists := []hwinfo.NICList{list1, list2}
	nets := []*xmlinput.Network{net1, net2}

	mapping := &deployer.OutputNetworkData{lists, nets}
	d := new(meta)
	str, err := d.SetNetworkData(mapping, "")
	if err != nil {
		t.Fatal(err)
	}

	fmt.Printf("Template: %s\n\n", TmpltDirect)
	fmt.Printf("Expected: %s\n", expectedSetNetworkDataDirect)
	fmt.Printf("Generated:%s\n", str)
}

func TestSetNetworkDataPassthrough(t *testing.T) {
	nic1 := &hwinfo.NIC{
		Name:    "eth0",
		Driver:  "igb",
		Vendor:  "Intel",
		Model:   "",
		PCIAddr: "pci@0000:05:00.0",
		Desc:    "test1_passthrough",
		Type:    hwinfo.NicTypePhys,
	}

	list1 := hwinfo.NewNICList()
	list1.Add(nic1)

	mode1 := new(xmlinput.Mode)
	mode1.Type = xmlinput.ConTypePassthrough

	net1 := new(xmlinput.Network)
	net1.Mandatory = true
	net1.MaxIfaces = 1
	net1.Modes = []*xmlinput.Mode{mode1}
	net1.Name = "net1"

	nic2 := &hwinfo.NIC{
		Name:    "eth1",
		Driver:  "igb",
		Vendor:  "Intel",
		Model:   "",
		PCIAddr: "pci@0000:06:00.0",
		Desc:    "test1_passthrough",
		Type:    hwinfo.NicTypePhys,
	}

	list2 := hwinfo.NewNICList()
	list2.Add(nic2)

	mode2 := new(xmlinput.Mode)
	mode2.Type = xmlinput.ConTypePassthrough

	net2 := new(xmlinput.Network)
	net2.Mandatory = true
	net2.MaxIfaces = 1
	net2.Modes = []*xmlinput.Mode{mode2}
	net2.Name = "net2"

	lists := []hwinfo.NICList{list1, list2}
	nets := []*xmlinput.Network{net1, net2}

	mapping := &deployer.OutputNetworkData{lists, nets}
	d := new(meta)
	str, err := d.SetNetworkData(mapping, "")
	if err != nil {
		t.Fatal(err)
	}

	fmt.Printf("Template: %s\n\n", TmpltPassthrough)
	fmt.Printf("Expected: %s\n", expectedSetNetworkDataPassthrough)
	fmt.Printf("Generated:%s\n", str)
}

func TestSetNetworkDataBridged(t *testing.T) {
	nic1 := &hwinfo.NIC{
		Name:    "br0",
		Driver:  "",
		Vendor:  "",
		Model:   "",
		PCIAddr: "",
		Desc:    "test1_bridge",
		Type:    hwinfo.NicTypeBridge,
	}

	list1 := hwinfo.NewNICList()
	list1.Add(nic1)

	mode1 := new(xmlinput.Mode)
	mode1.Type = xmlinput.ConTypeBridged
	mode1.VnicDriver = "virtio"

	net1 := new(xmlinput.Network)
	net1.Mandatory = true
	net1.MaxIfaces = 1
	net1.Modes = []*xmlinput.Mode{mode1}
	net1.Name = "net1"

	nic2 := &hwinfo.NIC{
		Name:    "br1",
		Driver:  "",
		Vendor:  "",
		Model:   "",
		PCIAddr: "",
		Desc:    "test1_bridge",
		Type:    hwinfo.NicTypeBridge,
	}

	list2 := hwinfo.NewNICList()
	list2.Add(nic2)

	mode2 := new(xmlinput.Mode)
	mode2.Type = xmlinput.ConTypeBridged
	mode2.VnicDriver = "virtio"

	net2 := new(xmlinput.Network)
	net2.Mandatory = true
	net2.MaxIfaces = 1
	net2.Modes = []*xmlinput.Mode{mode2}
	net2.Name = "net2"

	lists := []hwinfo.NICList{list1, list2}
	nets := []*xmlinput.Network{net1, net2}

	mapping := &deployer.OutputNetworkData{lists, nets}
	d := new(meta)
	str, err := d.SetNetworkData(mapping, "")
	if err != nil {
		t.Fatal(err)
	}

	fmt.Printf("Template: %s\n\n", TmpltBridged)
	fmt.Printf("Expected: %s\n", expectedSetNetworkDataBridged)
	fmt.Printf("Generated:%s\n", str)
}

func TestSetNetworkDataOVS(t *testing.T) {
	nic1 := &hwinfo.NIC{
		Name:    "br0-ovs",
		Driver:  "openvswitch",
		Vendor:  "",
		Model:   "",
		PCIAddr: "",
		Desc:    "test1_ovs",
		Type:    hwinfo.NicTypeOVS,
	}

	list1 := hwinfo.NewNICList()
	list1.Add(nic1)

	mode1 := new(xmlinput.Mode)
	mode1.Type = xmlinput.ConTypeOVS
	mode1.VnicDriver = "virtio"

	net1 := new(xmlinput.Network)
	net1.Mandatory = true
	net1.MaxIfaces = 1
	net1.Modes = []*xmlinput.Mode{mode1}
	net1.Name = "net1"

	nic2 := &hwinfo.NIC{
		Name:    "br1-ovs",
		Driver:  "openvswitch",
		Vendor:  "",
		Model:   "",
		PCIAddr: "",
		Desc:    "test1_ovs",
		Type:    hwinfo.NicTypeOVS,
	}

	list2 := hwinfo.NewNICList()
	list2.Add(nic2)

	mode2 := new(xmlinput.Mode)
	mode2.Type = xmlinput.ConTypeOVS
	mode2.VnicDriver = "virtio"

	net2 := new(xmlinput.Network)
	net2.Mandatory = true
	net2.MaxIfaces = 1
	net2.Modes = []*xmlinput.Mode{mode2}
	net2.Name = "net2"

	lists := []hwinfo.NICList{list1, list2}
	nets := []*xmlinput.Network{net1, net2}

	mapping := &deployer.OutputNetworkData{lists, nets}
	d := new(meta)
	str, err := d.SetNetworkData(mapping, "")
	if err != nil {
		t.Fatal(err)
	}

	fmt.Printf("Template: %s\n\n", TmpltBridgedOVS)
	fmt.Printf("Expected: %s\n", expectedSetNetworkDataBridgedOVS)
	fmt.Printf("Generated:%s\n", str)
}

func TestSetNetworkDataVirtualNet(t *testing.T) {
	nic1 := &hwinfo.NIC{
		Name:    "network1",
		Driver:  "",
		Vendor:  "",
		Model:   "",
		PCIAddr: "",
		Desc:    "test1_virtnet",
		Type:    hwinfo.NicTypeVirtualNetwork,
	}

	list1 := hwinfo.NewNICList()
	list1.Add(nic1)

	mode1 := new(xmlinput.Mode)
	mode1.Type = xmlinput.ConTypeVirtualNetwork
	mode1.VnicDriver = "virtio"

	net1 := new(xmlinput.Network)
	net1.Mandatory = true
	net1.MaxIfaces = 1
	net1.Modes = []*xmlinput.Mode{mode1}
	net1.Name = "net1"

	nic2 := &hwinfo.NIC{
		Name:    "network2",
		Driver:  "",
		Vendor:  "",
		Model:   "",
		PCIAddr: "",
		Desc:    "test1_virtnet",
		Type:    hwinfo.NicTypeVirtualNetwork,
	}

	list2 := hwinfo.NewNICList()
	list2.Add(nic2)

	mode2 := new(xmlinput.Mode)
	mode2.Type = xmlinput.ConTypeVirtualNetwork
	mode2.VnicDriver = "virtio"

	net2 := new(xmlinput.Network)
	net2.Mandatory = true
	net2.MaxIfaces = 1
	net2.Modes = []*xmlinput.Mode{mode2}
	net2.Name = "net2"

	lists := []hwinfo.NICList{list1, list2}
	nets := []*xmlinput.Network{net1, net2}

	mapping := &deployer.OutputNetworkData{lists, nets}
	d := new(meta)
	str, err := d.SetNetworkData(mapping, "")
	if err != nil {
		t.Fatal(err)
	}

	fmt.Printf("Template: %s\n\n", TmpltVirtNetwork)
	fmt.Printf("Expected: %s\n", expectedSetNetworkDataVirtNetwork)
	fmt.Printf("Generated:%s\n", str)
}
