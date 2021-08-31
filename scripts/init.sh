#!/usr/bin/env bash

set -e

# installing tools and golang
apt-get -q update
#apt-get -yq dist-upgrade
apt-get -yq install \
              bash-completion \
              curl \
              make \
              nano \
              tree \
              jq \
              software-properties-common \
              unzip \
              wget \
              net-tools \
              python3-pip \
              libzfslinux-dev \
              zfsutils-linux
curl -sSf https://raw.githubusercontent.com/owenthereal/goup/master/install.sh | sudo -u "$(id -nu 1000)" sh -s -- "--skip-prompt"
rm -f /home/"$(id -nu 1000)"/.bash_profile
sudo -u "$(id -nu 1000)" echo -e "\ncd $VAGRANT_DIR" >> /home/"$(id -nu 1000)"/.bashrc

# creating zpool and mariadb dataset
zpool create rootpool /dev/sdc
zfs create rootpool/maria -o mountpoint=/var/lib/mysql

# install, enable and start mariadb
apt-key adv --fetch-keys "https://mariadb.org/mariadb_release_signing_key.asc"
add-apt-repository "deb [arch=amd64] https://mirror.docker.ru/mariadb/repo/10.5/ubuntu focal main"
apt-get update
apt-get -yq install mariadb-server
sudo -u "$(id -nu 1000)" cp "$VAGRANT_DIR/configs/my.cnf" /home/"$(id -nu 1000)"/.my.cnf

# compile conductor and create systemd unit
cp "$VAGRANT_DIR/init/conductor.service" /etc/systemd/system/conductor.service
systemctl daemon-reload

# install python packages and tool
pip install --upgrade pip
pip install -r "$VAGRANT_DIR/tools/requirements.txt"
ln -s "$VAGRANT_DIR/tools/conductorctl.py" /usr/local/bin/conductorctl

# secure, stop mariadb && reboot
#cat configs/answers.txt | sudo mariadb-secure-installation
systemctl stop mariadb
systemctl reboot
