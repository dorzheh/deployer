package libvirt

var TmpltStorage = `<disk type='file' device='disk'>
	<driver name='qemu' type='raw' cache='none'/>
	<source file='{{.ImagePath}}'/>
	<target dev='vd{{.BlockDeviceSuffix}}' bus='virtio'/>
	</disk>
`
