package native

type Disk struct {
	Path, Size, Model, Transport, Label                                   string
	LogicalSectorSize, PhysicalSectorSize, MaxPartitions, PartitionsCount int
	Partitions                                                            []Partition
}

type Partition struct {
	Number                             int
	Start, End, Size, Type, Filesystem string
}
