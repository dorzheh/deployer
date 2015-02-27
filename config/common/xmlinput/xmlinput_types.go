package xmlinput

type ConnectionMode string

const (
	ConTypeBridged     ConnectionMode = "bridged"
	ConTypeOVS                        = "ovs"
	ConTypeDirect                     = "direct"
	ConTypePassthrough                = "passthrough"
        ConTypeVirtNetwork		  = "virtnetwork"
	ConTypeError                      = "error"
)

const Unlimited = -1

type XMLInputData struct {
	CPU
	RAM
	Disks
	Networks
	NICs
}

type CPU struct {
	Configure bool `xml:"cpu>configure"`
	Min       int  `xml:"cpu>min"`
	Max       int  `xml:"cpu>max"`
	Default   int  `xml:"cpu>default_value"`
}

type RAM struct {
	Configure bool `xml:"ram>configure"`
	Min       int  `xml:"ram>min_mb"`
	Max       int  `xml:"ram>max_mb"`
	Default   int  `xml:"ram>default_value_mb"`
}

type Disk struct {
	Min     int `xml:"min_mb"`
	Max     int `xml:"max_mb"`
	Default int `xml:"default_value_mb"`
}

type Disks struct {
	Configure bool    `xml:"disks>configure"`
	Configs   []*Disk `xml:"disks>disk"`
}

type Network struct {
	Name          string        `xml:"name,attr"`
	MaxIfaces     int           `xml:"max_ifaces,attr"`
	Mandatory     bool          `xml:"mandatory,attr"`
	Modes         []*Mode       `xml:"mode"`
	UiModeBinding []*Appearance `xml:"ui_mode_selection>appearance"`
}

type Networks struct {
	Configure bool       `xml:"networks>configure"`
	Configs   []*Network `xml:"networks>network"`
}

type Template struct {
	FileName string `xml:"file_name,attr"`
	Dir      string `xml:"dir,attr"`
}

type Mode struct {
	Type       ConnectionMode `xml:"type,attr"`
	VnicDriver string         `xml:"vnic_driver,attr"`
	Tmplt      *Template      `xml:"template"`
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
