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
		if diskStat.RepairThreadIsRunning {
			if diskStat.ActionStatus.Mount.IsRunning() {
				diskStat.ActionStatus.Mount = observer.Iddle
			}
			log.Debugf("[%s] MOUNT -> Repair is ON", disk.Name)
		} else if ddmData.isMountCanBeRun(disk, diskStat) {
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
	if linux.Mount(disk).IsFailed() {
		log.Errorf("[%s] Mount failed", diskStat.Name)
		return linux.CommandError
	}
	return linux.Mounted
}

func periodCheck(disk config.Disk, diskStat *observer.DiskStat) (int, error) {
	diskStat.ActionStatus.Mount = observer.Running
	if !diskStat.Active {
		log.Warningf("[%s] Disk not active in PeriodCheck thread", disk.Name)
		return 0, nil
	}
	if !linux.CheckMountStatus(diskStat.UUID, disk.Mount.Path).IsMountOrCommandError() {
		return 0, nil
	}
	if ForceRemount(disk, diskStat) == linux.CommandError {
		log.Errorf("[%s] ReMount failed", diskStat.Name)
		diskStat.Active = false //TODO may trigger a repair ...
		return 0, nil
	}
	log.Debugf("[%s] ReMount success", diskStat.Name)
	diskStat.ActionStatus.Mount = observer.Iddle
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
	if mountResult.IsMountOrCommandError() {
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
