package ddm

import (
	"fmt"
	"time"

	"github.com/alfonzso/dying-disk-manager/pkg/linux"
	"github.com/alfonzso/dying-disk-manager/pkg/observer"
	log "github.com/sirupsen/logrus"
)

func (ddmData *DDMData) setupRepairThread(diskData *DiskData) {
	if !diskData.Repair.IsRunning() {
		return
	}
	diskData.Repair.HealthCheck = observer.OK

	WaitForThreadToBeIddle(fmt.Sprintf("%s - repairSetup", diskData.Name), diskData.Repair.ActionsToStop)

	if ddmData.PreRepair(diskData).IsSucceed() {
		res := ddmData.Repair(diskData)
		diskData.Active = true
		if res.IsFailed() {
			diskData.Active = false
			log.Debugf("[%s] Current disk set Active to false", diskData.Name)
		}
	}
	StartThreads(diskData.Repair.ActionsToStop)

	diskData.Repair.SetToIddle()
	go func() {
		time.Sleep(10 * time.Second)
		diskData.Repair.HealthCheck = observer.None
	}()
}

func (ddmData *DDMData) PreRepair(diskData *DiskData) linux.LinuxCommands {
	log.Debugf("[%s] PreRepair ...", diskData.Name)

	if ddmData.Exec.UMount(*diskData.conf).IsFailed() {
		log.Debugf("[%s] PreRepair failed to umount disk ... ", diskData.Name)
		return linux.CommandError
	}
	return linux.CommandSuccess
}

func (ddmData *DDMData) Repair(diskData *DiskData) linux.LinuxCommands {
	if ddmData.Exec.RunFsck(diskData.UUID).IsFailed() {
		log.Debugf("[%s] Repair with fsck failed :( ", diskData.Name)
		return linux.CommandError
	}

	if ddmData.Exec.Mount(*diskData.conf).IsFailed() {
		log.Debugf("[%s] Mount after repair failed :( ", diskData.Name)
		return linux.CommandError
	}

	log.Debugf("[%s] Repair Succeeded ! ", diskData.Name)
	return linux.CommandSuccess
}
