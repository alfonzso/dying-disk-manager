package ddm

import (
	"fmt"

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
				// diskStat,
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

func (ddmData *DDMData) periodCheck(disk config.Disk) (int, error) {
	statForSelectedDisk := ddmData.GetDiskStat(disk)
	actions := []*observer.Action{&statForSelectedDisk.Test}
	statForSelectedDisk.Mount.Status = observer.Running

	if IsInActiveOrDisabled("Mount", statForSelectedDisk, statForSelectedDisk.Mount) {
		statForSelectedDisk.Mount.Status = observer.Iddle
		return 0, nil
	}

	if !ddmData.Exec.CheckMountStatus(statForSelectedDisk.UUID, disk.Mount.Path).IsMountOrCommandError() {
		statForSelectedDisk.Mount.Status = observer.Iddle
		return 0, nil
	}

	WaitForThreadToBeIddle(fmt.Sprintf("%s - periodCheck", disk.Name), actions)

	if ddmData.ForceRemount(disk, statForSelectedDisk).IsFailed() {
		log.Errorf("[%s] ReMount failed", statForSelectedDisk.Name)
		statForSelectedDisk.Active = false
		statForSelectedDisk.Repair.ThreadIsRunning = true
	} else {
		log.Debugf("[%s] ReMount success", statForSelectedDisk.Name)
	}

	StartThreads(actions)
	statForSelectedDisk.Mount.Status = observer.Iddle
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
