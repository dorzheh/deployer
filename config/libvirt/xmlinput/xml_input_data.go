package xmlinput

import (
	"bytes"
	"encoding/xml"
	"io/ioutil"
)

type XMLInputData struct {
	CPU      `xml:"CPU"`
	RAM      `xml:"RAM"`
	Networks `xml:"Networks"`
	NICs     `xml:"NICs"`
}

type CPU struct {
	Config  bool `xml:"CPU>Config"`
	Min     uint `xml:"CPU>Min"`
	Max     uint `xml:"CPU>Max"`
	Default uint `xml:"CPU>Default"`
}

type RAM struct {
	Config  bool `xml:"RAM>Config"`
	Min     uint `xml:"RAM>Min"`
	Max     uint `xml:"RAM>Max"`
	Default uint `xml:"RAM>Default"`
}

type Network struct {
	Name string `xml:"Name"`
}

type Networks struct {
	Config  bool       `xml:"Networks>Config"`
	Max     uint       `xml:"Networks>Max"`
	Default []*Network `xml:"Networks>Default>Network"`
}

type Allow struct {
	Vendor string `xml:"Vendor"`
	Model  string `xml:"Model"`
	Mode   string `xml:"Mode"`
}

type Deny struct {
	Vendor string `xml:"Vendor"`
	Model  string `xml:"Model"`
}

type NICs struct {
	Allowed []*Allow `xml:"NICs>Allow"`
	Denied  []*Deny  `xml:"NICs>Deny"`
}

func ParseXML(xmlpath string) (*XMLInputData, error) {
	fb, err := ioutil.ReadFile(xmlpath)
	if err != nil {
		return nil, err
	}
	return Parse(fb)
}

func Parse(fb []byte) (*XMLInputData, error) {
	buf := bytes.NewBuffer(fb)
	p := new(XMLInputData)
	decoded := xml.NewDecoder(buf)
	if err := decoded.Decode(p); err != nil {
		return nil, err
	}
	return p, nil
}
