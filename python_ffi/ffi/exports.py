from ctypes import POINTER, c_char_p

from ffi import __lib__

from .disk import Disk, Partition


def setup_exports():
    # LocateDisk
    __lib__.LocateDisk.argtypes = [c_char_p]
    __lib__.LocateDisk.restype = POINTER(Disk)

    # Mount
    __lib__.Mount.argtypes = [POINTER(Partition), c_char_p]
    __lib__.Mount.restype = None

    # UmountPartition
    __lib__.UmountPartition.argtypes = [POINTER(Partition)]
    __lib__.UmountPartition.restype = None

    # UmountDirectory
    __lib__.UmountDirectory.argtypes = [c_char_p]
    __lib__.UmountDirectory.restype = None
