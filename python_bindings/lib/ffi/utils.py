from ctypes import cdll
import os


__lib__ = None

def load_library():
    global __lib__

    if __lib__:
        return __lib__

    if os.getenv("ALBIUS_SO_PATH"):
        __lib__ = cdll.LoadLibrary(os.getenv("ALBIUS_SO_PATH"))
    else:
        __lib__ = cdll.LoadLibrary(os.path.abspath("albius.so"))

    return __lib__

def __array_repr__(arr):
    fmtstr = "[ "
    for elem in arr:
        fmtstr += elem.__repr__()
        fmtstr += ", "

    return fmtstr[:-2] + " ]"
