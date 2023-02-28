from ctypes import create_string_buffer
from ffi import __lib__, exports

exports.setup_exports()

diskname = create_string_buffer(b"/dev/sda")
diskinfo = __lib__.LocateDisk(diskname)
print(diskinfo.contents)

# for i, part in enumerate(diskinfo.contents.partitions):
#     print(f"Mounting {part.path}")
#     __lib__.Mount(part, bytes(f"/mnt/test_{i}", "utf-8"))
#     print(f"Unmounting {part.path} via partition")
#     __lib__.UmountPartition(part)

#     print(f"Mounting {part.path}")
#     __lib__.Mount(part, bytes(f"/mnt/test_{i}", "utf-8"))
#     print(f"Unmounting {part.path} via directory")
#     __lib__.UmountDirectory(bytes(f"/mnt/test_{i}", "utf-8"))
