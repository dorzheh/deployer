package xmlinput

import (
	"fmt"
	"testing"
)

var xmldata = []byte(`<?xml version="1.0" encoding="UTF-8"?>
<input_data>
  <cpu>
  <configure>true</configure>
 	<min>1</min>
	<max>16</max>
	<default_value>1</default_value>
  </cpu>
  <numa>
    <auto_config>true</auto_config>
    <warn_on_unpinned_cpus>true</warn_on_unpinned_cpus>
    <ui_edit_config>true</ui_edit_config>
  </numa>
  <ram>
  	<configure>true</configure>
  	<min_mb>2500</min_mb>
    <max_mb>-1</max_mb>
	<default_value_mb>2500</default_value_mb>
  </ram>
  <disks>
    <configure>true</configure>
   	<disk>
  		<min_mb>2500</min_mb>
    	<max_mb>40000</max_mb>
		<default_value_mb>2500</default_value_mb>
	</disk>
	<disk>
  		<min_mb>50000</min_mb>
    	<max_mb>70000</max_mb>
		<default_value_mb>60000</default_value_mb>
	</disk>
  </disks>
  <networks>
    <configure>true</configure>
	<network name="Management">
		<!-- In common cases the templates provided by deployer are good enough -->
		<!-- however you may provide your own templates -->
		<!-- template_name is the name of your template file and -->
		<!-- dir is where your file is located -->
		<!-- in order to be able provide the path further set dir=""-->
	    <mode type="bridged" vnic_driver="e1000">
	    	 <template file_name="mngmnt_network.tmplt" dir="/opt/mytemplates"/>
	    </mode>
		<mode type="direct" vnic_driver="e1000"/>
        <mode type="ovs" vnic_driver="virtio"/>   
	</network>
	<network name="Traffic">
		<mode type="bridged" vnic_driver="virtio"/>
		<mode type="ovs" vnic_driver="virtio"/>
		<mode type="passthrough"/>
		<mode type="sriov"/>
		<host_nics_disjunction>true</host_nics_disjunction>
		<ui_reset_counter>true</ui_reset_counter>
		<ui_mode_selection>
			<appearance mode_type="bridged" appear="virtio"/>
			<appearance mode_type="passthrough" appear="pass-through"/>
			<appearance mode_type="sriov" appear="pass-through"/>
		</ui_mode_selection>
	</network>
	<network name="Optional" optional="true"> 
		<mode type="bridged" vnic_driver="virtio"/>
		<mode type="ovs" vnic_driver="virtio"/>
		<mode type="passthrough"/>
		<mode type="sriov"/>
		<ui_reset_counter>true</ui_reset_counter>
		<ui_mode_selection>
			<appearance mode_type="bridged" appear="virtio"/>
			<appearance mode_type="passthrough" appear="pass-through"/>
			<appearance mode_type="sriov" appear="pass-through"/>
		</ui_mode_selection>
	</network>
  </networks>
  <host_nics>
	<!-- Allowed vendors and models -->
	<allow vendor="Intel" model="" priority="true"/>
	<allow vendor="Brocade" model="" disjunction="true"/>
	<allow vendor="Mellanox" model="" disjunction="true"/>
	<!-- Denied vendors and models -->
	<deny vendor="Broadcom" model=""/>
  </host_nics>
   <guest_nics>
    <pci>
	    <domain>0000</domain>
	    <bus>00</bus>
	    <function>0</function>
	    <first_slot>6</first_slot>
	</pci>
  </guest_nics>
</input_data>`)

func TestParseXMLInput(t *testing.T) {
	d, err := ParseXMLInputBuf(xmldata)
	if err != nil {
		t.Fatal(err)
	}
	for _, d := range d.Disks.Configs {
		fmt.Printf("%v\n", d)

	}
	for _, n := range d.Networks.Configs {
		fmt.Printf("%v\n", n)
		for _, m := range n.Modes {
			fmt.Printf("Network : %s mode: %v\n", n.Name, m.Type)
		}
	}

	fmt.Printf("AutoConfig %v\n", d.AutoConfig)
	fmt.Printf("Guest NIC PCI %v\n", d.GuestNic.PCI)
	fmt.Printf("Networks.Configs[1].UiResetCounter %v\n", d.Networks.Configs[1].UiResetCounter)

	for _, nic := range d.Allowed {
		fmt.Printf("\nAllowed : NIC Vendor =>%s|NIC Model => %s\n",
			nic.Vendor, nic.Model)
	}
	for _, nic := range d.Denied {
		fmt.Printf("\nDenied : NIC Vendor =>%s|NIC Model => %s\n",
			nic.Vendor, nic.Model)
	}
}

var bad_xmldata = []byte(`<?xml version="1.0" encoding="UTF-8"?>
<input_data>
  <cpu>
  <configure>true</configure>
 	<min>1</min>
	<max>16</max>
	<default_value>1</default_value>
  </cpu>
  <ram>
  	<configure>true</configure>
  	<min_mb>2500</min_mb>
    <max_mb>-1</max_mb>
	<default_value_mb>2500</default_value_mb>
  </ram>
   <disks>
    <configure>true</configure>
   	<disk>
  		<min_mb>2500</min_mb>
    	<max_mb>-1</max_mb>
		<default_value_mb>2500</default_value_mb>
	<disk>
  </disks>
  <networks>
    <configure>true</configure>
	<network name="Management">
	    <mode type="bridged" vnic_driver="e1000"/>
	    <mode type="ovs" vnic_driver="e1000"/>
		<mode type="direct" vnic_driver="e1000"/>
		<ui_mode_selection enable="false"/>
	</network>
	<network name="Traffic"> 
		<mode type="bridged" vnic_driver="virtio"/>
		<mode type="passthrough"/>
		<mode type="sriov"/>
		<mode type="direct"/>
		<ui_mode_selection enable="true">
			<appearance mode_type="bridged" appear="virtio"/>
			<appearance mode_type="passthrough" appear="passthrough"/>
			<appearance mode_type="sriov" appear="passthrough"/>
		</ui_mode_selection>
	</network>
	<network name="Bkp"> 
		<mode type="bridged" vnic_driver="virtio"/>
		<mode type="passthrough"/>
	</network>
  </networks>
   <host_nics>
	<!-- Allowed vendors and models -->
	<allow vendor="Intel" model=""/>
	<allow vendor="Broadcom" model=""/>
	<!-- Denied vendors and models -->
	<deny vendor="Broadcom" model=""/>
  </host_nics>
  <guest_nics>
    <pci>
	    <domain>0000</domain>
	    <bus>00</bus>
	    <function>0</function>
	    <first_slot>6</first_slot>
	</pci>
  </guest_nics>
</input_data>`)

func TestParseXMLInputBad(t *testing.T) {
	_, err := ParseXMLInputBuf(bad_xmldata)
	if err == nil {
		t.Fatalf("supposed to produce an error")
	}
}
