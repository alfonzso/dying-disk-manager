package ddm

import (
	"github.com/alfonzso/dying-disk-manager/pkg/config"
	"github.com/alfonzso/dying-disk-manager/pkg/observer"

	log "github.com/sirupsen/logrus"
)

func (ddmData *DDMData) isTestCanBeRun(disk config.Disk, diskStat *observer.DiskStat) bool {
	// return (disk.Test.Enabled || ddmData.Common.Test.Enabled) && diskStat.Test.IsStopped()
	// dbg := ddmData.CurrentActionsJobNotRunning("TEST", disk.Name, disk.UUID)
	// log.Debug("TEST ... JobNotRunning", dbg)
	// log.Debug("TEST ... JobNotRunning", dbg, disk.Name, disk.UUID)
	// return (disk.Test.Enabled || ddmData.Common.Test.Enabled) && dbg
	return (disk.Test.Enabled || ddmData.Common.Test.Enabled)
}

func (ddmData *DDMData) setupTestThread(disk config.Disk) {
	diskStat := ddmData.GetDiskStat(disk)

	RepairIsOn("TEST", diskStat)

	if !ddmData.isTestCanBeRun(disk, diskStat) {
		return
	}

	// log.Debugf("%s %s %s", "TEST", disk.Name, disk.UUID)
	// log.Debugf("%#v", ddmData.Scheduler.Jobs())
	// log.Debugf("jobNotRunning %t", ddmData.CurrentActionsJobNotRunning("TEST", disk.Name, disk.UUID))
	// log.Debugf(
	// 	"%s %s %s jobNotRunning %t", "TEST", disk.Name, disk.UUID, ddmData.CurrentActionsJobNotRunning("TEST", disk.Name, disk.UUID),
	// )

	// if !ddmData.CurrentActionsJobNotRunning("TEST", disk.Name, disk.UUID) {
	// 	return
	// }

	// ddmData.SetupCron(
	// 	"TEST",
	// 	ddmData.Test,
	// 	disk,
	// 	nil,
	// 	GetCronExpr(disk.Test.Cron, ddmData.Common.Test.Cron),
	// )

	// diskStat.Test.SetToRun()
	go ddmData.startTestAndWaitTillAlive(disk)
}

func (ddmData *DDMData) startTestAndWaitTillAlive(disk config.Disk) {
	if ddmData.CurrentActionsJobRunning("TEST", disk.Name, disk.UUID) {
		return
	}

	diskStat := ddmData.GetDiskStat(disk)

	ddmData.SetupCron(
		"TEST",
		ddmData.Test,
		disk,
		nil,
		GetCronExpr(disk.Test.Cron, ddmData.Common.Test.Cron),
	)
	// time.Sleep(2 * time.Second)
	// for {
	// 	if ddmData.CurrentActionsJobRunning("TEST", disk.Name, disk.UUID) {
	// 		log.Debugf("[%s] TEST action alive", disk.Name)
	// 		break
	// 	}
	// 	time.Sleep(1 * time.Second)
	// }

	diskStat.Test.SetToRun()
}

func (ddmData *DDMData) Test(disk config.Disk, action []*observer.Action) (int, error) {
	currentDiskStat := ddmData.GetDiskStat(disk)
	currentDiskStat.Test.SetToRun()

	if IsInActiveOrDisabled("Test", currentDiskStat, currentDiskStat.Test) {
		currentDiskStat.Test.SetToIddle()
		return 0, nil
	}

	if ddmData.Exec.RunDryFsck(disk.UUID).IsFailed() {
		log.Debugf("[%s] Fsck failed, triggering repair", disk.Name)
		currentDiskStat.Active = false
		currentDiskStat.Repair.SetToRun()
	}

	if ddmData.Exec.WriteIntoDisk(disk.Mount.Path).IsFailed() {
		log.Debugf("[%s] Write to disk failed, triggering repair", disk.Name)
		currentDiskStat.Active = false
		currentDiskStat.Repair.SetToRun()

	}

	currentDiskStat.Test.SetToIddle()
	return 0, nil
}
