package xmlinput

import (
	"fmt"
	"testing"
)

var xmldata = []byte(`<?xml version="1.0" encoding="UTF-8"?>
<InputData>
  <CPU>
    	<Config>true</Config>
 	<Min>1</Min>
	<Max>16</Max>
	<DefaultValue>1</DefaultValue>
  </CPU>
  <RAM>
	<Config>true</Config>
  	<Min>2500</Min>
      	<Max>0</Max>
	<DefaultValue>2500</DefaultValue>
  </RAM>
  <Networks>
	<Config>true</Config>
	<!-- Maximal amount of networks to configure -->
	<Max>10</Max>
	<Default>
		<Network>
			<Name>Management</Name>
		</Network>
		<Network>
			<Name>Data1</Name>
		</Network>
		<Network>
			<Name>Data2</Name>
		</Network>
	</Default>
  </Networks>
  <NICs>
	<!-- Allowed vendors and models -->
	<Allow> 
		<Vendor>Intel</Vendor>
		<Model></Model>
		<Mode>passthrough</Mode>
	</Allow>
	<Allow>
		<Vendor>Broadcom</Vendor>
		<Model></Model>
		<Mode>direct</Mode>
	</Allow>
	<!-- Denied vendors and models -->
	<Deny>
		<Vendor>Broadcom</Vendor>
		<Model></Model>
	</Deny>
   </NICs>
</InputData>`)

func TestParse(t *testing.T) {
	d, err := Parse(xmldata)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Printf("%v\n", d)

	for _, nic := range d.Allowed {
		fmt.Printf("\nNIC Vendor =>%s\nNIC Model => %s\nNIC Mode => %s\n",
			nic.Vendor, nic.Model, nic.Mode)
	}
}
