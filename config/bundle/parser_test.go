package bundle

import (
	"fmt"
	"testing"

	"github.com/dorzheh/deployer/builder/image"
	"github.com/dorzheh/deployer/config/xmlinput"
	"github.com/dorzheh/deployer/deployer"
)

var xmlstream = []byte(`<?xml version="1.0" encoding="UTF-8"?>
<bundle>
  	<config>
			<name>Test1</name>
			<cpus>2</cpus>
     		<ram>4096</ram>
     		<!-- storage configuration index -->
     		<storage_config_index>0</storage_config_index>
  	</config>
  	<config>
			<name>Test2</name>
			<cpus>2</cpus>
			<ram>8192</ram>
			<storage_config_index>1</storage_config_index>
  	</config>
  	<config>
			<name>Test3</name>
			<cpus>8</cpus>
     		<ram>16384</ram>
     		<storage_config_index>2</storage_config_index>
  	</config>
  	<advanced_config>true</advanced_config>
</bundle>`)

type ConfigTest struct {
	Name               string            `xml:"config>name"`
	CPUs               int               `xml:"config>cpus"`
	RAM                int               `xml:"config>ram"`
	StorageConfigIndex image.ConfigIndex `xml:"config>storage_config_index"`
}

type DefaultBundleTest struct {
	Configs        []*Config `xml:"config"`
	AdvancedConfig bool      `xml:"advanced_config"`
}

func TestParseConfig(t *testing.T) {

	p, err := NewParserBuff(xmlstream, new(DefaultBundleTest))
	if err != nil {
		t.Fatal(err)
	}

	m, err := p.Parse(nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println("=============================")
	for k, v := range m {
		fmt.Printf("%s => %v\n", k, v)
	}
	fmt.Println("=============================")

}

func (b *DefaultBundleTest) Parse(c *deployer.CommonData, d deployer.HostinfoDriver, x *xmlinput.XMLInputData) (map[string]interface{}, error) {
	first := b.Configs[0]
	if first.Name != "Test1" {
		return nil, fmt.Errorf("expected name is Test1, got %s", first.Name)
	}
	if first.CPUs != 2 {
		return nil, fmt.Errorf("expected amount of CPUs is 2, got %d", first.CPUs)
	}
	if first.RAM != 4096 {
		return nil, fmt.Errorf("expected amount of RAM is 4096, got %d", first.RAM)
	}
	if first.StorageConfigIndex != 0 {
		return nil, fmt.Errorf("expected storage configuration index is 0, got %d", first.StorageConfigIndex)
	}

	m := make(map[string]interface{})
	second := b.Configs[1]
	m["name"] = second.Name
	m["cpus"] = second.CPUs
	m["ram"] = second.RAM
	m["storage_config_index"] = second.StorageConfigIndex
	return m, nil
}
