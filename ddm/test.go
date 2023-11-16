package ddm

import (
	"github.com/alfonzso/dying-disk-manager/pkg/config"
	"github.com/alfonzso/dying-disk-manager/pkg/observer"

	log "github.com/sirupsen/logrus"
)

func (ddmData *DDMData) isTestCanBeRun(disk config.Disk, diskStat *observer.DiskStat) bool {
	return (disk.Test.Enabled || ddmData.Common.Test.Enabled) && !diskStat.TestThreadIsRunning
}

func (ddmData *DDMData) setupTestThread() {
	for _, disk := range ddmData.Disks {
		diskStat := ddmData.GetDiskStat(disk)
		if diskStat.RepairThreadIsRunning {
			if diskStat.ActionStatus.Test.IsRunning() {
				diskStat.ActionStatus.Test = observer.Iddle
			}
			log.Debugf("[%s] TEST -> Repair is ON", disk.Name)
		} else if ddmData.isTestCanBeRun(disk, diskStat) {
			go ddmData.SetupCron(
				"TEST",
				ddmData.Test,
				disk,
				diskStat,
				GetCronExpr(disk.Test.Cron, ddmData.Common.Test.Cron),
			)
			diskStat.TestThreadIsRunning = true
		}
	}
}

func (ddmData *DDMData) Test(disk config.Disk, diskStat *observer.DiskStat) (int, error) {
	diskStat.ActionStatus.Test = observer.Running
	if !diskStat.Active {
		log.Warningf("[%s] Disk deactivated in Test thread", disk.Name)
		return 0, nil
	}

	if ddmData.Exec.WriteIntoDisk(disk.Mount.Path).IsFailed() {
		log.Debugf("[%s] Write to disk failed, triggering repair", disk.Name)
		diskStat.Active = false
		diskStat.RepairThreadIsRunning = true
		return 0, nil
	}
	diskStat.ActionStatus.Test = observer.Iddle
	return 0, nil
}
