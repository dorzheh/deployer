// Responsible for customizing rootfs

package content

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"

	"github.com/dorzheh/deployer/utils"
	"github.com/dorzheh/infra/utils/ioutils"
)

const SWAP_LABEL = "SWAP"
const SLASH = "/"

const (
	ACTION_CREATE  = "create"
	ACTION_UPLOAD  = "upload"
	ACTION_REMOVE  = "remove"
	ACTION_START   = "start"
	ACTION_STOP    = "stop"
	ACTION_RESTART = "restart"
	ACTION_RELOAD  = "reload"
	ACTION_APPEND  = "append"
	ACTION_REPLACE = "replace"
	ACTION_INSTALL = "install"
)

const (
	SVC_STATUS_ON  = "on"
	SVC_STATUS_OFF = "off"
)

const (
	SVC_TYPE_SYSV    = "sysv"
	SVC_TYPE_UPSTART = "upstart"
)

const (
	PKG_TYPE_RPM = "rpm"
	PKG_TYPE_DEB = "deb"
)

const (
	ITEM_TYPE_FILE = "file"
	ITEM_TYPE_DIR  = "directory"
	ITEM_TYPE_LINK = "link"
)

const (
	PRE_SCRIPTS = iota
	POST_SCRIPTS
)

const (
	INJ_TYPE_FILE = iota
	INJ_TYPE_DIR
	INJ_TYPE_QUIT
)

// Represents Item to inject
type InjectItem struct {
	Name        string      `xml:"name"`
	BkpName     string      `xml:"bkp_name"`
	Action      string      `xml:"action"`
	Type        string      `xml:"type"`
	Location    string      `xml:"location"`
	Permissions os.FileMode `xml:"permissions"`
	UID         int         `xml:"owner_id"`
	GID         int         `xml:"group_id"`
}

// Represents a slice of Items for injection
type InjectItems struct {
	XMLName  xml.Name     `xml:"items"`
	InjItems []InjectItem `xml:"item"`
}

// Represents services
type Service struct {
	Name   string `xml:"name"`
	Type   string `xml:"type"`
	Action string `xml:"action"`
	Status string `xml:"status"`
	Chroot bool   `xml:"chroot"`
}

// Represents a slice of services
type Services struct {
	XMLName xml.Name  `xml:"services"`
	Srvcs   []Service `xml:"service"`
}

// Represents packages
type Package struct {
	Name   string `xml:"name"`
	Type   string `xml:"type"`
	Action string `xml:"action"`
	Chroot bool   `xml:"chroot"`
}

// Represents a slice of packages
type Packages struct {
	XMLName xml.Name  `xml:"packages"`
	Pkgs    []Package `xml:"package"`
}

// Represents a file content
type FileContent struct {
	Path       string `xml:"path"`
	BkpName    string `xml:"bkp_name"`
	Action     string `xml:"action"`
	OldPattern string `xml:"old_pattern"`
	NewPattern string `xml:"new_pattern"`
}

// Represents a slice of files to modify
type FilesContent struct {
	XMLName  xml.Name      `xml:"files"`
	FContent []FileContent `xml:"file"`
}

// ImageCustomize treating image customization according to XML config files
// Returns error or nil
func Customize(pathToSlash, pathToConfigDir string) error {
	//install/deinstall appropriate packages
	pathToXml := pathToConfigDir + "/packages.xml"
	if _, err := os.Stat(pathToXml); err == nil {
		if err := packageManip(pathToXml, pathToSlash); err != nil {
			return utils.FormatError(err)
		}
	}
	// inject appropriate stuff
	pathToXml = pathToConfigDir + "/inject_items.xml"
	if _, err := os.Stat(pathToXml); err == nil {
		if err := injectManip(pathToXml, pathToSlash); err != nil {
			return utils.FormatError(err)
		}
	}
	// services manipulation
	pathToXml = pathToConfigDir + "/services.xml"
	if _, err := os.Stat(pathToXml); err == nil {
		if err := serviceManip(pathToXml, pathToSlash); err != nil {
			return utils.FormatError(err)
		}
	}
	// file content modification
	pathToXml = pathToConfigDir + "/files_content.xml"
	if _, err := os.Stat(pathToXml); err == nil {
		if err := filesContentManip(pathToXml, pathToSlash); err != nil {
			return utils.FormatError(err)
		}
	}
	return nil
}

// packageManip is intended for installing or removind appropriate
// packages to the offline appliance
// TODO: currently supported while deploying on a 3d party appliance.
// Make it to support manipulation
// over chroot
// Example:
//<packages>
//	<package>
//     <name>tunctl</name>
//     <type>rpm</type>
//	   <action>install</action>
//	   <chroot>false</chroot>
//	</package>
//</packages>
func packageManip(pathToXml, pathToSlash string) error {
	// read the XML file to a buffer
	dataBuf, err := ioutil.ReadFile(pathToXml)
	if err != nil {
		return utils.FormatError(err)
	}
	// parse the data structure
	pkgsStruct := Packages{}
	if err := xml.Unmarshal(dataBuf, &pkgsStruct); err != nil {
		return utils.FormatError(err)
	}
	var pkgCmd string
	// iterate over the slice and treat each entry (package)
	for _, val := range pkgsStruct.Pkgs {
		switch val.Type {
		case PKG_TYPE_RPM:
			pkgCmd = "yum"
		case PKG_TYPE_DEB:
			pkgCmd = "apt-get"
		default:
			return utils.FormatError(errors.New("unsupported package format"))
		}
		var action string
		switch val.Action {
		case ACTION_INSTALL:
			action = ACTION_INSTALL
		case ACTION_REMOVE:
			action = ACTION_REMOVE
		default:
			return utils.FormatError(errors.New("unsupported package manip action"))
		}
		if val.Chroot {
			if err := exec.Command("chroot", pathToSlash,
				pkgCmd, "-y", action, val.Name).Run(); err != nil {
				return utils.FormatError(fmt.Errorf("chroot %s %s -y %s %s", pathToSlash, pkgCmd, action, val.Name))
			}
		} else {
			if err := exec.Command(pkgCmd, "-y", action, val.Name).Run(); err != nil {
				return utils.FormatError(fmt.Errorf("%s -y %s %s", pkgCmd, action, val.Name))
			}
		}
	}
	return nil
}

// injectStuff modifies a RAW image "on-the-fly"
// by injecting appropriate stuff to the mounted vHDD
// 1) it receives src and dst directories paths
// 2) it looks for a file inject.config inside the src directory
// 3) in case the file found parses it and inject appropriate stuff
//    according to the file.
// Example:
//<inject_items>
//	<inject_item>
//      <name>file1</name>
//	 	<bkp_name>file1.bkp</bkp_name>
//	  	<action>upload</action>
//      <type>file</type>
// 		<location>/opt</location>
//		<permissions>0755</permissions>
//  	<owner_id>0</owner_id>
//		<group_id>0</group_id>
//	</inject_item>
//</inject_items
func injectManip(pathToXml, pathToSlash string) error {
	dataBuf, err := ioutil.ReadFile(pathToXml)
	if err != nil {
		return utils.FormatError(err)
	}
	itemsStruct := InjectItems{}
	if err := xml.Unmarshal(dataBuf, &itemsStruct); err != nil {
		return utils.FormatError(err)
	}
	for _, val := range itemsStruct.InjItems {
		baseDir := filepath.Dir(pathToXml)
		srcPath := baseDir + "/items/" + val.Name
		targetLocationPath := filepath.Join(pathToSlash, val.Location)
		dstPath := filepath.Join(targetLocationPath, val.Name)
		dstBkpPath := filepath.Join(targetLocationPath, val.BkpName)
		switch val.Action {
		case ACTION_REMOVE:
			if val.BkpName == "" {
				if err := ioutils.RemoveIfExists(true, dstPath); err != nil {
					return utils.FormatError(err)
				}
			} else {
				if _, err := os.Stat(dstPath); err == nil {
					if err := os.Rename(dstPath, dstBkpPath); err != nil {
						return utils.FormatError(err)
					}
				}
			}
		case ACTION_UPLOAD, ACTION_CREATE:
			switch val.Type {
			case ITEM_TYPE_FILE:
				if err := ioutils.CreateDirRecursively(targetLocationPath, 0755,
					val.UID, val.GID, false); err != nil {
					if err != os.ErrExist {
						return utils.FormatError(err)
					}
				}
				if val.Action == ACTION_UPLOAD {
					if val.BkpName != "" {
						if _, err := os.Stat(dstPath); err == nil {
							if err := os.Rename(dstPath, dstBkpPath); err != nil {
								return utils.FormatError(err)
							}
						}
					}
					if err := ioutils.CopyFile(srcPath, dstPath, 0,
						val.UID, val.GID, false); err != nil {
						return utils.FormatError(err)
					}
				} else {
					fd, err := os.Create(dstPath)
					if err != nil {
						return utils.FormatError(err)
					}
					fd.Close()
				}
				// FIXME:
				// we shuld use val.Permissions for setting permissions ,example
				// if err := CopyFile(srcFilePath, dstFilePath,
				//	val.Permissions, val.UID, val.GID); err != nil {
				//	return err
				//}
				// Currently , it copies permission flags from the source
			case ITEM_TYPE_DIR:
				if err := ioutils.CreateDirRecursively(filepath.Join(targetLocationPath, val.Name),
					val.Permissions, val.UID, val.GID, false); err != nil {
					return utils.FormatError(err)
				}
				if val.Action == ACTION_UPLOAD {
					if val.BkpName != "" {
						if _, err := os.Stat(dstPath); err == nil {
							if err := os.Rename(dstPath, dstBkpPath); err != nil {
								return utils.FormatError(err)
							}
						}
					}
					if err := ioutils.CopyDir(srcPath, dstPath); err != nil {
						return utils.FormatError(err)
					}
				}
			case ITEM_TYPE_LINK:
				if _, err := os.Stat(val.BkpName); err != nil {
					return utils.FormatError(err)
				}
				if err := ioutils.RemoveIfExists(false, dstPath); err != nil {
					return utils.FormatError(err)
				}
				if err := ioutils.CreateDirRecursively(targetLocationPath,
					val.Permissions, val.UID, val.GID, false); err != nil {
					return utils.FormatError(err)
				}
				if err := os.Symlink(val.BkpName, dstPath); err != nil {
					return utils.FormatError(err)
				}
			default:
				return utils.FormatError(errors.New("injectManip: configuration error - unexpected element type"))
			}
		default:
			return utils.FormatError(errors.New("injectManip: configuration error - unexpected action"))
		}
	}
	return nil
}

// serviceManip allows services manipulation either over chroot
// (in case we need modify service state on an off-line image) or
// without chrooting (in case we are deploying upon a running system)
// Example:
//<services>
//	<service>
//     <name>iptables</name>
//     <type>sysv</type>
//     <status>off</status>
//	   <action></action>
//	</service>
//	<service>
//     <name>ip6tables</name>
//     <type>sysv</type>
//     <status>off</status>
//	   <action></action>
//	   <chroot>false</chroot>
//	</service>
//	<service>
//     <name>ssh</name>
//     <type>upstart</type>
//     <status></status>
//	   <action>reload</action>
//	   <chroot>false</chroot>
//	</service>
//</services>
func serviceManip(pathToXml, pathToSlash string) error {
	dataBuf, err := ioutil.ReadFile(pathToXml)
	if err != nil {
		return utils.FormatError(err)
	}
	servicesStruct := Services{}
	if err := xml.Unmarshal(dataBuf, &servicesStruct); err != nil {
		return utils.FormatError(err)
	}
	for _, val := range servicesStruct.Srvcs {
		// switch the service type
		switch val.Type {
		case SVC_TYPE_SYSV:
			switch val.Status {
			case SVC_STATUS_ON, SVC_STATUS_OFF:
				var cmd string
				var action string
				if val.Chroot {
					if err := exec.Command("chroot", pathToSlash, "which", "update-rc.d").Run(); err != nil {
						cmd = "chkconfig"
						action = val.Status
					} else {
						cmd = "update-rc.d"
						switch val.Status {
						case SVC_STATUS_ON:
							action = "enable"
						case SVC_STATUS_OFF:
							action = "disable"
						}
					}
					if err := exec.Command("chroot", pathToSlash, cmd, val.Name, action).Run(); err != nil {
						return utils.FormatError(fmt.Errorf("chroot %s %s %s %s", pathToSlash, cmd, val.Name, action))
					}
				} else {
					if err := exec.Command(cmd, val.Name, action).Run(); err != nil {
						return utils.FormatError(fmt.Errorf("%s %s %s", cmd, val.Name, action))
					}
				}
			default:
				return utils.FormatError(errors.New(`ServiceManip :sysv:status configuration error - unsupported service status`))
			}
			// switch appropriate action towards the service
			switch val.Action {
			case ACTION_STOP, ACTION_START, ACTION_RESTART, ACTION_RELOAD:
				if err := exec.Command("service",
					val.Name, val.Action).Run(); err != nil {
					return utils.FormatError(fmt.Errorf("service %s %s", val.Name, val.Action))
				}
			case "":
			default:
				return utils.FormatError(errors.New(`ServiceManip :sysv:action: configuration error - unsupported action ` + val.Action))
			}

		case SVC_TYPE_UPSTART:
			fullPathToChrootDir := filepath.Join(pathToSlash, "/etc/init/")
			servicePath := filepath.Join(fullPathToChrootDir, val.Name+".conf")
			dummyServicePath := filepath.Join(fullPathToChrootDir, val.Name+".override")
			switch val.Status {
			case SVC_STATUS_OFF:
				if _, err := os.Stat(servicePath); err == nil {
					if err := ioutil.WriteFile(dummyServicePath, []byte("manual"), 0644); err != nil {
						return utils.FormatError(err)
					}
				}
			case SVC_STATUS_ON:
				if err := ioutils.RemoveIfExists(false, dummyServicePath); err != nil {
					return utils.FormatError(err)
				}
			default:
				return utils.FormatError(errors.New(`configuration error - unsupported service status`))
			}
			switch val.Action {
			case ACTION_STOP, ACTION_START, ACTION_RESTART, ACTION_RELOAD:
				if err := exec.Command("initctl", val.Name,
					val.Action).Run(); err != nil {
					return utils.FormatError(fmt.Errorf("initctl %s %s", val.Name, val.Action))
				}
			case "":
			default:
				return utils.FormatError(errors.New(`ServiceManip : upstart :configuration error - unsupported action`))
			}
		}
	}
	return nil
}

// filesContentManip manipulates with the content of the files
// according to appropriate XML configuration file
// Example:
//<files>
//	<file>
//		<path>/etc/sysconfig/selinux</path>
//		<bkp_name>/etc/sysconfig/selinux.bkp</bkp_name>
//		<action>replace</action>
//		<old_pattern>SELINUX=\S+</old_pattern>
//		<new_pattern>SELINUX=disabled</new_pattern>
//	</file>
//	<file>
//		<path>/etc/passwd</path>
//		<bkp_name>/etc/passwd.bak</bkp_name>
//		<action>append</action>
//		<old_pattern></old_pattern>
//		<new_pattern>test:x:111:111::/root:/bin/bash</new_pattern>
//	</file>
//</files>
func filesContentManip(pathToXml, pathToSlash string) error {
	dataBuf, err := ioutil.ReadFile(pathToXml)
	if err != nil {
		return utils.FormatError(err)
	}
	fileContentStruct := FilesContent{}
	if err := xml.Unmarshal(dataBuf, &fileContentStruct); err != nil {
		return utils.FormatError(err)
	}
	for _, val := range fileContentStruct.FContent {
		targetPath := filepath.Join(pathToSlash, val.Path)
		if err != nil {
			return utils.FormatError(err)
		}
		finfo, err := os.Stat(targetPath)
		if err != nil {
			continue
		}
		if val.NewPattern == "" {
			return utils.FormatError(errors.New("configuration error - NewPattern is empty"))
		}
		if val.BkpName != "" {
			bkpFilePath := filepath.Join(pathToSlash, val.BkpName)
			if err := ioutils.CopyFile(targetPath, bkpFilePath, 0, -1, -1, false); err != nil {
				return utils.FormatError(err)
			}
		}
		fd, err := os.OpenFile(targetPath, os.O_RDWR|os.O_APPEND, finfo.Mode())
		if err != nil {
			return utils.FormatError(err)
		}
		defer fd.Close()

		switch val.Action {
		// if we need to append to the file
		case ACTION_APPEND:
			if err := ioutils.AppendToFd(fd, val.NewPattern+"\n", val.NewPattern); err != nil {
				return utils.FormatError(err)
			}
		// if we need to replace a pattern
		case ACTION_REPLACE:
			if val.OldPattern == "" {
				return utils.FormatError(errors.New("configuration error - replace action is set but OldPattern is empty"))
			}
			if err := ioutils.FindAndReplaceFd(fd, val.OldPattern, val.NewPattern); err != nil {
				return utils.FormatError(err)
			}
		default:
			return utils.FormatError(errors.New(`FilesContentManip:configuration error - unsupported action`))
		}
	}
	return nil
}

// ProcessHooks is intended for executing appropriate hooks.
// The hooks must contain the following prefix : [0-9]+_
// If arguments are being passed to the function, they will be
// passed tp the hooks as well
// Example:
// 01_pre_deploy
// 02_deploy
// 03_post_deploy
// 04_clean
func ProcessHooks(pathToHooksDir string, hookArgs ...string) error {
	d, err := os.Stat(pathToHooksDir)
	if err != nil {
		return utils.FormatError(err)
	}
	if !d.IsDir() {
		return utils.FormatError(fmt.Errorf("%s is not directory", d.Name()))
	}

	var scriptsSlice []string
	//find mapped loop device partition , create appropriate mount point for each partition
	err = filepath.Walk(pathToHooksDir, func(scriptName string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			if found, _ := regexp.MatchString("[0-9]+_", scriptName); found {
				scriptsSlice = append(scriptsSlice, scriptName)
			}
		}
		return nil
	})
	sort.Strings(scriptsSlice)
	for _, file := range scriptsSlice {
		if err := exec.Command(file, hookArgs[0:]...).Run(); err != nil {
			return utils.FormatError(err)
		}
	}
	return nil
}
