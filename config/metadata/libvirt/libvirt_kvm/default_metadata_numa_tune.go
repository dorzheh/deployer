package libvirt_kvm

var TmpltNUMATune = `<numatune>
    <memory mode='{{.MemMode}}' nodeset='{{.NUMACells}}'/>
    {{.MemNodes}}
  </numatune>`
