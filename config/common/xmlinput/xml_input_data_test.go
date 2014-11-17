package xmlinput

import (
	"fmt"
	"testing"

	"github.com/dorzheh/deployer/utils"
)

var xmldata = []byte(`<?xml version="1.0" encoding="UTF-8"?>
<input_data>
  <cpu>
    <Config>true</Config>
 	<min>1</min>
	<max>16</max>
	<default_value>1</default_value>
  </cpu>
  <ram>
	<config>true</config>
  	<min>2500</min>
      	<max>0</max>
	<default_value>2500</default_value>
  </ram>
  <networks>
	<config>true</config>
	<!-- Maximal amount of networks to configure -->
	<max>10</max>
	<default>
		<network>
			<name>Management</name>
		</network>
		<network>
			<name>Data1</name>
		</network>
		<network>
			<name>Data2</name>
		</network>
	</default>
  </networks>
  <nics>
	<!-- Allowed vendors and models -->
	<allow> 
		<vendor>Intel</vendor>
		<model></model>
		<mode>passthrough</mode>
	</allow>
	<allow>
		<vendor>Broadcom</vendor>
		<model></model>
		<mode>direct</mode>
	</allow>
	<!-- Denied vendors and models -->
	<deny>
		<vendor>Broadcom</vendor>
		<model></model>
	</deny>
   </nics>
</input_data>`)

func TestParse(t *testing.T) {
	d, err := utils.ParseXMLBuff(xmldata, new(XMLInputData))
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%v\n", d.(*XMLInputData))

	for _, nic := range d.(*XMLInputData).Allowed {
		fmt.Printf("\nNIC Vendor =>%s\nNIC Model => %s\nNIC Mode => %s\n",
			nic.Vendor, nic.Model, nic.Mode)
	}
}
