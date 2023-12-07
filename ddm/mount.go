package ddm

import (
	"fmt"
	"time"

	"github.com/alfonzso/dying-disk-manager/pkg/config"
	"github.com/alfonzso/dying-disk-manager/pkg/linux"
	"github.com/alfonzso/dying-disk-manager/pkg/observer"
	log "github.com/sirupsen/logrus"
)

func (ddmData *DDMData) setupMountThread(diskData *DiskData) {

	if RepairIsOn(diskData.Mount.Name, diskData) {
		return
	}

	if !(diskData.conf.Mount.Enabled || ddmData.Common.Mount.Enabled) {
		return
	}

	go ddmData.startMountAndWaitTillAlive(diskData)
}

func (ddmData *DDMData) startMountAndWaitTillAlive(diskData *DiskData) {

	if ddmData.ActionsJobRunning(diskData.Mount.Name, diskData.UUID, diskData.Mount.Cron) {
		return
	}

	ddmData.SetupCron(
		diskData.Mount.Name,
		ddmData.periodCheck,
		diskData,
		diskData.Mount.Cron,
	)

	diskData.Mount.SetToRun()
}

func (ddmData *DDMData) periodCheck(diskData *DiskData) (int, error) {
	res, err := ddmData.mountPeriodWrapper(diskData)
	go func() {
		times := ddmData.GetJobNextRun(diskData.Mount.Name, diskData.UUID)
		time.Sleep(times)
		diskData.Mount.HealthCheck = observer.None
	}()
	return res, err
}

func (ddmData *DDMData) mountPeriodWrapper(diskData *DiskData) (int, error) {
	diskData.Mount.SetToRun()
	diskData.Mount.HealthCheck = observer.OK

	if IsInActiveOrDisabled(diskData.Mount.Name, diskData, diskData.Mount) {
		diskData.Mount.SetToIddle()
		return 0, nil
	}

	// if !ddmData.Exec.CheckMountStatus(diskData.UUID, diskData.conf.Mount.Path).IsMountOrCommandError() {
	if ddmData.Exec.CheckMountStatus(diskData.UUID, diskData.conf.Mount.Path).IsMounted() {
		diskData.Mount.SetToIddle()
		return 0, nil
	}

	WaitForThreadToBeIddle(fmt.Sprintf("%s - periodCheck", diskData.Name), diskData.Mount.ActionsToStop)

	if error := ddmData.ForceRemount(diskData); error.IsForceRemountError() {
		log.Errorf("[%s] ReMount failed: %s", diskData.Name, error)
		diskData.Active = false
		diskData.Repair.SetToRun()
	} else {
		log.Debugf("[%s] ReMount success", diskData.Name)
	}

	StartThreads(diskData.Mount.ActionsToStop)
	diskData.Mount.SetToIddle()
	diskData.Mount.DisabledByAction = false
	return 0, nil
}

func (ddmData *DDMData) ForceRemount(diskData *DiskData) linux.LinuxCommands {
	if ddmData.Exec.UMount(*diskData.conf) != linux.UMounted {
		return linux.CommandError
	}
	if ddmData.Exec.Mount(*diskData.conf).IsFailed() {
		return linux.CantMounted
	}
	return linux.Mounted
}

///////////////////////////////////////
// Init task
///////////////////////////////////////

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
