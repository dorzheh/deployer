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
	  	<size_gb>5</size_gb>
    	<bootable>true</bootable>
	 	<fdisk_cmd>n\np\n1\n\n+3045M\nn\np\n2\n\n\nt\n2\n82\na\n1\nw\n</fdisk_cmd>
   	    <description>Topology for release xxxx</description>
  	 	<partition>
	 	    <sequence>1</sequence>
	 	    <size_mb>3045</size_mb>
   	    	<label>SLASH</label>
   	    	<mount_point>/</mount_point>
   	    	<file_system>ext4</file_system>
	 	    <file_system_args></file_system_args>
	 	 </partition>
	 	 <partition>
	 	    <sequence>2</sequence>
	 	    <size_mb>400</size_mb>
   	    	<label>SWAP</label>
   	    	<mount_point>SWAP</mount_point>
   	    	<file_system>swap</file_system>
	 	    <file_system_args></file_system_args>
	 	 </partition>
 	 </disk>
 </config>
</storage>`)

func TestParse(t *testing.T) {
	d, err := ParseConfig(data)
	if err != nil {
		t.Fatal(err)
	}

	for _, conf := range d.Configs {
		for _, disk := range conf.Disks {
			fmt.Printf("%v\n", disk)
		}
	}
}
