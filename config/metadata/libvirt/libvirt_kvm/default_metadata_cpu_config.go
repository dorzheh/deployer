package libvirt_kvm

var TmpltCpuConfig = `<cpu mode='custom' match='exact'>
    <model fallback='allow'>Westmere</model>
    <feature policy='require' name='pdpe1gb'/>
    {{.NUMAConfig}}
  </cpu>`
