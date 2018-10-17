# -*- mode: ruby -*-
# vi: set ft=ruby :

Vagrant.configure("2") do |config|

  config.vm.box = "ubuntu/bionic64"
  config.vm.network "public_network"
  config.vm.synced_folder ".", "/home/vagrant/go/src/github.com/imulab/homelab"

  config.vm.provider "virtualbox" do |vb|
    vb.memory = "2048"
  end

  config.vm.provision "shell", inline: <<-SHELL
    apt-get update
    apt-get install -y gcc

    wget https://dl.google.com/go/go1.11.1.linux-amd64.tar.gz > /dev/null 2>&1
    tar -C /usr/local -xzf go1.11.1.linux-amd64.tar.gz

    cat <<EOF >> /home/vagrant/.profile
export GOROOT=/usr/local/go
export GOPATH=/home/vagrant/go
export PATH=$PATH:/usr/local/go/bin:$GOPATH/bin
export GO111MODULE=on
EOF

    chown -R vagrant:vagrant /home/vagrant/go
  SHELL
end
