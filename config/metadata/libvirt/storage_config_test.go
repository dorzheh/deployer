package libvirt

import (
	"fmt"
	"testing"

	"github.com/dorzheh/deployer/builder/common/image"
)

var data = []byte(`<?xml version="1.0" encoding="UTF-8"?>
<Storage>
  <Config>
	 <Disk>
	  	<SizeGb>1</SizeGb>
    	<Bootable>true</Bootable>
	 	 <FdiskCmd>n\np\n1\n\n+800M\nn\np\n2\n\n\nt\n2\n82\na\n1\nw\n</FdiskCmd>
   	 <Description>Topology for release xxxx</Description>
  	 	 <Partition>
	 	    <Sequence>1</Sequence>
	 	    <SizeMb>800</SizeMb>
   	    <Label>SLASH</Label>
   	    <MountPoint>/</MountPoint>
   	    <FileSystem>ext4</FileSystem>
	 	    <FileSystemArgs></FileSystemArgs>
	 	 </Partition>
	 	 <Partition>
	 	    <Sequence>2</Sequence>
	 	    <SizeMb>200</SizeMb>
   	    <Label>SWAP</Label>
   	    <MountPoint>SWAP</MountPoint>
   	    <FileSystem>swap</FileSystem>
	 	    <FileSystemArgs></FileSystemArgs>
	 	 </Partition>
 	 </Disk>
 </Config>
</Storage>`)

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
