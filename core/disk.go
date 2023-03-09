package albius

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

type Sector struct {
	Start, End int
}

type Disk struct {
	Path, Size, Model, Transport, Label                  string
	LogicalSectorSize, PhysicalSectorSize, MaxPartitions int
	Partitions                                           []Partition
}

func (disk *Disk) AvailableSectors() []Sector {
	sectors := []Sector{}
	latestEnd := 0

	for i, part := range disk.Partitions {
		endInt, err := strconv.Atoi(part.End[:len(part.End)-1])
		if err != nil {
			panic("Failed to retireve end position of partition")
		}

		if i < len(disk.Partitions)-1 {
			nextStart := disk.Partitions[i+1].Start
			nextStartInt, err := strconv.Atoi(part.Start[:len(nextStart)-1])
			if err != nil {
				panic("Failed to retireve end position of next partition")
			}
			sectors = append(sectors, Sector{latestEnd + 1, nextStartInt - 1})
			latestEnd = endInt + disk.PhysicalSectorSize
		}
	}

	// Handle empty space after last partition
	lastPartitionEndStr := disk.Partitions[len(disk.Partitions)-1].End
	lastPartitionEnd, err := strconv.Atoi(lastPartitionEndStr)
	if err != nil {
		panic("Failed to retireve end position of last partition")
	}
	diskEnd, err := strconv.Atoi(disk.Size)
	if err != nil {
		panic("Failed to retireve disk end")
	}
	if lastPartitionEnd < diskEnd {
		sectors = append(sectors, Sector{lastPartitionEnd, diskEnd})
	}

	return sectors
}

type LocateDiskOutput struct {
	Disk Disk
}

func LocateDisk(diskname string) (*Disk, error) {
	findPartitionCmd := "parted -sj %s unit B print | sed -r 's/^(\\s*)\"(.)/\\1\"\\U\\2/g' | sed -r 's/(\\S)-(\\S)/\\1\\U\\2/g'"
	cmd := exec.Command("sh", "-c", fmt.Sprintf(findPartitionCmd, diskname))
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("Failed to list disk: %s", err)
	}

	var decoded LocateDiskOutput
	dec := json.NewDecoder(strings.NewReader(string(output)))
	err = dec.Decode(&decoded)
	if err != nil {
		return nil, fmt.Errorf("Failed to retrieve partition: %s", err)
	}

	device := decoded.Disk

	return &device, nil
}

func (disk *Disk) LabelDisk(label string) error {
	labelDiskCmd := "parted -s %s mklabel %s"

	err := RunCommand(fmt.Sprintf(labelDiskCmd, disk.Path, label))
	if err != nil {
		return fmt.Errorf("Failed to label disk: %s", err)
	}

	return nil
}

func (target *Disk) NewPartition(name, fsType string, start, end int) (*Partition, error) {
	createPartCmd := "parted -s %s mkpart%s \"%s\" %s %d %d"

	var partType string
	if target.Label == "msdos" {
		partType = " primary"
	} else {
		partType = ""
	}

	err := RunCommand(fmt.Sprintf(createPartCmd, target.Path, partType, name, fsType, start, end))
	if err != nil {
		return nil, fmt.Errorf("Failed to create partition: %s", err)
	}

	// TODO: Add Path to Partitions (or a pointer to parent)

	return &target.Partitions[len(target.Partitions)-1], nil
}
