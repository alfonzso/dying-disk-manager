package ddm

import (
	"github.com/alfonzso/dying-disk-manager/pkg/config"
	"github.com/alfonzso/dying-disk-manager/pkg/linux"
	"github.com/alfonzso/dying-disk-manager/pkg/observer"
	log "github.com/sirupsen/logrus"
)

func (ddmData *DDMData) isMountCanBeRun(disk config.Disk, diskStat *observer.DiskStat) bool {
	return (disk.Mount.Enabled || ddmData.Common.Mount.Enabled) && !diskStat.Mount.ThreadIsRunning
}

func (ddmData *DDMData) setupMountThread() {
	for _, disk := range ddmData.Disks {
		diskStat := ddmData.GetDiskStat(disk)
		if diskStat.Repair.ThreadIsRunning {
			if diskStat.Mount.Status.IsRunning() {
				diskStat.Mount.Status = observer.Iddle
			}
			log.Debugf("[%s] MOUNT -> Repair is ON", disk.Name)
		} else if ddmData.isMountCanBeRun(disk, diskStat) {
			go ddmData.SetupCron(
				"MOUNT",
				ddmData.periodCheck,
				disk,
				diskStat,
				GetCronExpr(disk.Mount.PeriodicCheck.Cron, ddmData.Common.Mount.PeriodicCheck.Cron),
			)
			diskStat.Mount.ThreadIsRunning = true
		}
	}
}

func (ddmData *DDMData) ForceRemount(disk config.Disk, diskStat *observer.DiskStat) linux.LinuxCommands {
	log.Debugf("Try to umount => %s", ddmData.Exec.UMount(disk))
	if ddmData.Exec.Mount(disk).IsFailed() {
		log.Errorf("[%s] Mount failed", diskStat.Name)
		return linux.CommandError
	}
	return linux.Mounted
}

func ThreadContinue(diskStat *observer.DiskStat) bool {
	if diskStat.IsNotActiv() || diskStat.Mount.DisabledByAction {
		log.Warningf("[%s] MountThread => Disk deactivated => active: %t, disabledBy: %t",
			diskStat.Name, diskStat.Active, diskStat.Mount.DisabledByAction,
		)
		return true
	}
	return false
}

func (ddmData *DDMData) periodCheck(disk config.Disk, diskStat *observer.DiskStat) (int, error) {
	diskStat.Mount.Status = observer.Running

	if ThreadContinue(diskStat) {
		return 0, nil
	}

	if !ddmData.Exec.CheckMountStatus(diskStat.UUID, disk.Mount.Path).IsMountOrCommandError() {
		return 0, nil
	}

	WaitForThreadToBeIddle([]observer.Action{diskStat.Test})

	if ddmData.ForceRemount(disk, diskStat).IsFailed() {
		log.Errorf("[%s] ReMount failed", diskStat.Name)
		diskStat.Active = false //TODO may trigger a repair ...
		return 0, nil
	}

	log.Debugf("[%s] ReMount success", diskStat.Name)
	diskStat.Mount.Status = observer.Iddle
	return 0, nil
}

func (ddmData *DDMData) BeforeMount() {
	for _, disk := range ddmData.Disks {
		diskStat := ddmData.GetDiskStat(disk)
		diskStat.Active = true
		if status := ddmData.Exec.CheckDiskAvailability(disk.UUID); status.IsDiskUnAvailableOrUUIDNotExists() {
			diskStat.Active = false
			diskStat.InactiveReason = append(diskStat.InactiveReason, linux.DetailedLinuxType[status])
		}
	}
}

func (ddmData *DDMData) linuxMount(disk config.Disk, diskStat *observer.DiskStat) {
	status := ddmData.Exec.MountCommand(disk)

	if status.IsMountOrCommandError() {
		diskStat.Active = false
		diskStat.InactiveReason = append(diskStat.InactiveReason, linux.DetailedLinuxType[status])
		log.Debugf("Inactive reason: %s", diskStat.InactiveReason)
	}

	log.Debugf("Mount status: %s", status)
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
			ddmData.linuxMount(disk, diskStat)
		} else {
			log.Debugf("Mount disabled: %s", disk.Name)
		}

	}
}
