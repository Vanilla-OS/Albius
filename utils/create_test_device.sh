#!/bin/bash

declare -a partitions

while :; do
    case "$1" in
        -o)
            output=$2
            shift
            shift
            ;;
        -s)
            size=$2
            shift
            shift
            ;;
        -p)
            partitions+=("$2;$3;$4;$5")
            shift
            shift
            shift
            shift
            shift
            ;;
        *)
            break
            ;;
    esac
done

dd if=/dev/zero of="$output" bs=1024 count="$size" > /dev/null 2>&1
device=$(losetup --find --show "$output")

parted -s "$device" mklabel gpt

for partition in "${partitions[@]}"; do
    IFS=";" read -r -a args <<< "${partition}"
    #                                    name          fs         start          end
    parted -s "$device" unit MiB mkpart "${args[0]}" "${args[1]}" "${args[2]}" "${args[3]}"
done

echo "$device"
