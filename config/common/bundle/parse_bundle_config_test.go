package bundle

import (
	"fmt"
	"testing"
)

var xmlstream = []byte(`<?xml version="1.0" encoding="UTF-8"?>
<Bundle>
  	<Config>
			<Name>Test1</Name>
			<CPUs>2</CPUs>
     		<RAM>4096</RAM>
     		<!-- storage configuration index -->
     		<StorageConfigIndex>0</StorageConfigIndex>
  	</Config>
  	<Config>
			<Name>Test2</Name>
			<CPUs>2</CPUs>
			<RAM>8192</RAM>
			<StorageConfigIndex>1</StorageConfigIndex>
  	</Config>
  	<Config>
			<Name>Test3</Name>
			<CPUs>8</CPUs>
     		<RAM>16384</RAM>
     		<StorageConfigIndex>2</StorageConfigIndex>
  	</Config>
</Bundle>`)

func TestParseConfig(t *testing.T) {
	b, err := ParseConfig(xmlstream)
	if err != nil {
		t.Fatal(err)
	}

	first := b.Configs[0]
	if first.Name != "Test1" {
		t.Fatalf("expected name is Test1, got %s", first.Name)
	}
	if first.CPUs != 2 {
		t.Fatalf("expected amount of CPUs is 2, got %d", first.CPUs)
	}
	if first.RAM != 4096 {
		t.Fatalf("expected amount of RAM is 4096, got %d", first.RAM)
	}
	if first.StorageConfigIndex != 0 {
		t.Fatalf("expected storage configuration index is 0, got %d", first.StorageConfigIndex)
	}
	for _, conf := range b.Configs {
		fmt.Printf("Name => %s , CPUs => %d , RAM => %d, StorageConfigIndex => %d\n",
			conf.Name, conf.CPUs, conf.RAM, conf.StorageConfigIndex)
	}
}
