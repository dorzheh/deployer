# -*- mode: ruby -*-
# vi: set ft=ruby :

# Vagrantfile API/syntax version. Don't touch unless you know what you're doing!
VAGRANTFILE_API_VERSION = '2'

@script = <<SCRIPT
GOVERSION="1.5.3"
SRCROOT="/opt/go"
SRCPATH="/opt/gopath"

# Get the ARCH
ARCH=`uname -m | sed 's|i686|386|' | sed 's|x86_64|amd64|'`

# Install Go
sudo apt-get update
sudo apt-get install -y build-essential git-core zip curl

# Install Go
cd /tmp
curl -s -O https://storage.googleapis.com/golang/go${GOVERSION}.linux-${ARCH}.tar.gz
tar -xvf go${GOVERSION}.linux-${ARCH}.tar.gz
sudo mv go $SRCROOT
sudo chmod 775 $SRCROOT
sudo chown vagrant:vagrant $SRCROOT

# Setup the GOPATH; even though the shared folder spec gives the consul
# directory the right user/group, we need to set it properly on the
# parent path to allow subsequent "go get" commands to work. We can't do
# normal -R here because VMWare complains if we try to update the shared
# folder permissions, so we just update the folders that matter.
sudo mkdir -p $SRCPATH
find /opt/gopath -type d -maxdepth 3 | xargs sudo chown vagrant:vagrant

cat <<EOF >/tmp/gopath.sh
export GOPATH="$SRCPATH"
export GOROOT="$SRCROOT"
export PATH="$SRCROOT/bin:$SRCPATH/bin:\$PATH"
EOF
sudo mv /tmp/gopath.sh /etc/profile.d/gopath.sh
sudo chmod 0755 /etc/profile.d/gopath.sh
source /etc/profile.d/gopath.sh
echo "source /etc/profile.d/gopath.sh" | sudo tee -a /root/.bashrc

# Install godep
go get github.com/tools/godep
cd /opt/gopath/src/github.com/dorzheh/deployer
godep restore
ln -s  /opt/gopath/src/github.com/dorzheh/deployer/example/myproduct/ /home/vagrant/example
cd /home/vagrant/
[ -f  /home/vagrant/deployer ] && rm -f /home/vagrant/deployer
go build  /home/vagrant/example/deployer/deployer.go
echo "example" > .product
SCRIPT

Vagrant.configure(VAGRANTFILE_API_VERSION) do |config|
  config.vm.provision 'shell', inline: @script, privileged: false
  config.vm.synced_folder '.', '/opt/gopath/src/github.com/dorzheh/deployer'

  %w[vmware_fusion vmware_workstation].each do |p|
    config.vm.provider p do |v|
      v.vmx['memsize'] = '1024'
      v.vmx['numvcpus'] = '2'
      v.vmx['cpuid.coresPerSocket'] = '1'
    end
  end

  config.vm.provider "virtualbox" do |v|
    v.memory = 1024
    v.cpus = 2
  end

  config.vm.define '64bit' do |n1|
    n1.vm.box = 'hashicorp/precise64'
  end
end
