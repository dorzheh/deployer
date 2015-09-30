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

	// Responsible for executing hooks
	// The method might be usefull for cases
	// when rootfs postprocessing is required
	RunHooks(string) error
}
