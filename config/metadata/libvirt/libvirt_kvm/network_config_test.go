package libvirt_kvm

import (
	"fmt"
	"testing"

	"github.com/dorzheh/deployer/config/xmlinput"
	"github.com/dorzheh/deployer/utils/hwinfo/guest"
	"github.com/dorzheh/deployer/utils/hwinfo/host"
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

const expectedSetNetworkDataSRIOV = `<interface type='hostdev' managed='yes'>
    <source>
      <address type='pci' domain='0x0000' bus='0x05' slot='0x00' function='0x0'/>
    </source>
  </interface>
  <interface type='hostdev' managed='yes'>
	 <source>
	   <address type='pci' domain='0x0000' bus='0x06' slot='0x00' function='0x0'/>
	 </source>
   </interface>
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
	hnic1 := &host.NIC{
		Name:    "eth0",
		Driver:  "tg3",
		Vendor:  "Broadcom",
		Model:   "",
		PCIAddr: "pci@0000:09:00.0",
		Desc:    "test1_direct",
		Type:    host.NicTypePhys,
	}

	gnic1 := guest.NewNIC()
	gnic1.Network = "net1"
	gnic1.PCIAddr.Domain = "0000"
	gnic1.PCIAddr.Slot = "5"
	gnic1.PCIAddr.Bus = "00"
	gnic1.PCIAddr.Function = "0"
	gnic1.HostNIC = hnic1

	list1 := guest.NewNICList()
	list1.Add(gnic1)

	mode1 := new(xmlinput.Mode)
	mode1.Type = xmlinput.ConTypeDirect
	mode1.VnicDriver = "virtio"

	net1 := new(xmlinput.Network)
	net1.Modes = []*xmlinput.Mode{mode1}
	net1.Name = "net1"

	hnic2 := &host.NIC{
		Name:    "eth1",
		Driver:  "tg3",
		Vendor:  "Broadcom",
		Model:   "",
		PCIAddr: "pci@0000:10:00.0",
		Desc:    "test1_direct",
		Type:    host.NicTypePhys,
	}

	gnic2 := guest.NewNIC()
	gnic2.Network = "net1"
	gnic2.PCIAddr.Domain = "0000"
	gnic2.PCIAddr.Slot = "6"
	gnic2.PCIAddr.Bus = "00"
	gnic2.PCIAddr.Function = "0"
	gnic2.HostNIC = hnic2

	list2 := guest.NewNICList()
	list2.Add(gnic2)

	mode2 := new(xmlinput.Mode)
	mode2.Type = xmlinput.ConTypeDirect
	mode2.VnicDriver = "virtio"

	net2 := new(xmlinput.Network)
	net2.Modes = []*xmlinput.Mode{mode2}
	net2.Name = "net2"

	lists := []guest.NICList{list1, list2}
	nets := []*xmlinput.Network{net1, net2}

	c := guest.NewConfig()
	c.NICLists = lists
	c.Networks = nets
	d := new(meta)
	str, err := d.SetNetworkData(c, "")
	if err != nil {
		t.Fatal(err)
	}

	fmt.Printf("Template: %s\n\n", TmpltDirect)
	fmt.Printf("Expected: %s\n", expectedSetNetworkDataDirect)
	fmt.Printf("Generated:%s\n", str)
}

func TestSetNetworkDataPassthrough(t *testing.T) {
	hnic1 := &host.NIC{
		Name:    "eth0",
		Driver:  "igb",
		Vendor:  "Intel",
		Model:   "",
		PCIAddr: "pci@0000:09:00.0",
		Desc:    "test1_passthrough",
		Type:    host.NicTypePhys,
	}

	gnic1 := guest.NewNIC()
	gnic1.Network = "net1"
	gnic1.PCIAddr.Domain = "0000"
	gnic1.PCIAddr.Slot = "5"
	gnic1.PCIAddr.Bus = "00"
	gnic1.PCIAddr.Function = "0"
	gnic1.HostNIC = hnic1

	list1 := guest.NewNICList()
	list1.Add(gnic1)

	mode1 := new(xmlinput.Mode)
	mode1.Type = xmlinput.ConTypePassthrough

	net1 := new(xmlinput.Network)
	net1.Modes = []*xmlinput.Mode{mode1}
	net1.Name = "net1"

	hnic2 := &host.NIC{
		Name:    "eth1",
		Driver:  "igb",
		Vendor:  "Intel",
		Model:   "",
		PCIAddr: "pci@0000:10:00.0",
		Desc:    "test1_passthrough",
		Type:    host.NicTypePhys,
	}

	gnic2 := guest.NewNIC()
	gnic2.Network = "net2"
	gnic2.PCIAddr.Domain = "0000"
	gnic2.PCIAddr.Slot = "6"
	gnic2.PCIAddr.Bus = "00"
	gnic2.PCIAddr.Function = "0"
	gnic2.HostNIC = hnic2
	list2 := guest.NewNICList()
	list2.Add(gnic2)

	mode2 := new(xmlinput.Mode)
	mode2.Type = xmlinput.ConTypePassthrough

	net2 := new(xmlinput.Network)
	net2.Modes = []*xmlinput.Mode{mode2}
	net2.Name = "net2"

	lists := []guest.NICList{list1, list2}
	nets := []*xmlinput.Network{net1, net2}

	c := guest.NewConfig()
	c.NICLists = lists
	c.Networks = nets
	d := new(meta)
	str, err := d.SetNetworkData(c, "")
	if err != nil {
		t.Fatal(err)
	}

	fmt.Printf("Template: %s\n\n", TmpltPassthrough)
	fmt.Printf("Expected: %s\n", expectedSetNetworkDataPassthrough)
	fmt.Printf("Generated:%s\n", str)
}

func TestSetNetworkDataSRIOV(t *testing.T) {
	hnic1 := &host.NIC{
		Name:    "eth0",
		Driver:  "igb",
		Vendor:  "Intel",
		Model:   "",
		PCIAddr: "pci@0000:09:00.0",
		Desc:    "test1_sriov",
		Type:    host.NicTypePhysVF,
	}

	gnic1 := guest.NewNIC()
	gnic1.Network = "net1"
	gnic1.PCIAddr.Domain = "0000"
	gnic1.PCIAddr.Slot = "5"
	gnic1.PCIAddr.Bus = "00"
	gnic1.PCIAddr.Function = "0"
	gnic1.HostNIC = hnic1

	list1 := guest.NewNICList()
	list1.Add(gnic1)

	mode1 := new(xmlinput.Mode)
	mode1.Type = xmlinput.ConTypeSRIOV

	net1 := new(xmlinput.Network)
	net1.Modes = []*xmlinput.Mode{mode1}
	net1.Name = "net1"

	hnic2 := &host.NIC{
		Name:    "eth1",
		Driver:  "igb",
		Vendor:  "Intel",
		Model:   "",
		PCIAddr: "pci@0000:10:00.0",
		Desc:    "test1_sriov",
		Type:    host.NicTypePhysVF,
	}

	gnic2 := guest.NewNIC()
	gnic2.Network = "net2"
	gnic2.PCIAddr.Domain = "0000"
	gnic2.PCIAddr.Slot = "6"
	gnic2.PCIAddr.Bus = "00"
	gnic2.PCIAddr.Function = "0"
	gnic2.HostNIC = hnic2
	list2 := guest.NewNICList()
	list2.Add(gnic2)

	mode2 := new(xmlinput.Mode)
	mode2.Type = xmlinput.ConTypeSRIOV

	net2 := new(xmlinput.Network)
	net2.Modes = []*xmlinput.Mode{mode2}
	net2.Name = "net2"

	lists := []guest.NICList{list1, list2}
	nets := []*xmlinput.Network{net1, net2}

	c := guest.NewConfig()
	c.NICLists = lists
	c.Networks = nets
	d := new(meta)
	str, err := d.SetNetworkData(c, "")
	if err != nil {
		t.Fatal(err)
	}

	fmt.Printf("Template: %s\n\n", TmpltSriovPassthrough)
	fmt.Printf("Expected: %s\n", expectedSetNetworkDataSRIOV)
	fmt.Printf("Generated:%s\n", str)
}

func TestSetNetworkDataBridged(t *testing.T) {
	hnic1 := &host.NIC{
		Name:    "br0",
		Driver:  "",
		Vendor:  "",
		Model:   "",
		PCIAddr: "",
		Desc:    "test1_bridge",
		Type:    host.NicTypeBridge,
	}

	gnic1 := guest.NewNIC()
	gnic1.Network = "net1"
	gnic1.PCIAddr.Domain = "0000"
	gnic1.PCIAddr.Slot = "5"
	gnic1.PCIAddr.Bus = "00"
	gnic1.PCIAddr.Function = "0"
	gnic1.HostNIC = hnic1

	list1 := guest.NewNICList()
	list1.Add(gnic1)

	mode1 := new(xmlinput.Mode)
	mode1.Type = xmlinput.ConTypeBridged
	mode1.VnicDriver = "virtio"

	net1 := new(xmlinput.Network)
	net1.Modes = []*xmlinput.Mode{mode1}
	net1.Name = "net1"

	hnic2 := &host.NIC{
		Name:    "br1",
		Driver:  "",
		Vendor:  "",
		Model:   "",
		PCIAddr: "",
		Desc:    "test1_bridge",
		Type:    host.NicTypeBridge,
	}

	gnic2 := guest.NewNIC()
	gnic2.Network = "net2"
	gnic2.PCIAddr.Domain = "0000"
	gnic2.PCIAddr.Slot = "6"
	gnic2.PCIAddr.Bus = "00"
	gnic2.PCIAddr.Function = "0"
	gnic2.HostNIC = hnic2

	list2 := guest.NewNICList()
	list2.Add(gnic2)

	mode2 := new(xmlinput.Mode)
	mode2.Type = xmlinput.ConTypeBridged
	mode2.VnicDriver = "virtio"

	net2 := new(xmlinput.Network)
	net2.Modes = []*xmlinput.Mode{mode2}
	net2.Name = "net2"

	lists := []guest.NICList{list1, list2}
	nets := []*xmlinput.Network{net1, net2}

	c := guest.NewConfig()
	c.NICLists = lists
	c.Networks = nets
	d := new(meta)
	str, err := d.SetNetworkData(c, "")
	if err != nil {
		t.Fatal(err)
	}

	fmt.Printf("Template: %s\n\n", TmpltBridged)
	fmt.Printf("Expected: %s\n", expectedSetNetworkDataBridged)
	fmt.Printf("Generated:%s\n", str)
}

func TestSetNetworkDataOVS(t *testing.T) {
	hnic1 := &host.NIC{
		Name:    "br0-ovs",
		Driver:  "openvswitch",
		Vendor:  "",
		Model:   "",
		PCIAddr: "",
		Desc:    "test1_ovs",
		Type:    host.NicTypeOVS,
	}
	gnic1 := guest.NewNIC()
	gnic1.Network = "net1"
	gnic1.PCIAddr.Domain = "0000"
	gnic1.PCIAddr.Slot = "5"
	gnic1.PCIAddr.Bus = "00"
	gnic1.PCIAddr.Function = "0"
	gnic1.HostNIC = hnic1

	list1 := guest.NewNICList()
	list1.Add(gnic1)

	mode1 := new(xmlinput.Mode)
	mode1.Type = xmlinput.ConTypeOVS
	mode1.VnicDriver = "virtio"

	net1 := new(xmlinput.Network)
	net1.Modes = []*xmlinput.Mode{mode1}
	net1.Name = "net1"

	hnic2 := &host.NIC{
		Name:    "br1-ovs",
		Driver:  "openvswitch",
		Vendor:  "",
		Model:   "",
		PCIAddr: "",
		Desc:    "test1_ovs",
		Type:    host.NicTypeOVS,
	}
	gnic2 := guest.NewNIC()
	gnic2.Network = "net2"
	gnic2.PCIAddr.Domain = "0000"
	gnic2.PCIAddr.Slot = "6"
	gnic2.PCIAddr.Bus = "00"
	gnic2.PCIAddr.Function = "0"
	gnic2.HostNIC = hnic2

	list2 := guest.NewNICList()
	list2.Add(gnic2)

	mode2 := new(xmlinput.Mode)
	mode2.Type = xmlinput.ConTypeOVS
	mode2.VnicDriver = "virtio"

	net2 := new(xmlinput.Network)
	net2.Modes = []*xmlinput.Mode{mode2}
	net2.Name = "net2"

	lists := []guest.NICList{list1, list2}
	nets := []*xmlinput.Network{net1, net2}

	c := guest.NewConfig()
	c.NICLists = lists
	c.Networks = nets
	d := new(meta)
	str, err := d.SetNetworkData(c, "")
	if err != nil {
		t.Fatal(err)
	}

	fmt.Printf("Template: %s\n\n", TmpltBridgedOVS)
	fmt.Printf("Expected: %s\n", expectedSetNetworkDataBridgedOVS)
	fmt.Printf("Generated:%s\n", str)
}

func TestSetNetworkDataVirtualNet(t *testing.T) {
	hnic1 := &host.NIC{
		Name:    "network1",
		Driver:  "",
		Vendor:  "",
		Model:   "",
		PCIAddr: "",
		Desc:    "test1_virtnet",
		Type:    host.NicTypeVirtualNetwork,
	}

	gnic1 := guest.NewNIC()
	gnic1.Network = "net1"
	gnic1.PCIAddr.Domain = "0000"
	gnic1.PCIAddr.Slot = "5"
	gnic1.PCIAddr.Bus = "00"
	gnic1.PCIAddr.Function = "0"
	gnic1.HostNIC = hnic1

	list1 := guest.NewNICList()
	list1.Add(gnic1)

	mode1 := new(xmlinput.Mode)
	mode1.Type = xmlinput.ConTypeVirtualNetwork
	mode1.VnicDriver = "virtio"

	net1 := new(xmlinput.Network)
	net1.Modes = []*xmlinput.Mode{mode1}
	net1.Name = "net1"

	hnic2 := &host.NIC{
		Name:    "network2",
		Driver:  "",
		Vendor:  "",
		Model:   "",
		PCIAddr: "",
		Desc:    "test1_virtnet",
		Type:    host.NicTypeVirtualNetwork,
	}

	gnic2 := guest.NewNIC()
	gnic2.Network = "net2"
	gnic2.PCIAddr.Domain = "0000"
	gnic2.PCIAddr.Slot = "6"
	gnic2.PCIAddr.Bus = "00"
	gnic2.PCIAddr.Function = "0"
	gnic2.HostNIC = hnic2

	list2 := guest.NewNICList()
	list2.Add(gnic2)

	mode2 := new(xmlinput.Mode)
	mode2.Type = xmlinput.ConTypeVirtualNetwork
	mode2.VnicDriver = "virtio"

	net2 := new(xmlinput.Network)
	net2.Modes = []*xmlinput.Mode{mode2}
	net2.Name = "net2"

	lists := []guest.NICList{list1, list2}
	nets := []*xmlinput.Network{net1, net2}

	c := guest.NewConfig()
	c.NICLists = lists
	c.Networks = nets
	d := new(meta)
	str, err := d.SetNetworkData(c, "")
	if err != nil {
		t.Fatal(err)
	}

	fmt.Printf("Template: %s\n\n", TmpltVirtNetwork)
	fmt.Printf("Expected: %s\n", expectedSetNetworkDataVirtNetwork)
	fmt.Printf("Generated:%s\n", str)
}
