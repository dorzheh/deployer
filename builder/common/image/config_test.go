package image

import (
	"testing"
)

var data = []byte(`<?xml version="1.0" encoding="UTF-8"?>
<Storage>
  <Config>
	 <Disk>
	  	<SizeGb>5</SizeGb>
    	<Bootable>true</Bootable>
	 	<FdiskCmd>n\np\n1\n\n+3045M\nn\np\n2\n\n\nt\n2\n82\na\n1\nw\n</FdiskCmd>
   	    <Description>Topology for release xxxx</Description>
  	 	<Partition>
	 	    <Sequence>1</Sequence>
	 	    <SizeMb>3045</SizeMb>
   	    	<Label>SLASH</Label>
   	    	<MountPoint>/</MountPoint>
   	    	<FileSystem>ext4</FileSystem>
	 	    <FileSystemArgs></FileSystemArgs>
	 	 </Partition>
	 	 <Partition>
	 	    <Sequence>2</Sequence>
	 	    <SizeMb>400</SizeMb>
   	    	<Label>SWAP</Label>
   	    	<MountPoint>SWAP</MountPoint>
   	    	<FileSystem>swap</FileSystem>
	 	    <FileSystemArgs></FileSystemArgs>
	 	 </Partition>
 	 </Disk>
 </Config>
</Storage>`)

func TestParse(t *testing.T) {
	d, err := ParseConfig(data)
	if err != nil {
		t.Fatal(err)
	}

	for _, conf := range d.Configs {
		for _, disk := range conf.Disks {
			t.Logf("%v\n", disk)
		}
	}
}
