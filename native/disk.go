package native

type Blockdevice struct {
	Name, Majmin, Size, Type string
	Rm, Ro                   bool
	Mountpoints              []string
	Children                 []Partition
}

type Partition struct {
	Name, Majmin, Size, Type string
	Rm, Ro                   bool
	Mountpoints              []string
}
