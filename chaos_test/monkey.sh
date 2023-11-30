#!/bin/bash
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"

function ddm_kill_hdd_nicely(){
  echo "$(date +"%Y-%m-%d--%H-%M-%S") -> ddm_kill_hdd_nicely"
  item_count=$((0 + $RANDOM % 3))
  ITER=0
  for device in $(echo "4f94179d-2ddf-404b-8475-f0a643bd1639 de67ce89-8a24-4068-a7c0-8c5d67eb1fac e65c47ed-d20f-4e83-9177-7a0330281252"); do
      if [[ "$(echo "$ITER <= $item_count" | bc)" -eq "1" ]]; then
        sudo debugfs -f $DIR/hdd_failure_cmd /dev/disk/by-uuid/$device -w || true
        # echo "weeeee"
      fi
      ((ITER++))
  done
}

function ddm_umount(){
  echo "$(date +"%Y-%m-%d--%H-%M-%S") -> ddm_umount"
  item_count=$((0 + $RANDOM % 3))
  ITER=0
  for device in $(echo "4f94179d-2ddf-404b-8475-f0a643bd1639 de67ce89-8a24-4068-a7c0-8c5d67eb1fac e65c47ed-d20f-4e83-9177-7a0330281252"); do
      if [[ "$(echo "$ITER <= $item_count" | bc)" -eq "1" ]]; then
        sudo umount /dev/disk/by-uuid/$device -l || true
        echo "woooo"
      fi
      ((ITER++))
  done
}
# ls -la /mnt/disks00*
while :
do
  roll=$(echo "$RANDOM <= (32767 * 0.25)" | bc)
  [[ "$roll" -eq "1" ]] && { ddm_kill_hdd_nicely && killed=true ; }

  roll=$(echo "$RANDOM <= (32767 * 0.25)" | bc)
  [[ "$roll" -eq "1" ]] && [[ -z "$killed" ]] && ddm_umount

	sleep 300 # 5 min
done
# [[ 0 -eq 1 ]] && echo OK || echo NOK