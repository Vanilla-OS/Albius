from ctypes import POINTER, Structure, c_char_p, c_int

from .utils import __array_repr__


class Partition(Structure):
    _fields_ = [
        ("_path", c_char_p),
        ("_number", c_int),
        ("_start", c_char_p),
        ("_end", c_char_p),
        ("_size", c_char_p),
        ("_type", c_char_p),
        ("_filesystem", c_char_p),
    ]

    def __repr__(self) -> str:
        fmtstr = f"\
Partition: {{ \
Path: {self.path}, \
Number: {self.number}, \
Start: {self.start}, \
End: {self.end}, \
Size: {self.size}, \
Type: {self.type}, \
Filesystem: {self.filesystem} }}"
        return fmtstr

    @property
    def path(self) -> str:
        return self._path.decode()

    @property
    def number(self) -> int:
        return self._number

    @property
    def start(self) -> str:
        return self._start.decode()

    @property
    def end(self) -> str:
        return self._end.decode()

    @property
    def size(self) -> str:
        return self._size.decode()

    @property
    def type(self) -> str:
        return self._type.decode()

    @property
    def filesystem(self) -> str:
        return self._filesystem.decode()


class Disk(Structure):
    _fields_ = [
        ("_path", c_char_p),
        ("_size", c_char_p),
        ("_model", c_char_p),
        ("_transport", c_char_p),
        ("_logical_sector_size", c_int),
        ("_physical_sector_size", c_int),
        ("_label", c_char_p),
        ("_max_partitions", c_int),
        ("_partitions", POINTER(Partition)),
        ("_partitions_count", c_int),
    ]

    def __repr__(self) -> str:
        fmtstr = f"\
Disk: {{ \
Path: {self.path}, \
Size: {self.size}, \
Model: {self.model}, \
Transport: {self.transport}, \
Logical_sector_size: {self.logical_sector_size}, \
Physical_sector_size: {self.physical_sector_size}, \
Label: {self.label}, \
Max_partitions: {self.max_partitions}, \
Partitions: {__array_repr__(self.partitions)} }}"
        return fmtstr

    @property
    def path(self) -> str:
        return self._path.decode()

    @property
    def size(self) -> str:
        return self._size.decode()

    @property
    def model(self) -> str:
        return self._model.decode()

    @property
    def transport(self) -> str:
        return self._transport.decode()

    @property
    def logical_sector_size(self) -> int:
        return self._logical_sector_size

    @property
    def physical_sector_size(self) -> int:
        return self._physical_sector_size

    @property
    def label(self) -> str:
        return self._label.decode()

    @property
    def max_partitions(self) -> int:
        return self._max_partitions

    @property
    def partitions(self):
        for i in range(self.partitions_count):
            yield self._partitions[i]

    @property
    def partitions_count(self) -> int:
        return self._partitions_count
