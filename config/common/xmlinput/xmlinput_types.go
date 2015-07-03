package xmlinput

type ConnectionMode string

const (
	ConTypeBridged        ConnectionMode = "bridged"
	ConTypeOVS                           = "ovs"
	ConTypeDirect                        = "direct"
	ConTypePassthrough                   = "passthrough"
	ConTypeVirtualNetwork                = "virtnetwork"
	ConTypeError                         = "error"
)

const UnlimitedAlloc = -1

type XMLInputData struct {
	CPU
	RAM
	Disks
	Networks
	HostNics
	GuestNic
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
	Name           string        `xml:"name,attr"`
	UiResetCounter bool          `xml:"ui_reset_counter"`
	Modes          []*Mode       `xml:"mode"`
	UiModeBinding  []*Appearance `xml:"ui_mode_selection>appearance"`
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
	Vendor   string `xml:"vendor,attr"`
	Model    string `xml:"model,attr"`
	Priority bool   `xml:"priority,attr"`
}

type Deny struct {
	Vendor string `xml:"vendor,attr"`
	Model  string `xml:"model,attr"`
}

type HostNics struct {
	Allowed []*Allow `xml:"host_nics>allow"`
	Denied  []*Deny  `xml:"host_nics>deny"`
}

type PciAddress struct {
	Domain    string `xml:"domain"`
	Bus       string `xml:"bus"`
	FirstSlot int    `xml:"first_slot"`
	Function  string `xml:"function"`
}

type GuestNic struct {
	PCI *PciAddress `xml:"guest_nics>pci"`
}
