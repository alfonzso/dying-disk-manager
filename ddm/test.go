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
		if ddmData.isTestCanBeRun(disk, diskStat) {
			go ddmData.SetupCron(
				"TEST",
				Test,
				disk,
				diskStat,
				GetCronExpr(disk.Test.Cron, ddmData.Common.Test.Cron),
			)
			diskStat.TestThreadIsRunning = true
		}
	}
	// time.Sleep(30 * time.Second)
}

func Test(disk config.Disk, diskStat *observer.DiskStat) (int, error) {
	if diskStat.Active {
		log.Debugf("[%s] Testing disk ", disk.Name)
	}
	return 0, nil
}
