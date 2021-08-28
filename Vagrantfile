ENV['VAGRANT_DISABLE_VBOXSYMLINKCREATE'] = "1"
ENV['VAGRANT_EXPERIMENTAL'] = "disks"

Vagrant.configure("2") do |config|
  config.vm.box = "ubuntu/focal64"
  config.vm.box_version = "20210709.0.0"
  config.vm.disk :disk, name: "zed", size: "4GB"
  config.vm.network "forwarded_port", guest: 8080, host: 8080
  config.vm.provider "virtualbox" do |v|
    v.memory = 2048
    v.cpus = 4
  end
  config.vm.provision "shell",
    path: "scripts/init.sh",
    env: {
      "VAGRANT_DIR" => "/vagrant"
    }
end
