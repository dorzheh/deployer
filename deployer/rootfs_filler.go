package deployer

// Implementers of Rootfs are responsible for populating rootfs
// and installing stuff belonging to application.
type RootfsFiller interface {
	// Responsible for extracting/populting rootfs.
	// Receives rootfs mount point.
	CustomizeRootfs(string) error

	// Responsible for application installation.
	// Receives rootfs mount point.
	InstallApp(string) error
}
