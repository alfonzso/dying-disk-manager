package ddm

import (
	"github.com/alfonzso/dying-disk-manager/pkg/config"
	"github.com/alfonzso/dying-disk-manager/pkg/linux"
	"github.com/alfonzso/dying-disk-manager/pkg/observer"
	log "github.com/sirupsen/logrus"
)

func (ddmData *DDMData) isMountCanBeRun(disk config.Disk, diskStat *observer.DiskStat) bool {
	return (disk.Mount.Enabled || ddmData.Common.Mount.Enabled) && !diskStat.MountThreadIsRunning
}

func (ddmData *DDMData) setupMountThread() {
	for _, disk := range ddmData.Disks {
		diskStat := ddmData.GetDiskStat(disk)
		if ddmData.isMountCanBeRun(disk, diskStat) {
			go ddmData.SetupCron(
				"MOUNT",
				periodCheck,
				disk,
				diskStat,
				GetCronExpr(disk.Mount.PeriodicCheck.Cron, ddmData.Common.Mount.PeriodicCheck.Cron),
			)
			diskStat.MountThreadIsRunning = true
		}
	}
}

func ForceRemount(disk config.Disk, diskStat *observer.DiskStat) linux.Linux {
	log.Debugf("Try to umount => %s", linux.UMount(disk))
	if linux.Mount(disk) == linux.CommandError {
		log.Errorf("[%s] Mount failed", diskStat.Name)
		return linux.CommandError
	}
	return linux.Mounted
}

func periodCheck(disk config.Disk, diskStat *observer.DiskStat) (int, error) {
	if diskStat.Active {
		if linux.IsDiskMountHasError(diskStat.UUID, disk.Mount.Path) {
			log.Debugf("[%s] IsDiskMountHasError", diskStat.Name)
			if ForceRemount(disk, diskStat) == linux.CommandError {
				diskStat.Active = false
				return 0, nil
			}
			log.Debugf("[%s] ReMount success", diskStat.Name)
		}
	}
	return 0, nil
}

func (ddmData *DDMData) BeforeMount() {
	for _, disk := range ddmData.Disks {
		currentDiskStat := ddmData.GetDiskStat(disk)
		currentDiskStat.Active = true
		if !linux.CheckDiskAvailability(disk.UUID) {
			currentDiskStat.Active = false
			currentDiskStat.InactiveReason = append(currentDiskStat.InactiveReason, "Disk UUID not found")
		}
	}
}

func linuxMount(disk config.Disk, diskStat *observer.DiskStat) {
	mountResult := linux.MountCommand(disk)
	if err := linux.IsMountOrCommandError(mountResult); err {
		diskStat.InactiveReason = append(diskStat.InactiveReason, linux.DetailedLinuxType[mountResult])
	}

	if len(diskStat.InactiveReason) > 0 {
		diskStat.Active = false
		log.Debugf("Inactive reason: %s", diskStat.InactiveReason)
	}

	log.Debugf("Mount status: %s", mountResult)
	log.Debug(diskStat)
}

func (ddmData *DDMData) Mount() {
	for _, disk := range ddmData.Disks {
		diskStat := ddmData.GetDiskStat(disk)

		if !diskStat.Active {
			log.WithFields(log.Fields{
				"disk":   disk.Name,
				"reason": diskStat.InactiveReason,
			}).Debug("Disk skipped cuz inactive")
			continue
		}

		if disk.Mount.Enabled || ddmData.Common.Mount.Enabled {
			log.Debugf("Mounting... %s", disk.Name)
			linuxMount(disk, diskStat)
		} else {
			log.Debugf("Mount disabled: %s", disk.Name)
		}

	}
}
