#!/usr/bin/env bash

set -eux

export DEBIAN_FRONTEND=noninteractive

apt-get update
apt-get install -y ca-certificates curl

install -m 0755 -d /etc/apt/keyrings

distribution="$(lsb_release --short --id | awk '{ print tolower($0) }')"
distribution_release="$(lsb_release --short --codename)"

packages_installed=()
packages_removed=()
services_enabled=()

install_docker() {
  curl -fsSL "https://download.docker.com/linux/${distribution}/gpg" -o /etc/apt/keyrings/docker.asc
  chmod a+r /etc/apt/keyrings/docker.asc

  cat > /etc/apt/sources.list.d/docker.sources << EOL
Enabled: yes
Types: deb
URIs: https://download.docker.com/linux/${distribution}
Suites: ${distribution_release}
Components: stable
Signed-By: /etc/apt/keyrings/docker.asc
EOL

  packages_installed+=(docker-ce docker-compose-plugin)
  services_enabled+=(docker)
}

remove_snapd() {
  packages_removed+=(snapd)
}

install_docker
remove_snapd

apt-get update
apt-get install -y "${packages_installed[@]}"

systemctl enable "${services_enabled[@]}"
