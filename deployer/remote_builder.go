package deployer

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/dorzheh/infra/comm/ssh"
	"github.com/dorzheh/infra/image"
	"github.com/dorzheh/infra/utils"
)

type RemoteImageBuilder struct {
	ConnFunc    func() (*ssh.SshConn, error)
	SshConfig   *ssh.Config
	ImagePath   string
	RootfsMp    string
	ImageConfig *image.Topology
	Filler      image.Rootfs
}

func (b *RemoteImageBuilder) Run() (Artifact, error) {
	sshfs, err := exec.LookPath("sshfs")
	if err != nil {
		return nil, err
	}
	fusermount, err := exec.LookPath("fusermount")
	if err != nil {
		return nil, err
	}

	println(sshfs, fusermount)

	conn, err := b.ConnFunc()
	if err != nil {
		return nil, err
	}
	defer conn.ConnClose()

	s, err := ProcessTemplate(hddBuildScript, b.ImageConfig)
	if err != nil {
		return nil, err
	}

	cmd := fmt.Sprintf("echo %s > /tmp/run_script;bash -x /tmp/run_script", s)
	if _, stderr, err := conn.Run(cmd); err != nil {
		return nil, fmt.Errorf("%s [%s]", stderr, err)
	}

	if err := utils.CreateDirRecursively(b.RootfsMp, 0755, 0, 0, false); err != nil {
		return nil, err
	}
	defer os.RemoveAll(b.RootfsMp)

	//argsPat := fmt.Sprintf("%s@%s:%s %s -o %s,idmap=user,compression=no,nonempty,Ciphers=arcfour",
	//	s.account.name, s.ip, remoteShare, localMount, mo)
	//if out, err := exec.Command("sshfs", strings.Fields(argsPat)...).CombinedOutput(); err != nil {
	//	return fmt.Errorf("%s:%s", out, err)
	//}

	//img, err := image.New(b.ImagePath, b.RootfsMp, b.ImageConfig)
	//if err != nil {
	//	return nil, err
	//}
	//if err := img.Parse(); err != nil {
	//	return nil, err
	//}
	// interrupt handler
	//img.ReleaseOnInterrupt()
	//defer func() {
	//	if err := img.Release(); err != nil {
	//		panic(err)
	//	}
	//}()
	// create and customize rootfs
	if b.Filler != nil {
		if err := b.Filler.MakeRootfs(b.RootfsMp); err != nil {
			return nil, err
		}
		// install application.
		if err := b.Filler.InstallApp(b.RootfsMp); err != nil {
			return nil, err
		}
	}
	return &RemoteArtifact{
		Name: b.ImageConfig.Name,
		Path: b.ImagePath,
		Type: ImageArtifact,
	}, nil
}

var hddBuildScript = `
  dd if=/dev/zero of={{ .ImagePath}} bs=1 count=1 seek={{ .HddSize }}
  loop_device=$(losetup -f)
  losetup $loop_device {{ .ImagePath }}
  echo -e $ | fdisk $loop_device || true
  kpartx -av $loop_device

  loop_name=${loop_device##/dev/}
  loop_root_partition=${free_loop_name}p1
  local loop_swap_partition=${free_loop_name}p2

  [ ! -e /dev/mapper/$loop_root_partition ] && {
   echo "FATAL: loop device doesn't have any partitons"
   return 1
  }

  mkfs.ext4 /dev/mapper/$loop_root_partition
  e2fsck -y -f /dev/mapper/$loop_root_partition
  e2label /dev/mapper/$loop_root_partition SLASH
  mkswap -L SWAP /dev/mapper/$loop_swap_partition
`

type RemoteMetadataBuilder struct {
	ConnFunc func() (*ssh.SshConn, error)
	Source   string
	Dest     string
	UserData interface{}
}

func (b *RemoteMetadataBuilder) Run() (Artifact, error) {
	f, err := ioutil.ReadFile(b.Source)
	if err != nil {
		return nil, err
	}
	data, err := ProcessTemplate(string(f), b.UserData)
	if err != nil {
		return nil, err
	}
	conn, err := b.ConnFunc()
	if err != nil {
		return nil, err
	}
	cmd := fmt.Sprintf("echo %s > %s", data, b.Dest)
	if _, stderr, err := conn.Run(cmd); err != nil {
		return nil, fmt.Errorf("%s [%s]", stderr, err)
	}
	return &RemoteArtifact{
		Name: "",
		Path: b.Dest,
		Type: MetadataArtifact,
	}, nil
}

// sshfs me@x.x.x.x:/remote/path /local/path/ -o IdentityFile=/path/to/key
