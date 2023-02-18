from ctypes import cdll
import os

__lib__ = cdll.LoadLibrary(os.path.abspath("albius.so"))
