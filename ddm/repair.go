package ddm

import (
	"github.com/alfonzso/dying-disk-manager/pkg/config"
	"github.com/alfonzso/dying-disk-manager/pkg/linux"
	"github.com/alfonzso/dying-disk-manager/pkg/observer"
	log "github.com/sirupsen/logrus"
)

func (ddmData *DDMData) setupRepairThread() {
	for _, disk := range ddmData.Disks {
		diskStat := ddmData.GetDiskStat(disk)
		if diskStat.Repair.ThreadIsRunning {
			ddmData.Scheduler.RemoveByTags(diskStat.UUID)
			if ddmData.PreRepair(disk, diskStat).IsSucceed() {
				ddmData.Repair(disk)
			}
			diskStat.Repair.ThreadIsRunning = false
		}
	}
}

func (ddmData *DDMData) PreRepair(disk config.Disk, diskStat *observer.DiskStat) linux.LinuxCommands {
	log.Debugf("[%s] PreRepair ...", diskStat.Name)

	WaitForThreadToBeIddle([]observer.Action{diskStat.Mount, diskStat.Test})

	if ddmData.Exec.UMount(disk).IsFailed() {
		log.Debugf("[%s] PreRepair failed to umount disk ... ", diskStat.Name)
		return linux.CommandError
	}
	return linux.CommandSuccess
}

func (ddmData *DDMData) Repair(disk config.Disk) {
	log.Debugf("[%s] Repair ...", disk.Name)
}
