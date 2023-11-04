#!/bin/sh

set -e

distrobox-create -r -I -ap "golang libbtrfs-dev libdevmapper-dev libgpgme-dev build-essential pkg-config lvm2 parted udev" -i ghcr.io/vanilla-os/vso:main albius_test