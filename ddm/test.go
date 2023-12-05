package ddm

import (
	"time"

	"github.com/alfonzso/dying-disk-manager/pkg/config"
	"github.com/alfonzso/dying-disk-manager/pkg/observer"

	log "github.com/sirupsen/logrus"
)

func (ddmData *DDMData) setupTestThread(disk config.Disk) {
	diskStat := ddmData.GetDiskStat(disk)

	if RepairIsOn(diskStat.Test.Name, diskStat) {
		return
	}

	if !(disk.Test.Enabled || ddmData.Common.Test.Enabled) {
		return
	}

	go ddmData.startTestAndWaitTillAlive(disk)
}

func (ddmData *DDMData) startTestAndWaitTillAlive(disk config.Disk) {
	diskStat := ddmData.GetDiskStat(disk)

	if ddmData.ActionsJobRunning(diskStat.Test.Name, disk.UUID) {
		return
	}

	ddmData.SetupCron(
		diskStat.Test.Name,
		ddmData.diskTest,
		disk,
		GetCronExpr(disk.Test.Cron, ddmData.Common.Test.Cron),
	)

	diskStat.Test.SetToRun()
}

func (ddmData *DDMData) diskTest(disk config.Disk) (int, error) {
	res, err := ddmData.testWrapper(disk)
	go func() {
		diskStat := ddmData.GetDiskStat(disk)
		times := ddmData.GetJobNextRun(diskStat.Test.Name, diskStat.UUID)
		time.Sleep(times)
		diskStat.Test.HealthCheck = observer.None
	}()
	return res, err
}

func (ddmData *DDMData) testWrapper(disk config.Disk) (int, error) {
	diskStat := ddmData.GetDiskStat(disk)
	diskStat.Test.SetToRun()
	diskStat.Test.HealthCheck = observer.OK

	if IsInActiveOrDisabled(diskStat.Test.Name, diskStat, diskStat.Test) {
		diskStat.Test.SetToIddle()
		return 0, nil
	}

	if ddmData.Exec.RunDryFsck(disk.UUID).IsFailed() {
		log.Debugf("[%s] Fsck failed, triggering repair", disk.Name)
		diskStat.Active = false
		diskStat.Repair.SetToRun()
	}

	if ddmData.Exec.WriteIntoDisk(disk.Mount.Path).IsFailed() {
		log.Debugf("[%s] Write to disk failed, triggering repair", disk.Name)
		diskStat.Active = false
		diskStat.Repair.SetToRun()

	}

	diskStat.Test.SetToIddle()
	return 0, nil
}
