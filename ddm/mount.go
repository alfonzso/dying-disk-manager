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
	// for {
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
	// time.Sleep(30 * time.Second)
	// }
}

func periodCheck(disk config.Disk, diskStat *observer.DiskStat) (int, error) {
	if diskStat.Active {
		log.Debugf("[%s] Mounting test", diskStat.Name)
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

	switch swch := linux.MountCommand(disk); swch {
	case linux.CommandError:
		diskStat.InactiveReason = append(diskStat.InactiveReason, "Command error happened")
	case linux.NotMounted:
		diskStat.InactiveReason = append(diskStat.InactiveReason, "Disk not or cannot mounted")
	case linux.MountedButWrongPlace:
		diskStat.InactiveReason = append(diskStat.InactiveReason, "Disk already mounted somewhere else")
	default:
		log.Debugf("Mount status: %s", swch)
		log.Debug(diskStat)
	}
	if len(diskStat.InactiveReason) > 0 {
		diskStat.Active = false
		log.Debugf("Inactive reason: %s", diskStat.InactiveReason)
	}
	// log.Debugf("%+v",diskStat)
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
