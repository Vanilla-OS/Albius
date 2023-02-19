from ctypes import POINTER, Structure, c_char_p, c_int, c_size_t

from .utils import __array_repr__


class Partition(Structure):
    _fields_ = [
        ("name", c_char_p),
        ("majmin", c_char_p),
        ("rm", c_int),
        ("fssize", c_char_p),
        ("fstype", c_char_p),
        ("ro", c_int),
        ("mountpoints", POINTER(c_char_p)),
        ("mountpoints_size", c_size_t),
    ]

    def __repr__(self) -> str:
        fmtstr = f"\
Partition: {{ \
Name: {self.name}, \
Maj:Min: {self.majmin}, \
Size: {self.fssize}, \
Type: {self.fstype}, \
RM: {'True' if self.rm == 1 else 'False'}, \
RO: {'True' if self.ro == 1 else 'False'}, \
Mountpoints: {__array_repr__(self.mountpoints, self.mountpoints_size)} }}"
        return fmtstr


class Disk(Structure):
    _fields_ = [
        ("name", c_char_p),
        ("majmin", c_char_p),
        ("fssize", c_char_p),
        ("pttype", c_char_p),
        ("rm", c_int),
        ("ro", c_int),
        ("mountpoints", POINTER(c_char_p)),
        ("mountpoints_size", c_size_t),
        ("partitions", POINTER(Partition)),
        ("partitions_size", c_size_t),
    ]

    def __repr__(self) -> str:
        fmtstr = f"\
Disk: {{ \
Name: {self.name}, \
Maj:Min: {self.majmin}, \
Size: {self.fssize}, \
Type: {self.pttype}, \
RM: {'True' if self.rm == 1 else 'False'}, \
RO: {'True' if self.ro == 1 else 'False'}, \
Mountpoints: {__array_repr__(self.mountpoints, self.mountpoints_size)}, \
Partitions: {__array_repr__(self.partitions, self.partitions_size)} }}"
        return fmtstr
