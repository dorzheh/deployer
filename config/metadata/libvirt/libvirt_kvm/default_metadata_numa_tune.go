package libvirt_kvm

var TmpltNUMATune = `<numatune>
    <memory mode='interleave' nodeset='{{.NUMACells}}'/>
  </numatune>`
