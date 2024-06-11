#!/usr/bin/env bash

set -eux

export DEBIAN_FRONTEND=noninteractive

apt-get update
apt-get upgrade -y
apt-get autopurge -y
