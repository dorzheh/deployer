package libvirt

import (
	"fmt"
	"testing"

	"github.com/dorzheh/deployer/builder/common/image"
)

var data = []byte(`<?xml version="1.0" encoding="UTF-8"?>
<storage>
  <config>
	 <disk>
	  	<size_gb>1</size_gb>
    	<bootable>true</bootable>
	 	<fdisk_cmd>n\np\n1\n\n+800M\nn\np\n2\n\n\nt\n2\n82\na\n1\nw\n</fdisk_cmd>
   	 	<description>Topology for release xxxx</description>
  	 	<partition>
	 	    <sequence>1</sequence>
	 	    <size_mb>800</size_mb>
   	    	<label>SLASH</label>
   	    	<mount_point>/</mount_point>
   	    	<file_system>ext4</file_system>
	 	    <file_system_args></file_system_args>
	 	</partition>
	 	<partition>
	 		<sequence>2</sequence>
		 	<size_mb>200</size_mb>
   		    <label>SWAP</label>
   		    <mount_point>SWAP</mount_point>
   		    <file_system>swap</file_system>
		 	<file_system_args></file_system_args>
	 	 </partition>
 	 </disk>
 </config>
</storage>`)

const expectedSetStorageData = `<disk type='file' device='disk'>
	<driver name='qemu' type='raw' cache='none'/>
	<source file='/var/lib/libvirt/images/maindisk'/>
	<target dev='vda' bus='virtio'/>
	</disk>

<disk type='file' device='disk'>
	<driver name='qemu' type='raw' cache='none'/>
	<source file='/var/lib/libvirt/images/maindisk_1'/>
	<target dev='vdb' bus='virtio'/>
	</disk>

<disk type='file' device='disk'>
	<driver name='qemu' type='raw' cache='none'/>
	<source file='/var/lib/libvirt/images/maindisk_2'/>
	<target dev='vdc' bus='virtio'/>
	</disk>
`

func TestSetStorageData(t *testing.T) {
	c, err := image.ParseConfig(data)
	if err != nil {
		t.Fatal(err)
	}

	config := c.Configs[0]
	config.Disks[0].Path = "/var/lib/libvirt/images/maindisk"

	d := new(meta)
	str, err := d.SetStorageData(config, "")
	if err != nil {
		t.Fatal(err)
	}

	fmt.Printf("Template: %s\n\n", TmpltStorage)
	fmt.Printf("Expected: %s\n", expectedSetStorageData)
	fmt.Printf("Generated:%s\n", str)
}
