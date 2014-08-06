package utils

import (
	"fmt"
	ssh "github.com/dorzheh/infra/comm/common"
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
	info, err := i.CpuInfo()
	if err != nil {
		t.Fatal(err)
	}

	fmt.Printf("===> CPUs: %d\n", info.Cpus)
}

func TestCpuInfoRemote(t *testing.T) {
	conf := &ssh.Config{
		Host:     "127.0.0.1",
		Port:     "22",
		User:     "root",
		Password: "<root_password>",
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
	_, err = i.NicsInfo(nil)
	fmt.Println("===> parsing info file #1")
	if err != nil {
		t.Fatal(err)
	}

	// read info file, do not run lshw
	fmt.Println("===> parsing info file #2")
	info, err := i.NicsInfo(nil)
	if err != nil {
		t.Fatal(err)
	}

	printNicInfo(info)
}

func TestNicsInfoRemote(t *testing.T) {
	conf := &ssh.Config{
		Host:     "127.0.0.1",
		Port:     "22",
		User:     "root",
		Password: "<root_password>",
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
	_, err = i.NicsInfo(nil)
	fmt.Println("===> parsing info file #1")
	if err != nil {
		t.Fatal(err)
	}

	// read info file, do not run lshw
	fmt.Println("===> parsing info file #2")
	info, err := i.NicsInfo(nil)
	if err != nil {
		t.Fatal(err)
	}

	printNicInfo(info)
}

func printNicInfo(info []*NicInfo) {
	fmt.Println("==== Slice Content ======")
	for _, n := range info {
		fmt.Printf("NIC type => %v\nNIC name => %v\nNIC PCI addr => %v\nNIC desc => %v\nNIC driver => %v\n",
			string(n.Type), n.Name, n.PCIAddr, n.Desc, n.Driver)
		fmt.Println("==================")
	}
}

func TestRAMSizeLocal(t *testing.T) {
	i, err := NewHwInfoParser("", "", nil)
	if err != nil {
		t.Fatal(err)
	}
	ramsize, err := i.RAMSize()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("===> local ramsize = %d ", ramsize)
}

func TestRAMSizeRemote(t *testing.T) {
	conf := &ssh.Config{
		Host:     "127.0.0.1",
		Port:     "22",
		User:     "root",
		Password: "<root_password>",
	}
	i, err := NewHwInfoParser("", "", conf)
	if err != nil {
		t.Fatal(err)
	}
	ramsize, err := i.RAMSize()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("===> remote ramsize = %d ", ramsize)
}
