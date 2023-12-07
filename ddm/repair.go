package ddm

import (
	"fmt"
	"time"

	"github.com/alfonzso/dying-disk-manager/pkg/config"
	"github.com/alfonzso/dying-disk-manager/pkg/linux"
	"github.com/alfonzso/dying-disk-manager/pkg/observer"
	log "github.com/sirupsen/logrus"
)

func (ddmData *DDMData) setupRepairThread(disk config.Disk) {
	diskStat := ddmData.GetDiskStat(disk)

	if !diskStat.Repair.IsRunning() {
		return
	}
	diskStat.Repair.HealthCheck = observer.OK

	WaitForThreadToBeIddle(fmt.Sprintf("%s - repairSetup", disk.Name), diskStat.Repair.ActionsToStop)

	if ddmData.PreRepair(disk).IsSucceed() {
		res := ddmData.Repair(disk)
		diskStat.Active = true
		if res.IsFailed() {
			diskStat.Active = false
			log.Debugf("[%s] Current disk set Active to false", disk.Name)
		}
	}
	StartThreads(diskStat.Repair.ActionsToStop)

	diskStat.Repair.SetToIddle()
	go func() {
		time.Sleep(10 * time.Second)
		diskStat := ddmData.GetDiskStat(disk)
		diskStat.Repair.HealthCheck = observer.None
	}()
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
