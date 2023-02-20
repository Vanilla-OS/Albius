from ctypes import POINTER, c_char_p

from ffi import __lib__

from .disk import Disk


def setup_exports():
    __lib__.LocateDisk.argtypes = [c_char_p]
    __lib__.LocateDisk.restype = POINTER(Disk)
