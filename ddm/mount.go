package ddm

import (
	"time"

	"github.com/alfonzso/dying-disk-manager/pkg/config"
	"github.com/alfonzso/dying-disk-manager/pkg/linux"
	"github.com/alfonzso/dying-disk-manager/pkg/observer"
	log "github.com/sirupsen/logrus"
)

func (ddmData *DDMData) ThreadMountPeriodCheck() {
	for {
		for _, disk := range ddmData.Disks {
			diskMount := disk.Mount
			commonMount := ddmData.Common.Mount
			diskStat := ddmData.GetDiskStat(disk)
			if (diskMount.Enabled || commonMount.Enabled) && diskStat.Active {
				// cron := getCronExpr(diskMount.PeriodicCheck.Cron, commonMount.PeriodicCheck.Cron)
				// if ddmData.IsCronDue(cron) {
				// 	ddmData.periodCheck()
				// }
			}
		}
		time.Sleep(30 * time.Second)
	}

}

func (ddmData *DDMData) periodCheck() {

}

func (ddmData *DDMData) BeforeMount() {
	for _, disk := range ddmData.Disks {
		currentDiskStat := ddmData.GetDiskStat(disk)
		currentDiskStat.Active = true
		if !linux.CheckDiskAvailability(disk.UUID) {
			currentDiskStat.Active = false
			currentDiskStat.InactiveReason = append(currentDiskStat.InactiveReason, "Disk UUID not found")
		}
		ddmData.DiskStat = append(ddmData.DiskStat, *currentDiskStat)
	}
}

func linuxMount(disk config.Disk, diskStat *observer.DiskStat) {
	switch linux.MountCommand(disk) {
	case linux.CommandError:
		diskStat.InactiveReason = append(diskStat.InactiveReason, "Command error happened")
	case linux.NotMounted:
		diskStat.InactiveReason = append(diskStat.InactiveReason, "Disk not or cannot mounted")
	case linux.MountedButWrongPlace:
		diskStat.InactiveReason = append(diskStat.InactiveReason, "Disk already mounted somewhere else")
	}
	if len(diskStat.InactiveReason) > 0 {
		diskStat.Active = false
		log.Debugf("Inactive reason: %s", diskStat.InactiveReason)
	}
	log.Debugf("%+v",diskStat)
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
