package remote

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/dorzheh/deployer/deployer"
	ssh "github.com/dorzheh/infra/comm/common"
	"github.com/dorzheh/infra/comm/gssh/sshfs"
	"github.com/dorzheh/infra/utils"
)

type ImageBuilder struct {
	*deployer.ImageBuilderData
	SshConfig           *ssh.Config
	BuildScriptPath     string
	BuildScriptUserData interface{}
}

func (b *ImageBuilder) Id() string {
	return "RemoteImageBuilder"
}

func (b *ImageBuilder) Run() (deployer.Artifact, error) {
	f, err := ioutil.ReadFile(b.BuildScriptPath)
	if err != nil {
		return nil, err
	}
	script, err := deployer.ProcessTemplate(string(f), b.BuildScriptUserData)
	if err != nil {
		return nil, err
	}

	run := deployer.RunFunc(b.SshConfig)
	cmd := fmt.Sprintf("echo %s > /tmp/run_script;sudo bash -x /tmp/run_script;rm -f  /tmp/run_script", script)
	if _, err := run(cmd); err != nil {
		return nil, err
	}

	if err := utils.CreateDirRecursively(b.RootfsMp, 0755, 0, 0, false); err != nil {
		return nil, err
	}
	defer os.RemoveAll(b.RootfsMp)

	conf := &sshfs.Config{
		Common:      b.SshConfig,
		SshfsPath:   "",
		FusrmntPath: "",
	}

	s, err := sshfs.NewClient(conf)
	if err != nil {
		return nil, err
	}
	if err := s.Attach("/dummy", b.RootfsMp); err != nil {
		return nil, err
	}
	defer s.Detach(b.RootfsMp)

	if b.Filler != nil {
		if err := b.Filler.MakeRootfs(b.RootfsMp); err != nil {
			return nil, err
		}
		// install application.
		if err := b.Filler.InstallApp(b.RootfsMp); err != nil {
			return nil, err
		}
	}
	return &deployer.RemoteArtifact{
		Name: b.ImageConfig.Name,
		Path: b.ImagePath,
		Type: deployer.ImageArtifact,
	}, nil
}

type MetadataBuilder struct {
	*deployer.MetadataBuilderData
	SshConfig *ssh.Config
}

func (b *MetadataBuilder) Id() string {
	return "RemoteMetadataBuilder"
}

func (b *MetadataBuilder) Run() (deployer.Artifact, error) {
	f, err := ioutil.ReadFile(b.Source)
	if err != nil {
		return nil, err
	}
	data, err := deployer.ProcessTemplate(string(f), b.UserData)
	if err != nil {
		return nil, err
	}

	run := deployer.RunFunc(b.SshConfig)
	cmd := fmt.Sprintf("echo %s > %s", data, b.Dest)
	if _, err := run(cmd); err != nil {
		return nil, err
	}
	return &deployer.RemoteArtifact{
		Name: filepath.Base(b.Dest),
		Path: b.Dest,
		Type: deployer.MetadataArtifact,
	}, nil
}

//var getLoopDevice = `#!/bin/env bash
//dd if=/dev/zero of={{ .PathToImage }} bs=1 count=1 seek={{ .HddSizeGb }}G
//loop_device=""
//for ((i=0;i<10;i++));do
//    loop_device=$(losetup -f)
//    losetup $loop_device $path_to_image && break
//    sleep 2
// done
// echo -e "{{ .FdiskCmd }}" | fdisk $loop_device
//  kpartx -av $loop_device

//  local free_loop_name=${loop_device##/dev/}
//  local loop_root_partition=${free_loop_name}p1
//  local loop_swap_partition=${free_loop_name}p2

//  [ ! -e /dev/mapper/$loop_root_partition ] && {
//   echo "FATAL: loop device doesn't have any partitons"
//   return 1
//  }

//  mkfs.ext4 /dev/mapper/$loop_root_partition
//  e2fsck -y -f /dev/mapper/$loop_root_partition
//  e2label /dev/mapper/$loop_root_partition SLASH
//  mkswap -L SWAP /dev/mapper/$loop_swap_partition
// __populate_loop_dev $loop_root_partition $type
//  $TOOLS_DIR/bin/grub-legacy/grub --device-map=/dev/null << EOF
//  device (hd0) $path_to_image
//  root (hd0,0)
//  setup (hd0)
//EOF`
