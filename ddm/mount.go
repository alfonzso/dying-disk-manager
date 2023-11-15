package ddm

import (
	"github.com/alfonzso/dying-disk-manager/pkg/linux"
	log "github.com/sirupsen/logrus"
)

func (ddmData *DDMData) BeforeMount() {
	for _, disk := range ddmData.Disks {
		currentDiskStat := ddmData.GetDiskStat(disk)
		currentDiskStat.Active = true
		if !linux.CheckDiskAvailability(disk.UUID) {
			currentDiskStat.Active = false
			currentDiskStat.InactiveReason = append(currentDiskStat.InactiveReason, "Disk UUID not found")
		}
		ddmData.DiskStat = append(ddmData.DiskStat, currentDiskStat)
	}
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
		}

	}
}
