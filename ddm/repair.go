package ddm

import (
	"time"

	"github.com/alfonzso/dying-disk-manager/pkg/config"
	"github.com/alfonzso/dying-disk-manager/pkg/linux"
	"github.com/alfonzso/dying-disk-manager/pkg/observer"
	log "github.com/sirupsen/logrus"
)

func (ddmData *DDMData) setupRepairThread() {
	for _, disk := range ddmData.Disks {
		diskStat := ddmData.GetDiskStat(disk)
		if diskStat.RepairThreadIsRunning {
			ddmData.Scheduler.RemoveByTags(diskStat.UUID)
			if ddmData.PreRepair(disk, diskStat).IsSucceed() {
				ddmData.Repair(disk)
			}
			diskStat.RepairThreadIsRunning = false
		}
	}
}

func (ddmData *DDMData) PreRepair(disk config.Disk, diskStat *observer.DiskStat) linux.Linux {
	log.Debugf("[%s] PreRepair ...", diskStat.Name)
	for {
		if diskStat.IsMountAndTestActionInIddleStatus() {
			break
		}
		log.Debugf("[%s] PreRepair wait actions to be done ", diskStat.Name)
		time.Sleep(10 * time.Second)
	}

	if linux.UMount(disk).IsFailed() {
		log.Debugf("[%s] PreRepair failed to umount disk ... ", diskStat.Name)
		return linux.CommandError
	}
	return linux.CommandSuccess
}

func (ddmData *DDMData) Repair(disk config.Disk) {
	log.Debugf("[%s] Repair ...", disk.Name)
}
