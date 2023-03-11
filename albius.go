package main

import (
	"fmt"
	"os"

	"github.com/vanilla-os/albius/core"
)

func main() {
	// Locate disk
	disk, err := albius.LocateDisk("/dev/sda")
	if err != nil {
		panic(err)
	}
	fmt.Println(disk)

	// Partition disk
	err = disk.LabelDisk(albius.GPT)
	if err != nil {
		panic(err)
	}

	_, err = disk.NewPartition("boot", albius.EXT4, 1, 1025)
	if err != nil {
		panic(err)
	}
	_, err = disk.NewPartition("efi", albius.FAT32, 1025, 1537)
	if err != nil {
		panic(err)
	}

	_, err = disk.NewPartition("a", albius.BTRFS, 1537, 13825)
	if err != nil {
		panic(err)
	}

	_, err = disk.NewPartition("b", albius.BTRFS, 13825, 26113)
	if err != nil {
		panic(err)
	}

	_, err = disk.NewPartition("home", albius.BTRFS, 26113, -1)
	if err != nil {
		panic(err)
	}

	disk, err = albius.LocateDisk("/dev/sda")
	if err != nil {
		panic(err)
	}

	// Create all paths necessary and mount partitions
	err = os.MkdirAll("/mnt/a/boot/efi", 0644)
	if err != nil {
		panic(err)
	}

	err = os.MkdirAll("/mnt/b", 0644)
	if err != nil {
		panic(err)
	}

	err = disk.Partitions[2].Mount("/mnt/a")
	if err != nil {
		panic(err)
	}

	err = disk.Partitions[0].Mount("/mnt/a/boot")
	if err != nil {
		panic(err)
	}

	err = disk.Partitions[1].SetPartitionFlag("esp", true)
	if err != nil {
		panic(err)
	}

	err = disk.Partitions[1].Mount("/mnt/a/boot/efi")
	if err != nil {
		panic(err)
	}

	err = disk.Partitions[4].Mount("/mnt/a/home")
	if err != nil {
		panic(err)
	}

	err = disk.Partitions[3].Mount("/mnt/b")
	if err != nil {
		panic(err)
	}

	// Copy files
	err = albius.Unsquashfs("/cdrom/casper/filesystem.squashfs", "/mnt/a", true)
	if err != nil {
		panic(err)
	}

	// System setup
	partUUIDs := []string{}
	for _, part := range disk.Partitions {
		uuid, err := part.GetUUID()
		if err != nil {
			panic(err)
		}

		partUUIDs = append(partUUIDs, uuid)
	}

	fstabEntries := [][]string{
		{fmt.Sprintf("UUID=%s", partUUIDs[1]), "/boot/efi", string(disk.Partitions[1].Filesystem), "umask=0077", "0", "0"},
		{fmt.Sprintf("UUID=%s", partUUIDs[0]), "/boot", string(disk.Partitions[0].Filesystem), "noatime,errors=remount-ro", "0", "0"},
		{fmt.Sprintf("UUID=%s", partUUIDs[2]), "/", string(disk.Partitions[2].Filesystem), "defaults", "0", "0"},
		{fmt.Sprintf("UUID=%s", partUUIDs[3]), "/", string(disk.Partitions[3].Filesystem), "defaults", "0", "0"},
		{fmt.Sprintf("UUID=%s", partUUIDs[4]), "/home", string(disk.Partitions[4].Filesystem), "defaults", "0", "0"},
	}
	err = albius.GenFstab("/mnt/a", fstabEntries)
	if err != nil {
		panic(err)
	}

	err = albius.UpdateInitramfs("/mnt/a")
	if err != nil {
		panic(err)
	}

	// TODO: Create user, call abroot-adapter
}
