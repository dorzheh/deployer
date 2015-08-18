package image

import (
	"fmt"
	"testing"
)

var data = []byte(`<?xml version="1.0" encoding="UTF-8"?>
<storage>
  <config>
	 <disk>
	 	<storage_type>qcow2</storage_type>
	  	<size_mb>5120</size_mb>
    	<bootable>true</bootable>
    	<bootloader>grub</bootloader>
    	<active_part>1</active_part>
	 	<fdisk_cmd>n\np\n1\n\n+3045M\nn\np\n2\n\n\nt\n2\n82\na\n1\nw\n</fdisk_cmd>
   	    <description>Topology for release xxxx</description>
  	 	<partition>
	 	    <sequence>1</sequence>
	 	    <type>83</type>
	 	    <size_mb>3045</size_mb>
	 	    <size_percents>-1</size_percents>
   	    	<label>SLASH</label>
   	    	<mount_point>/</mount_point>
   	    	<file_system>ext4</file_system>
	 	    <file_system_args></file_system_args>
	 	 </partition>
	 	 <partition>
	 	    <sequence>2</sequence>
	 	    <type>82</type>
	 	    <size_mb>400</size_mb>
	 	    <size_percents>-1</size_percents>
   	    	<label>SWAP</label>
   	    	<mount_point>SWAP</mount_point>
   	    	<file_system>swap</file_system>
	 	    <file_system_args></file_system_args>
	 	 </partition>
 	 </disk>
 </config>
</storage>`)

var data1 = []byte(`<?xml version="1.0" encoding="UTF-8"?>
<storage>
  <config>
	 <disk>
	 	<storage_type>qcow2</storage_type>
	  	<size_mb>5120</size_mb>
    	<bootable>true</bootable>
    	<active_part>1</active_part>
    	<bootloader>grub</bootloader>
	 	<fdisk_cmd></fdisk_cmd>
   	    <description>Topology for release xxxx</description>
  	 	<partition>
	 	    <sequence>1</sequence>
	 	    <size_mb>-1</size_mb>
	 	    <size_percents>90</size_percents>
   	    	<label>SLASH</label>
   	    	<mount_point>/</mount_point>
   	    	<file_system>ext4</file_system>
	 	    <file_system_args></file_system_args>
	 	 </partition>
	 	 <partition>
	 	    <sequence>2</sequence>
	 	    <size_mb>-1</size_mb>
	 	    <size_percents>10</size_percents>
   	    	<label>SWAP</label>
   	    	<mount_point>SWAP</mount_point>
   	    	<file_system>swap</file_system>
	 	    <file_system_args></file_system_args>
	 	 </partition>
 	 </disk>
 </config>
</storage>`)

func TestDataParse(t *testing.T) {
	d, err := ParseConfig(data)
	if err != nil {
		t.Fatal(err)
	}

	for _, conf := range d.Configs {
		for _, disk := range conf.Disks {
			fmt.Printf("%q\n", disk)
		}
	}
}

func TestData1Parse(t *testing.T) {
	d, err := ParseConfig(data1)
	if err != nil {
		t.Fatal(err)
	}

	for _, conf := range d.Configs {
		for _, disk := range conf.Disks {
			fmt.Printf("%v\n", disk)
		}
	}

}
