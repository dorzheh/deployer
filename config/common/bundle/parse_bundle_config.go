// Parses image configuration file (XML)
// Configuration example:
//
//<?xml version="1.0" encoding="UTF-8"?>
//<Bundle>
//  	<Config>
//			<Name>Test1</Name>
//			<CPUs>2</CPUs>
//     		<RAM>4096<RAM>
//			<StorageConfigIndex>0<StorageConfigIndex>
//  	</Config>
//  	<Config>
//			<Name>Test2</Name>
//			<CPUs>2</CPUs>
//			<RAM>8192<RAM>
//			<StorageConfigIndex>1<StorageConfigIndex>
//  	</Config>
//  	<Config>
//			<Name>Test3</Name>
//			<CPUs>8</CPUs>
//     		<RAM>16384<RAM>
//			<StorageConfigIndex>2<StorageConfigIndex>
//  	</Config>
//<Bundle>

package bundle

import (
	"bytes"
	"encoding/xml"
	"io/ioutil"

	"github.com/dorzheh/deployer/builder/common/image"
)

type Config struct {
	Name               string            `xml:"Name"`
	CPUs               uint              `xml:"CPUs"`
	RAM                uint              `xml:"RAM"`
	StorageConfigIndex image.ConfigIndex `xml:"StorageConfigIndex"`
}

type Bundle struct {
	Configs []*Config `xml:"Config"`
}

// Parse is responsible for parsing appropriate XML file
func ParseConfigFile(xmlpath string) (*Bundle, error) {
	fb, err := ioutil.ReadFile(xmlpath)
	if err != nil {
		return nil, err
	}
	return ParseConfig(fb)
}

func ParseConfig(fb []byte) (*Bundle, error) {
	buf := bytes.NewBuffer(fb)
	b := new(Bundle)
	decoded := xml.NewDecoder(buf)
	if err := decoded.Decode(b); err != nil {
		return nil, err
	}
	return b, nil
}

// GetConfigs looking for eligable bundle
func (b *Bundle) GetConfigs(ramsizeMb uint) []*Config {
	configs := make([]*Config, 0)
	for _, c := range b.Configs {
		if c.RAM <= ramsizeMb {
			configs = append(configs, c)
		}
	}
	return configs
}
