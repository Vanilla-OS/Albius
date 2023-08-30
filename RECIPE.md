# Recipe commands

## Setup 

### label

Create a new partition table on the disk.

**Accepts**:
- *LabelType* (`string`): The partitioning scheme. Either `mbr` or `gpt`.

### mkpart

Create a new partition on the disk.

**Accepts**:
- *Name* (`string`): The name for the partition.
- *FsType* (`string`): The filesystem for the partition. Can be either `btrfs`,
`ext[2,3,4]`, `linux-swap`, `ntfs`\*, `reiserfs`\*, `udf`\*, or `xfs`\*. If FsType
is prefixed with `luks-` (e.g. `luks-btrfs`), the partition will be encrypted using LUKS2.
- *Start* (`int`): The start position on disk for the new partition (in MiB).
- *End* (`int`): The end position on disk for the new partition (in MiB), or -1 for
using all the remaining space.
- *LUKSPassword* (optional `string`): The password used to encrypt the partition. Only
relevant if `FsType` is prefixed with `luks-`.

\* = Not fully tested. Please create an issue if you encouter problems.

### rm

Delete a partition from the disk.

**Accepts**:
- *PartNum* (`int`): The partition number on disk (e.g. `/dev/sda3` is partition 3).

### resizepart

Resize a partition on disk.

**Accepts**:
- *PartNum* (`int`): The partition number on disk (e.g. `/dev/sda3` is partition 3).
- *PartNewSize* (`int`): The new size in MiB for the partition.

### namepart

Rename the specified partition.

**Accepts**:
- *PartNum* (`int`): The partition number on disk (e.g. `/dev/sda3` is partition 3).
- *PartNewName* (`string`): The new name for the partition.

### setflag

Set the value of a partition flag, from the flags supported by parted.
See parted(8) for the full list.

**Accepts**:
- *PartNum* (`int`): The partition number on disk (e.g. `/dev/sda3` is partition 3).
- *FlagName* (`string`): The name of the flag.
- *State* (`bool`): The value to apply to the flag. Either `true` or `false`.

### format

Format an existing partition to a specified filesystem. This operation will destroy all data.

**Accepts**:
- *PartNum* (`int`): The partition number on disk (e.g. `/dev/sda3` is partition 3).
- *FsType* (`string`): The filesystem for the partition. Can be either `btrfs`, `ext[2,3,4]`, `linux-swap`, `ntfs`\*, `reiserfs`\*, `udf`\*, or `xfs`\*.

### luks-format

Same as `format` but encrypts the partition with LUKS2.

**Accepts**:
- *PartNum* (`int`): The partition number on disk (e.g. `/dev/sda3` is partition 3).
- *FsType* (`string`): The filesystem for the partition. Can be either `btrfs`, `ext[2,3,4]`, `linux-swap`, `ntfs`\*, `reiserfs`\*, `udf`\*, or `xfs`\*.
- *Password* (`string`): The password used to encrypt the partition.

--- 

## Post-Installation 

### adduser

Create a new user.

**Accepts**:
- *Username* (`string`): The username of the new user.
- *Fullname* (`string`): The full name (display name) of the new user.
- *Groups* (`[string]`): A list of groups the new user belongs to (the new user is automatically part of its own group).
- *Password* (optional `string`): The password for the user. If not provided, password login will be disabled.

### timezone

Set the timezone.

**Accepts**:
- *TZ* (`string`): The timezone code (e.g. `America/Sao_Paulo`).

### shell

Run a shell command. This command accepts a variable number or parameters, where each parameter is a separate command to run.

**Accepts**:
- *Command(s)* (`...string`): The shell command(s) to execute.

### pkgremove

Given a file containing a list of packages, use the specified package manager to remove them.

**Accepts**:
- *PkgRemovePath* (`string`): The path containing the list of packages to remove.
- *RemoveCmd* (`string`): The package manager command to remove packages (e.g. `apt remove`).

### hostname

Set the system's hostname.

**Accepts**:
- *NewHostname* (`string`): The hostname to set.

### locale

Set the system's locale, using `locale-gen` to generate the locale if not present.

**Accepts**:
- *LocaleCode* (`string`): The locale code to use. See `/etc/locale.gen` for the full list of locale codes.

### swapon

Use the provided partition as swap space.

**Accepts**:
- *Partition* (`string`): The partition to use as swap.

### keyboard

Set the system keyboard layout. See `keyboard(5)` for more details.

**Accepts**:
- *Layout* (`string`): The keyboard's layout (XKBLAYOUT).
- *Model* (`string`): The keyboard's model (XKBMODEL).
- *Variant* (`string`): The keyboard's variant (XKBVARIANT).

### grub-install

Install GRUB to the specified partition.

**Accepts**:
- *BootDirectory* (`string`): The path for the boot dir (usually `/boot`).
- *InstallDevice* (`string`): The disk where the boot partition is located.
- *Target* (`string`): The target firmware. Either `bios` for legacy systems or `efi` for UEFI systems.

### grub-default-config

Write key-value pairs into `/etc/default/grub`. This command accepts a variable number of parameters, where each parameter represents a new item to add to the file.

**Accepts**:
- *KV(s)* (`...string`): The `KEY=value` pair(s) to add to the GRUB default file.

### grub-add-script

Add one or more script files into `/etc/default/grub.d`. This command accepts a variable number of parameters, where each parameter represents a new file to add to the directory.

**Accepts**:
- *ScriptPath(s)* (`...string`): The file path(s) for each script to add.

### grub-remove-script

Remove one or more script files from `/etc/default/grub.d`. This command accepts a variable number of parameters, where each parameter represents a file to delete from the directory.

**Accepts**:
- *ScriptPath(s)* (`...string`): The file path(s) for each script to be removed.

### grub-mkconfig

Run the `grub-mkconfig` command to generate a new GRUB configuration into the specified output path.

**Accepts**:
- *OutputPath* (`string`): The target path for the generated config.

