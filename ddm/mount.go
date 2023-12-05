package ddm

import (
	"fmt"
	"time"

	"github.com/alfonzso/dying-disk-manager/pkg/config"
	"github.com/alfonzso/dying-disk-manager/pkg/linux"
	"github.com/alfonzso/dying-disk-manager/pkg/observer"
	log "github.com/sirupsen/logrus"
)

func (ddmData *DDMData) setupMountThread(disk config.Disk) {
	diskStat := ddmData.GetDiskStat(disk)

	if RepairIsOn(diskStat.Mount.Name, diskStat) {
		return
	}

	if !(disk.Mount.Enabled || ddmData.Common.Mount.Enabled) {
		return
	}

	go ddmData.startMountAndWaitTillAlive(disk)

}

func (ddmData *DDMData) startMountAndWaitTillAlive(disk config.Disk) {
	diskStat := ddmData.GetDiskStat(disk)

	if ddmData.ActionsJobRunning(diskStat.Mount.Name, disk.UUID) {
		return
	}

	ddmData.SetupCron(
		diskStat.Mount.Name,
		ddmData.periodCheck,
		disk,
		GetCronExpr(disk.Mount.PeriodicCheck.Cron, ddmData.Common.Mount.PeriodicCheck.Cron),
	)

	diskStat.Mount.SetToRun()
}

func (ddmData *DDMData) ForceRemount(disk config.Disk, diskStat *observer.DiskStat) linux.LinuxCommands {
	if ddmData.Exec.UMount(disk) != linux.UMounted {
		return linux.CommandError
	}
	if ddmData.Exec.Mount(disk).IsFailed() {
		return linux.CantMounted
	}
	return linux.Mounted
}

func (ddmData *DDMData) periodCheck(disk config.Disk) (int, error) {
	res, err := ddmData.mountWrapper(disk)
	go func() {
		diskStat := ddmData.GetDiskStat(disk)
		times := ddmData.GetJobNextRun(diskStat.Mount.Name, diskStat.UUID)
		time.Sleep(times)
		diskStat.Mount.HealthCheck = observer.Iddle
	}()
	return res, err
}

func (ddmData *DDMData) mountWrapper(disk config.Disk) (int, error) {
	diskStat := ddmData.GetDiskStat(disk)
	diskStat.Mount.SetToRun()
	diskStat.Mount.HealthCheck = observer.Running

	if IsInActiveOrDisabled(diskStat.Mount.Name, diskStat, diskStat.Mount) {
		diskStat.Mount.SetToIddle()
		return 0, nil
	}

	if !ddmData.Exec.CheckMountStatus(diskStat.UUID, disk.Mount.Path).IsMountOrCommandError() {
		diskStat.Mount.SetToIddle()
		return 0, nil
	}

	WaitForThreadToBeIddle(fmt.Sprintf("%s - periodCheck", disk.Name), diskStat.Mount.ActionsToStop)

	if error := ddmData.ForceRemount(disk, diskStat); error.IsForceRemountError() {
		log.Errorf("[%s] ReMount failed: %s", diskStat.Name, error)
		diskStat.Active = false
		diskStat.Repair.SetToRun()
	} else {
		log.Debugf("[%s] ReMount success", diskStat.Name)
	}

	StartThreads(diskStat.Mount.ActionsToStop)
	diskStat.Mount.SetToIddle()
	diskStat.Mount.DisabledByAction = false
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
