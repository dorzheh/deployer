package config

import (
	"fmt"
	"github.com/dorzheh/infra/comm/ssh"
	"os"
	"testing"
)

const tmpFile = "/tmp/lshw_cache"

func TestCpuInfoLocal(t *testing.T) {
	i, err := NewHwInfoParser(tmpFile, "", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile)

	// get HW info and write the info file
	fmt.Println("===> executing lshw locally,writing info file")
	if err := i.Parse(); err != nil {
		t.Fatal(err)
	}
	_, err = i.CpuInfo()
	fmt.Println("===> parsing info file #1")
	if err != nil {
		t.Fatal(err)
	}

	// read info file, do not run lshw
	fmt.Println("===> parsing info file #2")
	_, err = i.CpuInfo()
	if err != nil {
		t.Fatal(err)
	}
}

func TestCpuInfoRemote(t *testing.T) {
	conf := &ssh.Config{
		Host:   "127.0.0.1",
		Port:   "22",
		User:   "root",
		Passwd: "<root_password>",
	}
	i, err := NewHwInfoParser(tmpFile, "", conf)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile)

	// get HW info and write the info file
	fmt.Println("===> executing lshw locally,writing info file")
	if err := i.Parse(); err != nil {
		t.Fatal(err)
	}
	_, err = i.CpuInfo()
	fmt.Println("===> parsing info file #1")
	if err != nil {
		t.Fatal(err)
	}

	// read info file, do not run lshw
	fmt.Println("===> parsing info file #2")
	_, err = i.CpuInfo()
	if err != nil {
		t.Fatal(err)
	}
}

func TestNicsInfoLocal(t *testing.T) {
	i, err := NewHwInfoParser(tmpFile, "", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile)

	// get HW info and write the info file
	fmt.Println("===> executing lshw locally,writing info file")
	if err := i.Parse(); err != nil {
		t.Fatal(err)
	}
	_, err = i.NicsInfo()
	fmt.Println("===> parsing info file #1")
	if err != nil {
		t.Fatal(err)
	}

	// read info file, do not run lshw
	fmt.Println("===> parsing info file #2")
	info, err := i.NicsInfo()
	if err != nil {
		t.Fatal(err)
	}

	printNicInfo(info)
}

func TestNicsInfoRemote(t *testing.T) {
	conf := &ssh.Config{
		Host:   "127.0.0.1",
		Port:   "22",
		User:   "root",
		Passwd: "<root_password>",
	}
	i, err := NewHwInfoParser(tmpFile, "", conf)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile)

	// get HW info and write the info file
	fmt.Println("===> executing lshw remotely,writing info file")
	if err := i.Parse(); err != nil {
		t.Fatal(err)
	}
	_, err = i.NicsInfo()
	fmt.Println("===> parsing info file #1")
	if err != nil {
		t.Fatal(err)
	}

	// read info file, do not run lshw
	fmt.Println("===> parsing info file #2")
	info, err := i.NicsInfo()
	if err != nil {
		t.Fatal(err)
	}

	printNicInfo(info)
}

func printNicInfo(info map[NicType][]*NicInfo) {
	fmt.Println("==== Map Content ======")
	for k, v := range info {
		for _, n := range v {
			fmt.Printf("NIC type => %v\nNIC name => %v\nNIC desc => %v\nNIC driver => %v\n",
				k, n.Name, n.Desc, n.Driver)
			fmt.Println("==================")
		}
	}
}
