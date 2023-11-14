package albius

import (
	"encoding/json"
	"fmt"
	"strconv"
)

const (
	MSDOS = "msdos"
	GPT   = "gpt"
)

type DiskLabel string

type Sector struct {
	Start, End int
}

type Disk struct {
	Path, Size, Model, Transport string
	Label                        DiskLabel
	LogicalSectorSize            int `json:"logical-sector-size"`
	PhysicalSectorSize           int `json:"physical-sector-size"`
	MaxPartitions                int `json:"max-partitions"`
	Partitions                   []Partition
}

func (disk *Disk) AvailableSectors() ([]Sector, error) {
	sectors := []Sector{}

	for i, part := range disk.Partitions {
		endInt, err := strconv.Atoi(part.End[:len(part.End)-3])
		if err != nil {
			return []Sector{}, fmt.Errorf("failed to retrieve end position of partition: %s", err)
		}

		if i < len(disk.Partitions)-1 {
			nextStart := disk.Partitions[i+1].Start
			nextStartInt, err := strconv.Atoi(nextStart[:len(nextStart)-3])
			if err != nil {
				return []Sector{}, fmt.Errorf("failed to retrieve start position of next partition: %s", err)
			}

			if endInt != nextStartInt {
				sectors = append(sectors, Sector{endInt, nextStartInt})
			}
		}
	}

	// Handle empty space after last partition
	lastPartitionEndStr := disk.Partitions[len(disk.Partitions)-1].End
	lastPartitionEnd, err := strconv.Atoi(lastPartitionEndStr[:len(lastPartitionEndStr)-3])
	if err != nil {
		return []Sector{}, fmt.Errorf("failed to retrieve end position of last partition: %s", err)
	}
	diskEnd, err := strconv.Atoi(disk.Size[:len(disk.Size)-3])
	if err != nil {
		return []Sector{}, fmt.Errorf("failed to retrieve disk end")
	}
	if lastPartitionEnd < diskEnd {
		sectors = append(sectors, Sector{lastPartitionEnd, diskEnd})
	}

	return sectors, nil
}

func LocateDisk(diskname string) (*Disk, error) {
	findPartitionCmd := "parted -sj %s unit MiB print"
	output, err := OutputCommand(fmt.Sprintf(findPartitionCmd, diskname))
	// If disk is unformatted, parted returns the expected json but also throws an error.
	// We can assume we have all the necessary information if output isn't empty.
	if err != nil && output == "" {
		return nil, fmt.Errorf("failed to list disk: %s", err)
	}

	var decoded struct {
		Disk Disk
	}
	err = json.Unmarshal([]byte(output), &decoded)
	if err != nil {
		return nil, fmt.Errorf("could not find device %s", diskname)
	}

	for i := 0; i < len(decoded.Disk.Partitions); i++ {
		decoded.Disk.Partitions[i].FillPath(decoded.Disk.Path)
	}

	return &decoded.Disk, nil
}

func (disk *Disk) Update() error {
	updatedInfo, err := LocateDisk(disk.Path)
	if err != nil {
		return err
	}

	disk.Path = updatedInfo.Path
	disk.Size = updatedInfo.Size
	disk.Transport = updatedInfo.Transport
	disk.Label = updatedInfo.Label
	disk.LogicalSectorSize = updatedInfo.LogicalSectorSize
	disk.PhysicalSectorSize = updatedInfo.PhysicalSectorSize
	disk.MaxPartitions = updatedInfo.MaxPartitions
	disk.Partitions = updatedInfo.Partitions

	return nil
}

func (disk *Disk) LabelDisk(label DiskLabel) error {
	labelDiskCmd := "parted -s %s mklabel %s"

	for _, part := range disk.Partitions {
		if err := part.UnmountPartition(); err != nil {
			return fmt.Errorf("failed to unmount partition %s: %s", part.Path, err)
		}
	}

	err := RunCommand(fmt.Sprintf(labelDiskCmd, disk.Path, label))
	if err != nil {
		return fmt.Errorf("failed to label disk: %s", err)
	}

	return nil
}

// NewPartition creates a new partition on Disk with the provided name,
// filesystem type, and start and end locations.
//
// If fsType is an empty string, the function will skip creating the filesystem.
// This can be useful when creating LUKS-encrypted partitions, where the format
// operation needs to be executed first.
func (target *Disk) NewPartition(name string, fsType PartitionFs, start, end int) (*Partition, error) {
	createPartCmd := "parted -s %s unit MiB mkpart%s%s %s %d %s"

	var partType string
	if target.Label == MSDOS {
		partType = " primary"
	} else {
		partType = ""
	}

	var endStr string
	if end == -1 {
		endStr = "100%"
	} else {
		endStr = fmt.Sprint(end)
	}

	partName := ""
	if name != "" {
		partName = fmt.Sprintf(" \"%s\"", name)
	}

	err := RunCommand(fmt.Sprintf(createPartCmd, target.Path, partType, partName, fsType, start, endStr))
	if err != nil {
		return nil, fmt.Errorf("failed to create partition: %s", err)
	}

	// Update partition list because we made changes to the disk
	err = target.Update()
	if err != nil {
		return nil, fmt.Errorf("failed to create partition: %s", err)
	}

	newPartition := &target.Partitions[len(target.Partitions)-1]
	newPartition.FillPath(target.Path)

	// Create filesystem
	if fsType != "" {
		newPartition.Filesystem = fsType
		err := MakeFs(newPartition)
		if err != nil {
			return nil, err
		}

		// Label partition with name
		err = newPartition.SetLabel(name)
		if err != nil {
			return nil, err
		}
	}

	if name != "" {
		err = newPartition.NamePartition(name)
		if err != nil {
			return nil, err
		}
	}

	return newPartition, nil
}
