package native

import "strconv"

type Sector struct {
	Start, End int
}

type Disk struct {
	Path, Size, Model, Transport, Label                                   string
	LogicalSectorSize, PhysicalSectorSize, MaxPartitions, PartitionsCount int
	Partitions                                                            []Partition
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

type Partition struct {
	Number                             int
	Start, End, Size, Type, Filesystem string
}
