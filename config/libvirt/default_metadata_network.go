package libvirt

var TmpltBridged = `<interface type='bridge'>
      <source bridge='{{.Bridge}}'/>
	  <model type='virtio'/>
	  <driver name='vhost'/>
 </interface>`

var TmpltBridgedOVS = `<interface type='bridge'>
      <source bridge='{{.OVSBridge}}'/>
      <virtualport type='openvswitch'/>
	  <model type='virtio'/>
	  <driver name='vhost'/>
</interface>`

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
