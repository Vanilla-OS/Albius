#!/bin/sh

set -e

distrobox-create -r -I -Y -ap "golang libbtrfs-dev libdevmapper-dev libgpgme-dev build-essential pkg-config lvm2 cryptsetup parted udev" -i ghcr.io/vanilla-os/dev:main albius_test
