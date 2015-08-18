package image

import (
	"os"
	"testing"

	ssh "github.com/dorzheh/infra/comm/common"
	"github.com/dorzheh/infra/comm/sshfs"
)

const (
	imagePath = "/tmp/testImage"
	rootfsMp  = "/tmp/mnt"
)

var storage = []byte(`<?xml version="1.0" encoding="UTF-8"?>
<storage>
  <config>
	<disk>
		<storage_type>qcow2</storage_type>
        <size_mb>1024</size_mb>
        <bootable>true</bootable>
        <active_part>1</active_part>
        <bootloader>grub</bootloader>
        <fdisk_cmd></fdisk_cmd>
        <description>Test configuration</description>
        <partition>
           <sequence>1</sequence>
           <type>83</type>
           <size_mb>800</size_mb>
           <size_percents>-1</size_percents>
           <label>SLASH</label>
           <mount_point>/</mount_point>
           <file_system>ext4</file_system>
           <file_system_args></file_system_args>
        </partition>
        <partition>
           <sequence>2</sequence>
           <type>82</type>
           <size_mb>-1</size_mb>
           <size_percents>-2</size_percents>
           <label>SWAP</label>
           <mount_point>SWAP</mount_point>
           <file_system>swap</file_system>
           <file_system_args></file_system_args>
        </partition>
	</disk>
  </config>
</storage>
`)

var u = &Utils{
	Grub:   "/tmp/grub",
	Kpartx: "/tmp/kpartx",
}

func getConfig() (*Config, error) {
	p, err := ParseConfig(storage)
	if err != nil {
		return nil, err
	}
	return p.IndexToConfig(0)
}

func TestLocalImage(t *testing.T) {
	t.Log("=> MKdirAll")
	if err := os.MkdirAll(rootfsMp, 0755); err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(rootfsMp)

	t.Log("=> getConfig")
	config, err := getConfig()
	if err != nil {
		t.Fatal(err)
	}

	for _, disk := range config.Disks {
		t.Logf("=> new disk description => %s", disk.Description)
		img, err := New(disk, rootfsMp, u, nil)
		if err != nil {
			t.Fatal(err)
		}

		// interrupt handler
		img.ReleaseOnInterrupt()

		defer func() {
			t.Log("=> CleanupPre")
			if err := img.CleanupPre(); err != nil {
				t.Fatal(err)
			}
			t.Log("=> MakeBootable")
			if err := img.MakeBootable(); err != nil {
				t.Fatal(err)
			}
			t.Log("=> Convert")
			if err := img.Convert(); err != nil {
				t.Fatal(err)
			}
			t.Log("=>CleanupPost")
			if err := img.CleanupPost(); err != nil {
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
}

func TestRemoteImage(t *testing.T) {
	t.Log("=> MkdirAll")
	if err := os.MkdirAll(rootfsMp, 0755); err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(rootfsMp)

	t.Log("=> getConfig")
	config, err := getConfig()
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

	for _, disk := range config.Disks {
		t.Logf("=> new disk description => %s", disk.Description)
		img, err := New(disk, rootfsMp, u, sshfsConf)
		if err != nil {
			t.Fatal(err)
		}
		img.ReleaseOnInterrupt()

		defer func() {
			t.Log("=> CleanupPre")
			if err := img.CleanupPre(); err != nil {
				t.Fatal(err)
			}
			t.Log("=> MakeBootable")
			if err := img.MakeBootable(); err != nil {
				t.Fatal(err)
			}
			t.Log("=> Convert")
			if err := img.Convert(); err != nil {
				t.Fatal(err)
			}
			t.Log("=>CleanupPost")
			if err := img.CleanupPost(); err != nil {
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
}
