# -*- mode: ruby -*-
# vi: set ft=ruby :

$proxy = <<SCRIPT
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
echo "export GOROOT=/go" >> /home/vagrant/.bashrc
echo "export GOPATH=/gopath" >> /home/vagrant/.bashrc
echo "export PATH=/gopath/bin:/go/bin:$PATH" >> /home/vagrant/.bashrc
SCRIPT

VAGRANTFILE_API_VERSION = "2"

Vagrant.configure(VAGRANTFILE_API_VERSION) do |config|
  config.ssh.forward_agent = true
  config.vm.box = "precise64"
  config.vm.box_url = "http://files.vagrantup.com/precise64.box"
  config.vm.provision :shell, inline: $proxy
  config.vm.network :forwarded_port, host: 8080, guest: 8080
end
