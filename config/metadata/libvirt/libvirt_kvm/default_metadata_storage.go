package libvirt_kvm

var TmpltStorage = `<disk type='file' device='disk'>
	<driver name='qemu' type='{{.StorageType}}' cache='none'/>
	<source file='{{.ImagePath}}'/>
	<target dev='vd{{.BlockDeviceSuffix}}' bus='virtio'/>
	</disk>
`
