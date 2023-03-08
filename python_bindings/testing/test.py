import unittest
from lib import albius
import os

# Tests
# -------------------------
# - Create partition
# - Label disk
# - Mount
# - Unmount (disk and directory)
# - New partition
# - Name partition
# - Set partition flag
# - Resize partition
# - Remove partition

class TestDiskOps(unittest.TestCase):
    self.__disk = None

    def test_locate_disk(self):
        self.__disk = albius.LocateDisk(b"/dev/nvme0n1")
        self.assertIsNotNone(self.__disk)

    # @unittest.skipUnless(os.getenv("UNSAFE_TESTS"))
    # def test_mount_partition(self):
    #     os.mkdir("/mnt/albius_mount_test")

# class TestPartitioning(unittest.TestCase):
#     pass

# class TestFileOps(unittest.TestCase):
#     pass
