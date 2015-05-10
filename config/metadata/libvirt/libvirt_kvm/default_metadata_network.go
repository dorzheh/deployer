package libvirt_kvm

var TmpltVirtNetwork = ` <interface type='network'>
      <source network='{{.NetworkName}}'/>
      <model type='{{.Driver}}'/>
      <driver name='vhost'/>
      <address type='pci' domain='0x{{.GuestNicDomain}}' bus='0x{{.GuestNicBus}}' slot='0x{{.GuestNicSlot}}' function='0x{{.GuestNicFunction}}'/>
    </interface>`

var TmpltBridged = `<interface type='bridge'>
      <source bridge='{{.Bridge}}'/>
	    <model type='{{.Driver}}'/>
	    <driver name='vhost'/>
      <address type='pci' domain='0x{{.GuestNicDomain}}' bus='0x{{.GuestNicBus}}' slot='0x{{.GuestNicSlot}}' function='0x{{.GuestNicFunction}}'/>
 </interface>`

var TmpltBridgedOVS = `<interface type='bridge'>
      <source bridge='{{.OVSBridge}}'/>
      <virtualport type='openvswitch'/>
	    <model type='{{.Driver}}'/>
	    <driver name='vhost'/>
      <address type='pci' domain='0x{{.GuestNicDomain}}' bus='0x{{.GuestNicBus}}' slot='0x{{.GuestNicSlot}}' function='0x{{.GuestNicFunction}}'/>
</interface>`

var TmpltDirect = `<interface type='direct'>
      <source dev='{{.IfaceName}}' mode='private'/>
      <model type='{{.Driver}}'/>
      <address type='pci' domain='0x{{.GuestNicDomain}}' bus='0x{{.GuestNicBus}}' slot='0x{{.GuestNicSlot}}' function='0x{{.GuestNicFunction}}'/>
    </interface>
`

var TmpltSriovPassthrough = `<interface type='hostdev' managed='yes'>
      <source>
      <address type='pci' domain='0x0000' bus='0x{{.HostNicBus}}' slot='0x{{.HostNicSlot}}' function='0x{{.HostNicFunction}}'/>
      </source>
      <address type='pci' domain='0x{{.GuestNicDomain}}' bus='0x{{.GuestNicBus}}' slot='0x{{.GuestNicSlot}}' function='0x{{.GuestNicFunction}}'/>
      <rom bar='off'/>
    </interface>
`
var TmpltPassthrough = `<hostdev mode='subsystem' type='pci' managed='yes'>
      <source>
      <address type='pci' domain='0x0000' bus='0x{{.HostNicBus}}' slot='0x{{.HostNicSlot}}' function='0x{{.HostNicFunction}}'/>
      </source>
      <address type='pci' domain='0x{{.GuestNicDomain}}' bus='0x{{.GuestNicBus}}' slot='0x{{.GuestNicSlot}}' function='0x{{.GuestNicFunction}}'/>
  </hostdev>
`
