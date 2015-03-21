package hwinfo

import (
	"fmt"
	"os"
	"testing"

	ssh "github.com/dorzheh/infra/comm/common"
)

const tmpFile = "/tmp/lshw_cache"

var conf = &ssh.Config{
	Host:     "127.0.0.1",
	Port:     "22",
	User:     "root",
	Password: "<root_password>",
}

func TestCPUInfoLocal(t *testing.T) {
	c, err := NewCollector(nil, "", tmpFile)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile)

	// get HW info and write the info file
	fmt.Println("===> executing lshw locally,writing info file")
	if err := c.Hwinfo2Json(); err != nil {
		t.Fatal(err)
	}
	_, err = c.CPUInfo()
	fmt.Println("===> parsing info file #1")
	if err != nil {
		t.Fatal(err)
	}

	// read info file, do not run lshw
	fmt.Println("===> parsing info file #2")
	_, err = c.CPUInfo()
	if err != nil {
		t.Fatal(err)
	}

	cpus, err := c.CPUs()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("===> CPUs: %d\n", cpus)
}

func TestCPUInfoRemote(t *testing.T) {
	c, err := NewCollector(conf, "", tmpFile)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile)

	// get HW info and write the info file
	fmt.Println("===> executing lshw remotely,writing info file")
	if err := c.Hwinfo2Json(); err != nil {
		t.Fatal(err)
	}
	_, err = c.CPUInfo()
	fmt.Println("===> parsing info file #1")
	if err != nil {
		t.Fatal(err)
	}

	// read info file, do not run lshw
	fmt.Println("===> parsing info file #2")
	_, err = c.CPUInfo()
	if err != nil {
		t.Fatal(err)
	}

	cpus, err := c.CPUs()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("===> CPUs: %d\n", cpus)
}

func TestNICsInfoLocal(t *testing.T) {
	c, err := NewCollector(nil, "", tmpFile)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile)

	// get HW info and write the info file
	fmt.Println("===> executing lshw locally,writing info file")
	if err := c.Hwinfo2Json(); err != nil {
		t.Fatal(err)
	}
	_, err = c.NICInfo()
	fmt.Println("===> parsing info file #1")
	if err != nil {
		t.Fatal(err)
	}

	// read info file, do not run lshw
	fmt.Println("===> parsing info file #2")
	info, err := c.NICInfo()
	if err != nil {
		t.Fatal(err)
	}

	printNICInfo(info)
}

func TestNICInfoRemote(t *testing.T) {
	c, err := NewCollector(conf, "", tmpFile)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile)

	// get HW info and write the info file
	fmt.Println("===> executing lshw remotely,writing info file")
	if err := c.Hwinfo2Json(); err != nil {
		t.Fatal(err)
	}
	_, err = c.NICInfo()
	fmt.Println("===> parsing info file #1")
	if err != nil {
		t.Fatal(err)
	}

	// read info file, do not run lshw
	fmt.Println("===> parsing info file #2")
	info, err := c.NICInfo()
	if err != nil {
		t.Fatal(err)
	}

	printNICInfo(info)
}

func printNICInfo(info []*NIC) {
	fmt.Println("==== Slice Content ======")
	for _, n := range info {
		fmt.Printf("NIC type => %v\nNIC name => %v\nNIC PCI addr => %v\nNIC vendor => %v\nNIC model => %v\nNIC desc => %v\nNIC driver => %v\n",
			string(n.Type), n.Name, n.PCIAddr, n.Vendor, n.Model, n.Desc, n.Driver)
		fmt.Println("==================")
	}
}

func TestNUMANodesLocal(t *testing.T) {
	c, err := NewCollector(nil, "", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile)
	n, err := c.NUMAInfo()
	if err != nil {
		t.Fatal(err)
	}
	for k, v := range n {
		fmt.Printf("===> NUMA node %d => CPUs %v\n", k, v)
	}
}

func TestNUMANodesRemote(t *testing.T) {
	c, err := NewCollector(conf, "", "")
	if err != nil {
		t.Fatal(err)
	}
	n, err := c.NUMAInfo()
	if err != nil {
		t.Fatal(err)
	}
	for k, v := range n {
		fmt.Printf("===> NUMA node %d => CPUs %v\n", k, v)
	}
}

func TestRAMSizeLocal(t *testing.T) {
	c, err := NewCollector(nil, "", "")
	if err != nil {
		t.Fatal(err)
	}
	ramsize, err := c.RAMSize()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("===> local ramsize = %d ", ramsize)
}

func TestRAMSizeRemote(t *testing.T) {
	c, err := NewCollector(conf, "", "")
	if err != nil {
		t.Fatal(err)
	}
	ramsize, err := c.RAMSize()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("===> remote ramsize = %d ", ramsize)
}
