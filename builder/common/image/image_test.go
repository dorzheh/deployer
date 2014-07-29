package image

import (
	"os"
	"testing"

	ssh "github.com/dorzheh/infra/comm/common"
	"github.com/dorzheh/infra/comm/sshfs"
)

const imagePath = "/tmp/testImage"
const rootfsMp = "/tmp/mnt"
const grubPath = "/tmp/grub"
const topology = `<?xml version="1.0" encoding="UTF-8"?>
<Platforms>
  <Topology>
        <Name>TestImage</Name>
        <Type>1</Type>
        <HddSizeGb>1</HddSizeGb>
        <Bootable>true</Bootable>
        <FdiskCmd>o\nn\np\n1\n\n+800M\nn\np\n2\n\n\nt\n2\n82\na\n1\nw\n</FdiskCmd>
        <Description>Test platform toplogy</Description>
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
           <SizeMb>2024</SizeMb>
           <Label>SWAP</Label>
           <MountPoint>SWAP</MountPoint>
           <FileSystem>swap</FileSystem>
           <FileSystemArgs></FileSystemArgs>
        </Partition>
  </Topology>
</Platforms>
`

func getToplogy() (*Topology, error) {
	p, err := ParseConfig([]byte(topology))
	if err != nil {
		return nil, err
	}
	return p.TypeToTopology(0)
}

func TestLocalImage(t *testing.T) {
	t.Log("=> MKdirAll")
	if err := os.MkdirAll(rootfsMp, 0755); err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(rootfsMp)

	t.Log("=> getToplogy")
	topo, err := getToplogy()
	if err != nil {
		t.Fatal(err)
	}

	t.Log("=> New")
	img, err := New(imagePath, rootfsMp, topo, nil)
	if err != nil {
		t.Fatal(err)
	}

	// interrupt handler
	img.ReleaseOnInterrupt()

	defer func() {
		t.Log("=> Release")
		if err := img.Release(); err != nil {
			t.Fatal(err)
		}
		t.Log("=> MakeBootable")
		if err := img.MakeBootable(grubPath); err != nil {
			t.Fatal(err)
		}
		t.Log("=> Remove")
		os.Remove(imagePath)
	}()

	t.Log("=> Parse")
	if err := img.Parse(); err != nil {
		t.Fatal(err)
	}
}

func TestRemoteImage(t *testing.T) {
	t.Log("=> MkdirAll")
	if err := os.MkdirAll(rootfsMp, 0755); err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(rootfsMp)

	t.Log("=> getToplogy")
	topo, err := getToplogy()
	if err != nil {
		t.Fatal(err)
	}
	sshConf := &ssh.Config{
		Host:        "127.0.0.1",
		Port:        "22",
		User:        "username",
		Password:    "password",
		PrvtKeyFile: "",
	}
	sshfsConf := &sshfs.Config{sshConf, "", ""}

	t.Log("=> New")
	img, err := New(imagePath, rootfsMp, topo, sshfsConf)
	if err != nil {
		t.Fatal(err)
	}
	img.ReleaseOnInterrupt()

	defer func() {
		t.Log("=> Release")
		if err := img.Release(); err != nil {
			t.Fatal(err)
		}
		t.Log("=> MakeBootable")
		if err := img.MakeBootable(grubPath); err != nil {
			t.Fatal(err)
		}
		t.Log("=> Remove")
		os.Remove(imagePath)
	}()

	t.Log("=> Parse")
	if err := img.Parse(); err != nil {
		t.Fatal(err)
	}
}
