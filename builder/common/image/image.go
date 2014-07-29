// Responsible for creating , partitioning and customizing RAW image
// and writing MBR thus making the image bootable

package image

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/dorzheh/deployer/utils"
	"github.com/dorzheh/infra/comm/sshfs"
	infrautils "github.com/dorzheh/infra/utils"
)

// the structure represents a mapper device
type mapperDevice struct {
	// device name(full path)
	name string

	// mount point the device is mounted
	mountPoint string
}

type loopDevice struct {
	// device name(full path)
	name string

	// slice of mapperDevice pointers
	mappers []*mapperDevice

	// counter - amount of mappers the loop device contains
	amountOfMappers uint8
}

// the structure represents loop device manipulation
type image struct {
	// path to the image
	imgpath string

	//
	localmount string

	// path to sirectory intended for mounting the image
	slashpath string

	// platform configuration file (image config)
	conf *Topology

	// do we need to partition and format the image
	needToFormat bool

	// loop device structure
	*loopDevice

	// Is this a remote image
	remote bool

	// executes commands locally or remotely
	run func(string) (string, error)

	// sshfs client
	client *sshfs.Client
}

//// Public methods ////

// New gets a path to configuration directory
// path to temporary directory where the vHDD image supposed to be mounted
// and path to vHDD image.
// Returns a pointer to the structure and error/nil
func New(pathToRawImage, rootfsMp string, imageConfig *Topology, rconf *sshfs.Config) (i *image, err error) {
	needToFormat := false
	i = new(image)
	if rconf == nil {
		i.run = utils.RunFunc(nil)
		i.slashpath = rootfsMp
		i.remote = false
	} else {
		i.run = utils.RunFunc(rconf.Common)
		i.client, err = sshfs.NewClient(rconf)
		if err != nil {
			return
		}
		i.localmount = rootfsMp
		i.remote = true
	}
	i.imgpath = pathToRawImage
	if _, err = i.run("ls" + pathToRawImage); err != nil {
		if err = i.create(imageConfig.HddSizeGb); err != nil {
			return
		}
		needToFormat = true
	}
	i.conf = imageConfig
	i.needToFormat = needToFormat
	i.loopDevice = &loopDevice{}
	i.loopDevice.amountOfMappers = 0
	return
}

// parse processes the RAW image
// Returns error/nil
func (i *image) Parse() error {
	var err error
	if i.loopDevice.name, err = i.bind(); err != nil {
		return err
	}
	if i.needToFormat {
		if err := i.partTableMakefs(); err != nil {
			return err
		}
	} else {
		if err := i.addMappers(); err != nil {
			return err
		}
	}
	if i.remote {
		if i.slashpath, err = i.run("mktemp -d --suffix _deployer_rootfs"); err != nil {
			return err
		}
		if err := i.client.Attach(i.slashpath, i.localmount); err != nil {
			return err
		}
	}
	return nil
}

// Customize intended for the target customization
// - pathToPlatformDir - path to directory containing platform configuration XML file
func (i *image) Customize(pathToPlatformDir string) error {
	if i.amountOfMappers == 0 {
		return fmt.Errorf("amount of mappers is 0.Seems you didn't call Parse().")
	}
	return Customize(i.slashpath, pathToPlatformDir)
}

// Release is trying to release the image correctly
// by unmounting the mappers in reverse order
// and the cleans up a temporary stuff
// Returns error or nil
func (i *image) Release() error {
	if i.remote {
		if err := i.client.Detach(i.localmount); err != nil {
			return err
		}
	}
	// Release registered mount points
	var index uint8 = i.loopDevice.amountOfMappers - 1
	for i.loopDevice.amountOfMappers != 0 {
		mounted, err := i.mounted(i.loopDevice.mappers[index].name, i.loopDevice.mappers[index].mountPoint)
		if err != nil {
			return err
		}
		if mounted {
			if out, err := i.run(fmt.Sprintf("umount -l %s", i.loopDevice.mappers[index].mountPoint)); err != nil {
				return fmt.Errorf("%s [%v]", out, err)
			}
			i.loopDevice.amountOfMappers--
		}
	}
	// unbind mappers and image
	if out, err := i.run("kpartx -d " + i.loopDevice.name); err != nil {
		return fmt.Errorf("%s [%v]", out, err)
	}
	if out, err := i.run("losetup -d " + i.loopDevice.name); err != nil {
		return fmt.Errorf("%s [%v]", out, err)
	}
	// remove mount point
	if out, err := i.run("rm -rf " + i.slashpath); err != nil {
		return fmt.Errorf("%s [%v]", out, err)
	}
	return nil
}

// MakeBootable is responsible for making RAW disk bootable.
// The target disk could be either local or remote image
func (i *image) MakeBootable(localPathToGrubBin string) error {
	var pathToGrubBin string
	if i.remote {
		remoteMp, err := i.run("mktemp -d --suffix _deployer_bin")
		if err != nil {
			return fmt.Errorf("%s [%s]", remoteMp, err)
		}

		localMp, err := ioutil.TempDir("", "deployer_bin_")
		if err != nil {
			return err
		}
		if err := i.client.Attach(remoteMp, localMp); err != nil {
			return err
		}
		defer func() {
			i.client.Detach(localMp)
			i.run("rm -rf " + remoteMp)
			os.RemoveAll(localMp)
		}()

		pathToGrubBin = filepath.Join(remoteMp, filepath.Base(localPathToGrubBin))
		if err := infrautils.CopyFile(localPathToGrubBin, pathToGrubBin, 0755, 0, 0, false); err != nil {
			return err
		}
	} else {
		pathToGrubBin = localPathToGrubBin
	}

	cmd := fmt.Sprintf("echo -e \"device (hd0) %s\nroot (hd0,0)\nsetup (hd0)\n\"|%s",
		i.imgpath, pathToGrubBin)
	if out, err := i.run(cmd); err != nil {
		return fmt.Errorf("%s [%v]", out, err)
	}
	return nil
}

// ReleaseOnInterrupt is trying to release appropriate image
// in case SIGHUP, SIGINT or SIGTERM signal received
func (i *image) ReleaseOnInterrupt() {
	//create a channel for interrupt handler
	interrupt := make(chan os.Signal, 1)
	// create an interrupt handler
	signal.Notify(interrupt, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)
	// run a seperate goroutine.
	go func() {
		for {
			select {
			case <-interrupt:
				i.Release()
			}
		}
	}()
}

// Exports amount of mappers to the outside of the class
func (i *image) AmountOfMappers() uint8 {
	return i.amountOfMappers
}

/// Private stuff ///

func (i *image) bind() (loopDevice string, err error) {
	loopDevice, err = i.run("losetup -f")
	if err != nil {
		return
	}
	cmd := fmt.Sprintf("losetup %s %s", loopDevice, i.imgpath)
	if out, er := i.run(cmd); err != nil {
		err = fmt.Errorf("%s [%v]", out, er)
		return
	}
	return
}

// partTable creates partition table on the RAW disk
func (i *image) partTableMakefs() error {
	i.run(fmt.Sprintf("echo -e  \"%s\"|%s %s", i.conf.FdiskCmd, "fdisk", i.loopDevice.name))
	mappers, err := i.getMappers(i.loopDevice.name)
	if err != nil {
		return err
	}
	if err := i.validatePconf(len(mappers)); err != nil {
		return err
	}
	for index, part := range i.conf.Partitions {
		mapper := mappers[index]
		// create SWAP and do not add to the mappers slice
		if strings.ToLower(part.FileSystem) == "swap" {
			if out, err := i.run(fmt.Sprintf("mkswap -L %s %s", part.Label, mapper)); err != nil {
				return fmt.Errorf("%s [%v]", out, err)
			}
		} else {
			cmd := fmt.Sprintf("mkfs -t %v -L %s %s %s", part.FileSystem,
				part.Label, part.FileSystemArgs, mapper)
			if out, err := i.run(cmd); err != nil {
				return fmt.Errorf("%s [%v]", out, err)
			}
			if err := i.addMapper(mapper, part.MountPoint); err != nil {
				return err
			}
		}
	}
	return nil
}

// addMappers registers appropriate mappers
func (i *image) addMappers() error {
	mappers, err := i.getMappers(i.loopDevice.name)
	if err != nil {
		return err
	}
	if err := i.validatePconf(len(mappers)); err != nil {
		return err
	}
	for index, part := range i.conf.Partitions {
		mapper := mappers[index]
		// create SWAP and do not add to the mappers slice
		if strings.ToLower(part.FileSystem) != "swap" {
			if err := i.addMapper(mapper, part.MountPoint); err != nil {
				return err
			}
		}
	}
	return nil
}

// addMapper registers appropriate mapper and it's mount point
func (i *image) addMapper(mapperDeviceName, path string) error {
	mountPoint := filepath.Join(i.slashpath, path)
	if out, err := i.run("mkdir -p " + mountPoint); err != nil {
		return fmt.Errorf("%s [%v]", out, err)
	}
	// check if the volume is already mounted
	mounted, err := i.mounted(mapperDeviceName, mountPoint)
	if err != nil {
		return err
	}
	if !mounted {
		if out, err := i.run(fmt.Sprintf("mount %s %s", mapperDeviceName, mountPoint)); err != nil {
			return fmt.Errorf("%s [%v]", out, err)
		}
	}
	// add mapper
	i.mappers = append(i.mappers,
		&mapperDevice{
			name:       mapperDeviceName,
			mountPoint: mountPoint,
		},
	)
	// advance amount of mappers
	i.loopDevice.amountOfMappers++
	return nil
}

// validatePconf is responsible for platform configuration file validation
func (i *image) validatePconf(amountOfMappers int) error {
	if amountOfMappers != len(i.conf.Partitions) {
		return fmt.Errorf("amount of partitions defined = %v, actual amount is %v",
			len(i.conf.Partitions), amountOfMappers)
	}
	return nil
}

// create is intended for creating RAW image
func (i *image) create(hddSize int) error {
	out, err := i.run(fmt.Sprintf("dd if=/dev/zero of=%s count=1 bs=1 seek=%vG", i.imgpath, hddSize))
	if err != nil {
		return fmt.Errorf("%s [%v]", out, err)
	}
	return nil
}

// getMappers is responsible for finding mappers bound to appropriate loop device
// and providing the stuff as a slice
func (i *image) getMappers(loopDeviceName string) ([]string, error) {
	var mappers []string
	if _, err := i.run("kpartx -a " + loopDeviceName); err != nil {
		return mappers, err
	}
	// somehow on RHEL based systems refresh might take some time therefore
	// no mappers are available until then
	duration, err := time.ParseDuration("1s")
	if err != nil {
		return mappers, err
	}
	time.Sleep(duration)

	cmd := fmt.Sprintf("find /dev -name loop%sp[0-9]",
		strings.TrimSpace(strings.SplitAfter(loopDeviceName, "/dev/loop")[1]))
	out, err := i.run(cmd)
	if err != nil {
		return mappers, fmt.Errorf("%s [%v]", out, err)
	}
	for _, line := range strings.Split(out, "\n") {
		if line != "" {
			mappers = append(mappers, line)
		}
	}
	if len(mappers) == 0 {
		return mappers, errors.New("mappers not found")
	}
	sort.Strings(mappers)
	return mappers, nil
}

// check mountinfo

const (
	mountinfoFormat = "%d %d %d:%d %s %s %s"
)

type procEntry struct {
	id, parent, major, minor int
	source, mountpoint, opts string
}

func (i *image) parseMountTable() ([]*procEntry, error) {
	out, err := i.run("cat /proc/self/mountinfo")
	if err != nil {
		return nil, err
	}
	entries := []*procEntry{}
	p := &procEntry{}
	for _, line := range strings.Split(out, "\n") {
		if _, err := fmt.Sscanf(line, mountinfoFormat,
			&p.id, &p.parent, &p.major, &p.minor,
			&p.source, &p.mountpoint, &p.opts); err != nil {
			return nil, fmt.Errorf("Scanning '%s' failed: %s", line, err)
		}
		entries = append(entries, p)
	}
	return entries, nil
}

// Looks at /proc/self/mountinfo to determine of the specified
// mountpoint has been mounted
func (i *image) mounted(device, mountpoint string) (bool, error) {
	entries, err := i.parseMountTable()
	if err != nil {
		return false, err
	}
	// Search the table for the mountpoint
	for _, entry := range entries {
		if entry.mountpoint == mountpoint || strings.Contains(entry.opts, device) {
			return true, nil
		}
	}
	return false, nil
}
