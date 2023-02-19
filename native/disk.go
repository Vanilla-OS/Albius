package native

type Blockdevice struct {
	Name, Majmin, Fssize, Pttype string
	Rm, Ro                       bool
	Mountpoints                  []string
	Children                     []Partition
}

type Partition struct {
	Name, Majmin, Fssize, Fstype string
	Rm, Ro                       bool
	Mountpoints                  []string
}
