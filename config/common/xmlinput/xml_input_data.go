package xmlinput

type XMLInputData struct {
	CPU      `xml:"cpu"`
	RAM      `xml:"ram"`
	Networks `xml:"networks"`
	NICs     `xml:"nics"`
}

type CPU struct {
	Config  bool `xml:"cpu>config"`
	Min     uint `xml:"cpu>min"`
	Max     uint `xml:"cpu>max"`
	Default uint `xml:"cpu>default"`
}

type RAM struct {
	Config  bool `xml:"ram>config"`
	Min     uint `xml:"ram>min"`
	Max     uint `xml:"ram>max"`
	Default uint `xml:"ram>default"`
}

type Network struct {
	Name   string `xml:"name"`
	Driver string `xml:"vnic_driver"`
}

type Networks struct {
	Config        bool       `xml:"networks>config"`
	Max           uint       `xml:"networks>max"`
	DefaultDriver string     `xml:"networks>default>vnic_driver"`
	Default       []*Network `xml:"networks>default>network"`
}

type Allow struct {
	Vendor string `xml:"vendor"`
	Model  string `xml:"model"`
	Mode   string `xml:"mode"`
}

type Deny struct {
	Vendor string `xml:"vendor"`
	Model  string `xml:"model"`
}

type NICs struct {
	Allowed []*Allow `xml:"nics>allow"`
	Denied  []*Deny  `xml:"nics>deny"`
}
