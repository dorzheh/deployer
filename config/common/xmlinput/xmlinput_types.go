package xmlinput

type ConnectionMode string

const (
	ConTypeBridged     ConnectionMode = "bridged"
	ConTypeDirect                     = "direct"
	ConTypePassthrough                = "passthrough"
	ConTypeError                      = "error"
)

type XMLInputData struct {
	CPU
	RAM
	Nets
	NICs
}

type CPU struct {
	Configure bool `xml:"cpu>configure"`
	Min       uint `xml:"cpu>min"`
	Max       uint `xml:"cpu>max"`
	Default   uint `xml:"cpu>default_value"`
}

type RAM struct {
	Configure bool `xml:"ram>configure"`
	Min       uint `xml:"ram>min"`
	Max       uint `xml:"ram>max"`
	Default   uint `xml:"ram>default_value"`
}

type Network struct {
	Name          string        `xml:"name,attr"`
	MaxIfaces     uint          `xml:"max_ifaces,attr"`
	Mandatory     bool          `xml:"mandatory,attr"`
	Modes         []*Mode       `xml:"mode"`
	UiModeBinding []*Appearance `xml:"ui_mode_selection>appearance"`
}

type Nets struct {
	Configure bool       `xml:"networks>configure"`
	Networks  []*Network `xml:"networks>network"`
}

type Mode struct {
	Type       ConnectionMode `xml:"type,attr"`
	VnicDriver string         `xml:"vnic_driver,attr"`
}

type Appearance struct {
	Type   ConnectionMode `xml:"mode_type,attr"`
	Appear string         `xml:"appear,attr"`
}

type Allow struct {
	Vendor string `xml:"vendor,attr"`
	Model  string `xml:"model,attr"`
}

type Deny struct {
	Vendor string `xml:"vendor,attr"`
	Model  string `xml:"model,attr"`
}

type NICs struct {
	Allowed []*Allow `xml:"nics>allow"`
	Denied  []*Deny  `xml:"nics>deny"`
}
