#!/usr/bin/env bash

set -eux

clean_cloud_init() {
  cloud-init clean --logs --machine-id --seed --configs all

  rm -rf /run/cloud-init/*
  rm -rf /var/lib/cloud/*
}

clean_apt() {
  export DEBIAN_FRONTEND=noninteractive

  apt-get -y autopurge
  apt-get -y clean

  rm -rf /var/lib/apt/lists/*
}

clean_ssh_keys() {
  rm -f /etc/ssh/ssh_host_*_key /etc/ssh/ssh_host_*_key.pub
}

clean_logs() {
  journalctl --flush
  journalctl --rotate --vacuum-time=0

  find /var/log -type f -exec truncate --size 0 {} \; # truncate system logs
  find /var/log -type f -name '*.[1-9]' -delete # remove archived logs
  find /var/log -type f -name '*.gz' -delete # remove compressed archived logs
}

clean_root() {
  unset HISTFILE

  rm -rf /root/.cache
  rm -rf /root/.ssh
  rm -f /root/.bash_history
  rm -f /root/.lesshst
  rm -f /root/.viminfo
}

flush_disk() {
  fstrim --all || true
  sync
}

clean_cloud_init
clean_apt
clean_ssh_keys
clean_logs
clean_root

flush_disk
