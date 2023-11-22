START
 ˇˇˇ
check config
no config        -> error
parse config
incorrect config -> error
 ˘˘˘
check disk available
 no -> error
check mount path exists
 yes -> error
 ˇˇˇ
???? already mounted
! Mount it !
 ? error -> log & deactivate current disk
THREAD
* periodCheck is enabledInConfig
* periodCheck is enabled
THREAD
* test if is enabledInConfig
* test if is enabled
  TEST:
  * check disk mounted
    yes -> cont
    no  -> mount it
      succ -> cont
      fail -> triggerRepair
  * write in disk
    succ -> cont
    fail -> triggerRepair
  triggerRepair:
   repair if is enabledInConfig:
    - wait for periodCheck is done
    - wait for test is done
    - disable periodCheck Thread
    - disable test Thread
    - umount disk
    - trigger repair
THREAD
* repair
  - commandBefore
  - command
  - commandAfter
  - enable periodCheck Thread
  - enable test Thread



task -> loop disks
  disk
    -> is enabled
    -> is disk active
    task -> test

[1334152.944264] EXT4-fs (sde1): initial error at time 1699921467: ext4_find_extent:921: inode 5898266
[1334152.944275] EXT4-fs (sde1): last error at time 1700029631: ext4_find_extent:921: inode 5898271
[1426432.083661] EXT4-fs (sde1): error count since last fsck: 2382
[1426432.083675] EXT4-fs (sde1): initial error at time 1699921467: ext4_find_extent:921: inode 5898266
[1426432.083689] EXT4-fs (sde1): last error at time 1700029631: ext4_find_extent:921: inode 5898271
