# -*- mode: ruby -*-
# vi: set ft=ruby :

$bootstrap = <<SCRIPT
apt-key adv --keyserver hkp://keyserver.ubuntu.com:80 --recv 7F0CEB10
echo 'deb http://downloads-distro.mongodb.org/repo/ubuntu-upstart dist 10gen' | tee /etc/apt/sources.list.d/mongodb.list
apt-get update
apt-get install -y bzr mercurial git-core build-essential mongodb-10gen

mkdir -p /gopath/src/github.com/vokalinteractive
ln -s /vagrant /gopath/src/github.com/vokalinteractive/vip 

wget -c https://go.googlecode.com/files/go1.2.linux-amd64.tar.gz
tar -xzvf go1.2.linux-amd64.tar.gz
rm *.tar.gz
mv go /go

chown -R vagrant:vagrant /gopath
chown -R vagrant:vagrant /go

echo 'export GOROOT=/go' >> /home/vagrant/.profile
echo 'export GOPATH=/gopath' >> /home/vagrant/.profile
echo 'export PATH=$GOPATH/bin:$GOROOT/bin:$PATH' >> /home/vagrant/.profile
SCRIPT

VAGRANTFILE_API_VERSION = "2"

Vagrant.configure(VAGRANTFILE_API_VERSION) do |config|
  config.vm.box = "phusion-open-ubuntu-12.04-amd64"
  config.vm.box_url = "https://oss-binaries.phusionpassenger.com/vagrant/boxes/ubuntu-12.04.3-amd64-vbox.box"

  config.vm.provider :vmware_fusion do |f, override|
      override.vm.box_url = "https://oss-binaries.phusionpassenger.com/vagrant/boxes/ubuntu-12.04.3-amd64-vmwarefusion.box"
  end

  if Dir.glob("#{File.dirname(__FILE__)}/.vagrant/machines/default/*/id").empty?
      # Install Docker
      pkg_cmd = "wget -q -O - https://get.docker.io/gpg | apt-key add -;" \
          "echo deb http://get.docker.io/ubuntu docker main > /etc/apt/sources.list.d/docker.list;" \
          "apt-get update -qq; apt-get install -q -y --force-yes lxc-docker; "
      # Add vagrant user to the docker group
      pkg_cmd << "usermod -a -G docker vagrant; "
      config.vm.provision :shell, :inline => pkg_cmd
  end

  config.vm.provision :shell, inline: $bootstrap

  config.vm.network :forwarded_port, host: 8080, guest: 8080
end
