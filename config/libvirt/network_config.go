package libvirt

var Bridged = `<interface type='bridge'>
      <source bridge='{{ .Bridge }}'/>
 </interface>`

var BridgedOVS = `<interface type='bridge'>
      <source bridge='{{ .OVSBridge }}'/>
      <virtualport type='openvswitch'/>
</interface>`

var Passthrough = `<interface type='hostdev' managed='yes'>
      <source>
        <address type='pci' domain='0x0000' bus='0x{{ .Bus }}' slot='0x{{. Slot }}' function='0x{{ .Function }}'/>
      </source>
    </interface>
`

//type Networks struct {

//    map[string]*utils.NicInfo

//type InterfaceData struct {

//}
