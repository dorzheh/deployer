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

func printNicInfo(info map[int]*NicInfo) {
	fmt.Println("==== Map Content ======")
	for k, v := range info {
		fmt.Printf("Index: %d\nNIC type => %v\nNIC name => %v\nNIC PCI addr => %v\nNIC desc => %v\nNIC driver => %v\n",
			k, string(v.Type), v.Name, v.PCIAddr, v.Desc, v.Driver)
		fmt.Println("==================")
	}
}

func TestRAMsizeLocal(t *testing.T) {
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
	_, err = i.RAMSize()
	fmt.Println("===> parsing info file #1")
	if err != nil {
		t.Fatal(err)
	}

	// read info file, do not run lshw
	fmt.Println("===> parsing info file #2")
	var ramsize uint
	if ramsize, err = i.RAMSize(); err != nil {
		t.Fatal(err)
	}
	fmt.Printf("===> ramsize = %d ", ramsize)
}

func TestRAMsizeRemote(t *testing.T) {
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
	_, err = i.RAMSize()
	fmt.Println("===> parsing info file #1")
	if err != nil {
		t.Fatal(err)
	}

	// read info file, do not run lshw
	fmt.Println("===> parsing info file #2")
	var ramsize uint
	if ramsize, err = i.RAMSize(); err != nil {
		t.Fatal(err)
	}
	fmt.Printf("===> ramsize = %d ", ramsize)
}
