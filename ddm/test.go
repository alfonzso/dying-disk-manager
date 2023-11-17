package ddm

import (
	"github.com/alfonzso/dying-disk-manager/pkg/config"
	"github.com/alfonzso/dying-disk-manager/pkg/observer"

	log "github.com/sirupsen/logrus"
)

func (ddmData *DDMData) isTestCanBeRun(disk config.Disk, diskStat *observer.DiskStat) bool {
	return (disk.Test.Enabled || ddmData.Common.Test.Enabled) && !diskStat.Test.ThreadIsRunning
}

func (ddmData *DDMData) setupTestThread() {
	for _, disk := range ddmData.Disks {
		diskStat := ddmData.GetDiskStat(disk)
		if diskStat.Repair.ThreadIsRunning {
			if diskStat.Test.Status.IsRunning() {
				diskStat.Test.Status = observer.Iddle
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
			diskStat.Test.ThreadIsRunning = true
		}
	}
}

func (ddmData *DDMData) Test(disk config.Disk, diskStat *observer.DiskStat) (int, error) {
	diskStat.Test.Status = observer.Running

	if diskStat.IsNotActiv() || diskStat.Test.DisabledByAction {
		log.Warningf("[%s] TestThread => Disk deactivated => active: %t, disabledBy: %t",
			diskStat.Name, diskStat.Active, diskStat.Test.DisabledByAction,
		)
		diskStat.Test.Status = observer.Iddle
		return 0, nil
	}

	if ddmData.Exec.WriteIntoDisk(disk.Mount.Path).IsFailed() {
		log.Debugf("[%s] Write to disk failed, triggering repair", disk.Name)
		diskStat.Active = false
		diskStat.Repair.ThreadIsRunning = true
		return 0, nil
	}

	diskStat.Test.Status = observer.Iddle
	return 0, nil
}
