package ddm

import (
	"github.com/alfonzso/dying-disk-manager/pkg/config"
	"github.com/alfonzso/dying-disk-manager/pkg/observer"

	log "github.com/sirupsen/logrus"
)

func (ddmData *DDMData) isTestCanBeRun(disk config.Disk, diskStat *observer.DiskStat) bool {
	return (disk.Test.Enabled || ddmData.Common.Test.Enabled) && !diskStat.Test.ThreadIsRunning
}

func (ddmData *DDMData) setupTestThread(disk config.Disk) {
	diskStat := ddmData.GetDiskStat(disk)

	if diskStat.Repair.IsRunning() {
		if diskStat.Test.IsRunning() {
			diskStat.Test.SetToIddle()
		}
		log.Debugf("[%s] TEST -> Repair is ON", disk.Name)
	} else if ddmData.isTestCanBeRun(disk, diskStat) {
		go ddmData.SetupCron(
			"TEST",
			ddmData.Test,
			disk,
			nil,
			GetCronExpr(disk.Test.Cron, ddmData.Common.Test.Cron),
		)
		diskStat.Test.ThreadIsRunning = true
	}
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
		// currentDiskStat.Repair.ThreadIsRunning = true
		currentDiskStat.Repair.SetToRun()
	}

	if ddmData.Exec.WriteIntoDisk(disk.Mount.Path).IsFailed() {
		log.Debugf("[%s] Write to disk failed, triggering repair", disk.Name)
		currentDiskStat.Active = false
		// currentDiskStat.Repair.ThreadIsRunning = true
		currentDiskStat.Repair.SetToRun()

	}

	currentDiskStat.Test.SetToIddle()
	return 0, nil
}
