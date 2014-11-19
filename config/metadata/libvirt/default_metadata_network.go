package libvirt

var TmpltBridged = `<interface type='bridge'>
      <source bridge='{{.Bridge}}'/>
	  <model type='{{.Driver}}'/>
	  <driver name='vhost'/>
 </interface>`

var TmpltBridgedOVS = `<interface type='bridge'>
      <source bridge='{{.OVSBridge}}'/>
      <virtualport type='openvswitch'/>
	  <model type='virtio'/>
	  <driver name='{{.Driver}}'/>
</interface>`

var TmpltDirect = `<interface type='direct'>
      <source dev='{{.IfaceName}}' mode='private'/>
      <model type='{{.Driver}}'/>
    </interface>
`

var TmpltSriovPassthrough = `<interface type='hostdev' managed='yes'>
      <source>
        <address type='pci' domain='0x0000' bus='0x{{.Bus}}' slot='0x{{.Slot}}' function='0x{{.Function}}'/>
      </source>
    </interface>
`
var TmpltPassthrough = `<hostdev mode='subsystem' type='pci' managed='yes'>
    <source>
      <address type='pci' domain='0x0000' bus='0x{{.Bus}}' slot='0x{{.Slot}}' function='0x{{.Function}}'/>
    </source>
  </hostdev>
`
