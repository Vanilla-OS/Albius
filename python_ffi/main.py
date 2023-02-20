from ctypes import create_string_buffer
from ffi import __lib__, exports

exports.setup_exports()

diskname = create_string_buffer(b"/dev/sda")
diskinfo = __lib__.LocateDisk(diskname)
print(diskinfo.contents)

for i, part in enumerate(diskinfo.contents.partition_iter()):
    print(f"Mounting {part.name}")
    __lib__.Mount(part, bytes(f"/mnt/test_{i}", "utf-8"))
