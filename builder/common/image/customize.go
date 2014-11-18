// Responsible for customizing rootfs

package image

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/dorzheh/infra/utils/ioutils"
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
	XMLName  xml.Name     `xml:"inject_items"`
	InjItems []InjectItem `xml:"inject_item"`
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
func Customize(pathToSlash, pathToPlatformDir string) error {
	//install/deinstall appropriate packages
	pathToXml := pathToPlatformDir + "/packages.xml"
	if _, err := os.Stat(pathToXml); err == nil {
		if err := packageManip(pathToXml, pathToSlash); err != nil {
			return err
		}
	}
	// inject appropriate stuff
	pathToXml = pathToPlatformDir + "/inject_items.xml"
	if _, err := os.Stat(pathToXml); err == nil {
		if err := injectManip(pathToXml, pathToSlash); err != nil {
			return err
		}
	}
	// services manipulation
	pathToXml = pathToPlatformDir + "/services.xml"
	if _, err := os.Stat(pathToXml); err == nil {
		if err := serviceManip(pathToXml, pathToSlash); err != nil {
			return err
		}
	}
	// file content modification
	pathToXml = pathToPlatformDir + "/files_content.xml"
	if _, err := os.Stat(pathToXml); err == nil {
		if err := filesContentManip(pathToXml, pathToSlash); err != nil {
			return err
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
//<Packages>
//	<Package>
//      <Name>tunctl</Name>
//      <Type>rpm</Type>
//	<Action>install</Action>
//	<Chroot>false</Chroot>
//	</Package>
//</Packages>
func packageManip(pathToXml, pathToSlash string) error {
	// read the XML file to a buffer
	dataBuf, err := ioutil.ReadFile(pathToXml)
	if err != nil {
		return err
	}
	// parse the data structure
	pkgsStruct := Packages{}
	if err := xml.Unmarshal(dataBuf, &pkgsStruct); err != nil {
		return err
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
			return errors.New("unsupported package format")
		}
		var action string
		switch val.Action {
		case ACTION_INSTALL:
			action = ACTION_INSTALL
		case ACTION_REMOVE:
			action = ACTION_REMOVE
		default:
			return errors.New("unsupported package manip action")
		}
		if val.Chroot {
			if err := exec.Command("chroot", pathToSlash,
				pkgCmd, "-y", action, val.Name).Run(); err != nil {
				return fmt.Errorf("chroot %s %s -y %s %s", pathToSlash, pkgCmd, action, val.Name)
			}
		} else {
			if err := exec.Command(pkgCmd, "-y", action, val.Name).Run(); err != nil {
				return fmt.Errorf("%s -y %s %s", pkgCmd, action, val.Name)
			}
		}
	}
	return nil
}

// injectStuff modifies a RAW image "on-the-fly"
// by injecting appropriate stuff to the mounted vHDD
// Also eligable for a "software only" Alteon installation
// 1) it receives src and dst directories paths
// 2) it looks for a file inject.config inside the src directory
// 3) in case the file found parses it and inject appropriate stuff
//    according to the file.
// Example:
//<InjectItems>
//	<InjectItem>
//      	<Name>file1</Name>
//	 	<BkpName>file1.adcva</BkpName>
//	  	<Action>upload</Action>
//      	<Type>file</Type>
// 		<Location>/opt</Location>
//		<Permissions>0755</Permissions>
//  		<OwnerID>0</OwnerID>
//		<GroupID>0</GroupID>
//	</InjectItem>
//</InjectItems
func injectManip(pathToXml, pathToSlash string) error {
	dataBuf, err := ioutil.ReadFile(pathToXml)
	if err != nil {
		return err
	}
	itemsStruct := InjectItems{}
	if err := xml.Unmarshal(dataBuf, &itemsStruct); err != nil {
		return err
	}
	for _, val := range itemsStruct.InjItems {
		baseDir := filepath.Dir(pathToXml)
		srcPath := baseDir + "/inject/" + val.Name
		targetLocationPath := filepath.Join(pathToSlash, val.Location)
		dstPath := filepath.Join(targetLocationPath, val.Name)
		dstBkpPath := filepath.Join(targetLocationPath, val.BkpName)
		switch val.Action {
		case ACTION_REMOVE:
			if val.BkpName == "" {
				if err := ioutils.RemoveIfExists(true, dstPath); err != nil {
					return err
				}
			} else {
				if _, err := os.Stat(dstPath); err == nil {
					if err := os.Rename(dstPath, dstBkpPath); err != nil {
						return err
					}
				}
			}
		case ACTION_UPLOAD, ACTION_CREATE:
			switch val.Type {
			case ITEM_TYPE_FILE:
				if err := ioutils.CreateDirRecursively(targetLocationPath, 0755,
					val.UID, val.GID, false); err != nil {
					if err != os.ErrExist {
						return err
					}
				}
				if val.Action == ACTION_UPLOAD {
					if val.BkpName != "" {
						if _, err := os.Stat(dstPath); err == nil {
							if err := os.Rename(dstPath, dstBkpPath); err != nil {
								return err
							}
						}
					}
					if err := ioutils.CopyFile(srcPath, dstPath, 0,
						val.UID, val.GID, false); err != nil {
						return err
					}
				} else {
					fd, err := os.Create(dstPath)
					if err != nil {
						return err
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
					return err
				}
				if val.Action == ACTION_UPLOAD {
					if val.BkpName != "" {
						if _, err := os.Stat(dstPath); err == nil {
							if err := os.Rename(dstPath, dstBkpPath); err != nil {
								return err
							}
						}
					}
					if err := ioutils.CopyDir(srcPath, dstPath); err != nil {
						return err
					}
				}
			case ITEM_TYPE_LINK:
				if _, err := os.Stat(val.BkpName); err != nil {
					return err
				}
				if err := ioutils.RemoveIfExists(false, dstPath); err != nil {
					return err
				}
				if err := ioutils.CreateDirRecursively(targetLocationPath,
					val.Permissions, val.UID, val.GID, false); err != nil {
					return err
				}
				if err := os.Symlink(val.BkpName, dstPath); err != nil {
					return err
				}
			default:
				return errors.New("injectManip: configuration error - unexpected element type")
			}
		default:
			return errors.New("injectManip: configuration error - unexpected action")
		}
	}
	return nil
}

// serviceManip allows services manipulation either over chroot
// (in case we need modify service state on an off-line image) or
// without chrooting (in case we are deploying upon a running system)
// Example:
//<Services>
//	<Service>
//     		 <Name>iptables</Name>
//      	 <Type>sysv</Type>
//      	 <Status>off</Status>
//		 <Action></Action>
//	</Service>
//	<Service>
//     		 <Name>ip6tables</Name>
//      	 <Type>sysv</Type>
//      	<Status>off</Status>
//		<Action></Action>
//		<Chroot>false</Chroot>
//	</Service>
//	<Service>
//     		 <Name>ssh</Name>
//      	<Type>upstart</Type>
//      	<Status></Status>
//		<Action>reload</Action>
//		<Chroot>false</Chroot>
//	</Service>
//</Services>
func serviceManip(pathToXml, pathToSlash string) error {
	dataBuf, err := ioutil.ReadFile(pathToXml)
	if err != nil {
		return err
	}
	servicesStruct := Services{}
	if err := xml.Unmarshal(dataBuf, &servicesStruct); err != nil {
		return err
	}
	for _, val := range servicesStruct.Srvcs {
		// switch the service type
		switch val.Type {
		case SVC_TYPE_SYSV:
			switch val.Status {
			case SVC_STATUS_ON, SVC_STATUS_OFF:
				if val.Chroot {
					if err := exec.Command("chroot", pathToSlash,
						"chkconfig", "--list", val.Name).Run(); err == nil {
						if err := exec.Command("chroot", pathToSlash,
							"chkconfig", val.Name, val.Status).Run(); err != nil {
							return fmt.Errorf("chroot %s chkconfig %s %s", pathToSlash,
								val.Name, val.Status)
						}
					}
				} else {
					if err := exec.Command("chkconfig", "--list",
						val.Name).Run(); err == nil {
						if err := exec.Command("chkconfig", val.Name,
							val.Status).Run(); err != nil {
							return fmt.Errorf("chkconfig %s %s", val.Name, val.Status)
						}
					}
				}
			default:
				return errors.New(`ServiceManip :sysv:status configuration error - unsupported service status`)
			}
			// switch appropriate action towards the service
			switch val.Action {
			case ACTION_STOP, ACTION_START, ACTION_RESTART, ACTION_RELOAD:
				if err := exec.Command("service",
					val.Name, val.Action).Run(); err != nil {
					return fmt.Errorf("service %s %s", val.Name, val.Action)
				}
			case "":
			default:
				return errors.New(`ServiceManip :sysv:action: configuration error - unsupported action ` + val.Action)
			}

		case SVC_TYPE_UPSTART:
			fullPathToChrootDir := filepath.Join(pathToSlash, "/etc/init/")
			servicePath := filepath.Join(fullPathToChrootDir, val.Name+".conf")
			dummyServicePath := filepath.Join(fullPathToChrootDir, val.Name+".override")
			switch val.Status {
			case SVC_STATUS_OFF:
				if _, err := os.Stat(servicePath); err == nil {
					if err := ioutil.WriteFile(dummyServicePath, []byte("manual"), 0644); err != nil {
						return err
					}
				}
			case SVC_STATUS_ON:
				if err := ioutils.RemoveIfExists(false, dummyServicePath); err != nil {
					return err
				}
			default:
				return errors.New(`configuration error - unsupported service status`)
			}
			switch val.Action {
			case ACTION_STOP, ACTION_START, ACTION_RESTART, ACTION_RELOAD:
				if err := exec.Command("initctl", val.Name,
					val.Action).Run(); err != nil {
					return fmt.Errorf("initctl %s %s", val.Name, val.Action)
				}
			case "":
			default:
				return errors.New(`ServiceManip : upstart :configuration error - unsupported action`)
			}
		}
	}
	return nil
}

// filesContentManip manipulates with the content of the files
// according to appropriate XML configuration file
// Example:
//<Files>
//	<File>
//		<Path>/etc/sysconfig/selinux</Path>
//		<BkpName>/etc/sysconfig/selinux.adcva</BkpName>
//		<Action>replace</Action>
//		<OldPattern>SELINUX=\S+</OldPattern>
//		<NewPattern>SELINUX=disabled</NewPattern>
//	</File>
//	<File>
//		<Path>/etc/passwd</Path>
//		<BkpName>/etc/passwd.bak</BkpName>
//		<Action>append</Action>
//		<OldPattern></OldPattern>
//		<NewPattern>test:x:111:111::/root:/bin/bash</NewPattern>
//	</File>
//</Files>
func filesContentManip(pathToXml, pathToSlash string) error {
	dataBuf, err := ioutil.ReadFile(pathToXml)
	if err != nil {
		return err
	}
	fileContentStruct := FilesContent{}
	if err := xml.Unmarshal(dataBuf, &fileContentStruct); err != nil {
		return err
	}
	for _, val := range fileContentStruct.FContent {
		targetPath := filepath.Join(pathToSlash, val.Path)
		if err != nil {
			return err
		}
		finfo, err := os.Stat(targetPath)
		if err != nil {
			continue
		}
		if val.NewPattern == "" {
			return errors.New("configuration error - NewPattern is empty")
		}
		if val.BkpName != "" {
			bkpFilePath := filepath.Join(pathToSlash, val.BkpName)
			if err := ioutils.CopyFile(targetPath, bkpFilePath, 0, -1, -1, false); err != nil {
				return err
			}
		}
		fd, err := os.OpenFile(targetPath, os.O_RDWR|os.O_APPEND, finfo.Mode())
		if err != nil {
			return err
		}
		defer fd.Close()

		switch val.Action {
		// if we need to append to the file
		case ACTION_APPEND:
			if err := ioutils.AppendToFd(fd, val.NewPattern, val.NewPattern); err != nil {
				return err
			}
		// if we need to replace a pattern
		case ACTION_REPLACE:
			if val.OldPattern == "" {
				return errors.New("configuration error - replace action is set but OldPattern is empty")
			}
			if err := ioutils.FindAndReplaceFd(fd, val.OldPattern, val.NewPattern); err != nil {
				return err
			}
		default:
			return errors.New(`FilesContentManip:configuration error - unsupported action`)
		}
	}
	return nil
}
