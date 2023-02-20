from ctypes import create_string_buffer
from ffi import __lib__, exports

exports.setup_exports()

diskname = create_string_buffer(b"/dev/sda")
diskinfo = __lib__.LocateDisk(diskname)
print(diskinfo.contents)
