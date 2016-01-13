package libvirt_kvm

var TmpltNUMATune = `<numatune>
    <memory mode='strict' nodeset='{{.NUMACells}}'/>
    {{.MemNodes}}
  </numatune>`
