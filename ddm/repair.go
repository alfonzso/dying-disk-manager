package ddm

import (
	"fmt"

	"github.com/alfonzso/dying-disk-manager/pkg/config"
	"github.com/alfonzso/dying-disk-manager/pkg/linux"
	"github.com/alfonzso/dying-disk-manager/pkg/observer"
	log "github.com/sirupsen/logrus"
)

func (ddmData *DDMData) setupRepairThread(disk config.Disk) {
	statForSelectedDisk := ddmData.GetDiskStat(disk)
	actions := []*observer.Action{&statForSelectedDisk.Mount, &statForSelectedDisk.Test}

	if !statForSelectedDisk.Repair.IsRunning() {
		return
	}
	ddmData.Scheduler.RemoveByTags(statForSelectedDisk.UUID)
	statForSelectedDisk.Mount.ThreadIsRunning = false
	statForSelectedDisk.Test.ThreadIsRunning = false

	WaitForThreadToBeIddle(fmt.Sprintf("%s - repairSetup", disk.Name), actions)

	if ddmData.PreRepair(disk).IsSucceed() {
		res := ddmData.Repair(disk)
		statForSelectedDisk.Active = true
		if res.IsFailed() {
			statForSelectedDisk.Active = false
			log.Debugf("[%s] Current disk set Active to false", disk.Name)
		}
	}
	StartThreads(actions)

	statForSelectedDisk.Repair.SetToIddle()
}

func (ddmData *DDMData) PreRepair(disk config.Disk) linux.LinuxCommands {
	statForSelectedDisk := ddmData.GetDiskStat(disk)
	log.Debugf("[%s] PreRepair ...", statForSelectedDisk.Name)

	if ddmData.Exec.UMount(disk).IsFailed() {
		log.Debugf("[%s] PreRepair failed to umount disk ... ", statForSelectedDisk.Name)
		return linux.CommandError
	}
	return linux.CommandSuccess
}

func (ddmData *DDMData) Repair(disk config.Disk) linux.LinuxCommands {
	if ddmData.Exec.RunFsck(disk.UUID).IsFailed() {
		log.Debugf("[%s] Repair with fsck failed :( ", disk.Name)
		return linux.CommandError
	}

	if ddmData.Exec.Mount(disk).IsFailed() {
		log.Debugf("[%s] Mount after repair failed :( ", disk.Name)
		return linux.CommandError
	}

	log.Debugf("[%s] Repair Succeeded ! ", disk.Name)
	return linux.CommandSuccess
}
